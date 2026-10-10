package sdb

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// scanRows is the subset of pgx.Rows used by Scan. pgx.Rows implements it.
// Keeping the interface small also makes the scanner easy to test without a database.
type scanRows interface {
	FieldDescriptions() []pgconn.FieldDescription
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

// Scan reads rows into dest. dest must be a non-nil pointer to one of:
//   - a scalar (one column, first row only),
//   - a struct (first row only),
//   - a map[string]any (first row only),
//   - a slice of scalars, structs, pointers to structs, or map[string]any (all rows).
//
// Struct fields are matched by the `json:"name"` tag or, when absent, by the
// snake_case version of the Go field name. The options after a comma are ignored;
// `json:"-"` excludes a field. GORM tags are not read. Unknown result columns are ignored.
// SQL NULL is converted to the zero value of the destination field/type.
// Scan closes rows before returning. It returns nil for an empty result set when
// dest is a struct, map, or scalar; use QueryRow().Scan for ErrNoRows semantics.
func Scan(rows scanRows, dest any) error {
	if rows == nil {
		return errors.New("sdb: Scan: rows is nil")
	}
	defer rows.Close()

	v := reflect.ValueOf(dest)
	if !v.IsValid() || v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("sdb: Scan: dest must be a non-nil pointer")
	}
	v = v.Elem()

	isSlice := v.Kind() == reflect.Slice
	if isSlice {
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}

	for rows.Next() {
		if isSlice {
			itemType := v.Type().Elem()
			ptrItem := itemType.Kind() == reflect.Ptr
			baseType := itemType
			if ptrItem {
				baseType = itemType.Elem()
			}
			item := reflect.New(baseType).Elem()
			if err := scanCurrentRow(rows, item); err != nil {
				return err
			}
			if ptrItem {
				p := reflect.New(baseType)
				p.Elem().Set(item)
				v.Set(reflect.Append(v, p))
			} else {
				v.Set(reflect.Append(v, item))
			}
			continue
		}

		if err := scanCurrentRow(rows, v); err != nil {
			return err
		}
		// For non-slice destinations, only the first row is consumed.
		for rows.Next() {
			// Drain remaining rows so pgx can release the connection cleanly.
		}
		break
	}
	return rows.Err()
}

// scanFirst scans the first row and returns ErrRecordNotFound when the query
// returns no rows. It is used by SelectOne, whose public contract distinguishes
// an empty result from an empty destination value.
func scanFirst(rows scanRows, dest any) error {
	if rows == nil {
		return errors.New("sdb: Scan: rows is nil")
	}
	defer rows.Close()

	v := reflect.ValueOf(dest)
	if !v.IsValid() || v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("sdb: Scan: dest must be a non-nil pointer")
	}
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return ErrRecordNotFound
	}
	if err := scanCurrentRow(rows, v.Elem()); err != nil {
		return err
	}
	for rows.Next() {
		// SelectOne limits the query to one row, but drain defensively if callers
		// construct a query that returns more than one.
	}
	return rows.Err()
}

func scanCurrentRow(rows scanRows, target reflect.Value) error {
	if !target.IsValid() {
		return errors.New("sdb: Scan: invalid destination")
	}
	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		return scanCurrentRow(rows, target.Elem())
	}

	columns := rows.FieldDescriptions()
	switch target.Kind() {
	case reflect.Struct:
		fields := structColumnMap(target.Type())
		dests := make([]any, len(columns))
		setters := make([]reflect.Value, len(columns))
		for i, col := range columns {
			if idx, ok := fields[normalizeColumn(col.Name)]; ok {
				field := target.Field(idx)
				if !field.CanSet() {
					dests[i] = new(any) // Ignore unexported fields.
					continue
				}
				// Scan into **T so pgx can distinguish SQL NULL from a concrete
				// value. Always scan into the non-pointer base type: wrapping a
				// pointer field directly would produce ***T (or deeper), which
				// pgx codecs do not support for values such as time.Time.
				baseType := dereferenceType(field.Type())
				scanType := scanTypeForColumn(col.DataTypeOID, baseType)
				d := reflect.New(reflect.PointerTo(scanType))
				dests[i] = d.Interface()
				setters[i] = d
			} else {
				dests[i] = new(any) // Ignore columns absent from the destination struct.
			}
		}
		if err := rows.Scan(dests...); err != nil {
			return fmt.Errorf("sdb: scan struct row: %w", err)
		}
		for i, d := range setters {
			if !d.IsValid() {
				continue
			}
			holder := d.Elem() // *T
			fieldIndex := fields[normalizeColumn(columns[i].Name)]
			field := target.Field(fieldIndex)
			if holder.IsNil() {
				field.Set(reflect.Zero(field.Type()))
			} else {
				value, err := valueForColumnType(field.Type(), holder.Elem(), columns[i].DataTypeOID)
				if err != nil {
					return fmt.Errorf("sdb: assign column %q to field %s: %w", columns[i].Name, field.Type(), err)
				}
				field.Set(value)
			}
		}
		return nil

	case reflect.Map:
		if target.Type().Key().Kind() != reflect.String || target.Type().Elem().Kind() != reflect.Interface {
			return fmt.Errorf("sdb: Scan: map destination must be map[string]any, got %s", target.Type())
		}
		if target.IsNil() {
			target.Set(reflect.MakeMap(target.Type()))
		}
		dests := make([]any, len(columns))
		values := make([]reflect.Value, len(columns))
		for i := range columns {
			p := new(any)
			dests[i] = p
			values[i] = reflect.ValueOf(p)
		}
		if err := rows.Scan(dests...); err != nil {
			return fmt.Errorf("sdb: scan map row: %w", err)
		}
		for i, p := range values {
			value := p.Elem().Interface()
			mapValue := reflect.ValueOf(value)
			if !mapValue.IsValid() {
				mapValue = reflect.Zero(target.Type().Elem())
			}
			target.SetMapIndex(reflect.ValueOf(columns[i].Name), mapValue)
		}
		return nil

	default:
		if len(columns) != 1 {
			return fmt.Errorf("sdb: Scan: scalar destination %s requires exactly one column, got %d", target.Type(), len(columns))
		}
		if !target.CanSet() {
			return fmt.Errorf("sdb: Scan: destination %s is not settable", target.Type())
		}
		baseType := dereferenceType(target.Type())
		scanType := scanTypeForColumn(columns[0].DataTypeOID, baseType)
		d := reflect.New(reflect.PointerTo(scanType))
		if err := rows.Scan(d.Interface()); err != nil {
			return fmt.Errorf("sdb: scan scalar row: %w", err)
		}
		holder := d.Elem()
		if holder.IsNil() {
			target.Set(reflect.Zero(target.Type()))
		} else {
			value, err := valueForColumnType(target.Type(), holder.Elem(), columns[0].DataTypeOID)
			if err != nil {
				return fmt.Errorf("sdb: assign scalar to %s: %w", target.Type(), err)
			}
			target.Set(value)
		}
		return nil
	}
}

// scanTypeForColumn chooses a pgx-compatible intermediate type for destination fields.
// pgx decodes PostgreSQL timestamps as time.Time, not string, when rows use binary format.
func scanTypeForColumn(oid uint32, targetType reflect.Type) reflect.Type {
	if targetType.Kind() == reflect.String && isDateTimeOID(oid) {
		return reflect.TypeOf(time.Time{})
	}
	return targetType
}

func isDateTimeOID(oid uint32) bool {
	switch oid {
	case pgtype.DateOID, pgtype.TimestampOID, pgtype.TimestamptzOID:
		return true
	default:
		return false
	}
}

func formatDateTimeString(t time.Time, oid uint32) string {
	switch oid {
	case pgtype.DateOID:
		return t.Format("2006-01-02")
	case pgtype.TimestamptzOID:
		return t.Format("2006-01-02 15:04:05.999999999-07:00")
	default: // timestamp without time zone
		return t.Format("2006-01-02 15:04:05.999999999")
	}
}

func valueForColumnType(targetType reflect.Type, value reflect.Value, oid uint32) (reflect.Value, error) {
	if dereferenceType(targetType).Kind() == reflect.String && value.IsValid() && value.Type() == reflect.TypeOf(time.Time{}) && isDateTimeOID(oid) {
		formatted := reflect.ValueOf(formatDateTimeString(value.Interface().(time.Time), oid))
		return valueForType(targetType, formatted)
	}
	return valueForType(targetType, value)
}

// dereferenceType returns the non-pointer base type of t.
func dereferenceType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// valueForType reconstructs the requested pointer depth around a decoded base
// value. A nil SQL value is handled by the caller and becomes the target's zero value.
func valueForType(targetType reflect.Type, value reflect.Value) (reflect.Value, error) {
	if targetType.Kind() == reflect.Ptr {
		p := reflect.New(targetType.Elem())
		inner, err := valueForType(targetType.Elem(), value)
		if err != nil {
			return reflect.Value{}, err
		}
		p.Elem().Set(inner)
		return p, nil
	}
	if value.Type().AssignableTo(targetType) {
		return value, nil
	}
	if value.Type().ConvertibleTo(targetType) {
		return value.Convert(targetType), nil
	}
	return reflect.Value{}, fmt.Errorf("cannot assign %s to %s", value.Type(), targetType)
}

func structColumnMap(t reflect.Type) map[string]int {
	out := make(map[string]int, t.NumField())
	// Explicit JSON names take precedence over inferred names if two fields
	// happen to resolve to the same database column.
	explicit := make(map[string]bool, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" { // unexported
			continue
		}
		tag, hasTag := f.Tag.Lookup("json")
		name := ""
		isExplicit := false
		if hasTag {
			name = strings.SplitN(tag, ",", 2)[0]
			if name == "-" {
				continue
			}
			if name != "" {
				isExplicit = true
			} else {
				name = snakeCase(f.Name)
			}
		} else {
			name = snakeCase(f.Name)
		}
		key := normalizeColumn(name)
		if previousExplicit, exists := explicit[key]; exists && previousExplicit && !isExplicit {
			continue
		}
		if _, exists := out[key]; !exists || isExplicit || !explicit[key] {
			out[key] = i
			explicit[key] = isExplicit
		}
	}
	return out
}

func normalizeColumn(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func snakeCase(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
