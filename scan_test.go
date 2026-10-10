package sdb

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type fakeRows struct {
	columns []string
	oids    []uint32
	data    [][]any
	index   int
	closed  bool
	err     error
}

func newFakeRows(columns []string, data ...[]any) *fakeRows {
	return &fakeRows{columns: columns, data: data, index: -1}
}

func newFakeRowsWithOIDs(columns []string, oids []uint32, data ...[]any) *fakeRows {
	return &fakeRows{columns: columns, oids: oids, data: data, index: -1}
}
func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription {
	out := make([]pgconn.FieldDescription, len(r.columns))
	for i, name := range r.columns {
		out[i].Name = name
		if i < len(r.oids) {
			out[i].DataTypeOID = r.oids[i]
		}
	}
	return out
}
func (r *fakeRows) Next() bool {
	if r.index+1 >= len(r.data) {
		return false
	}
	r.index++
	return true
}
func (r *fakeRows) Scan(dest ...any) error {
	if r.index < 0 || r.index >= len(r.data) {
		return errors.New("no current row")
	}
	if len(dest) != len(r.columns) {
		return errors.New("column count mismatch")
	}
	for i, d := range dest {
		if err := fakeAssign(d, r.data[r.index][i]); err != nil {
			return err
		}
	}
	return nil
}
func (r *fakeRows) Err() error { return r.err }
func (r *fakeRows) Close()     { r.closed = true }

func fakeAssign(dest any, value any) error {
	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr || d.IsNil() {
		return errors.New("destination is not a pointer")
	}
	v := d.Elem()
	if v.Kind() == reflect.Interface {
		if value == nil {
			v.SetZero()
		} else {
			v.Set(reflect.ValueOf(value))
		}
		return nil
	}
	if v.Kind() == reflect.Ptr {
		if value == nil {
			v.SetZero()
			return nil
		}
		p := reflect.New(v.Type().Elem())
		if err := fakeSet(p.Elem(), value); err != nil {
			return err
		}
		v.Set(p)
		return nil
	}
	return fakeSet(v, value)
}
func fakeSet(dst reflect.Value, value any) error {
	src := reflect.ValueOf(value)
	if !src.IsValid() {
		dst.SetZero()
		return nil
	}
	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
		return nil
	}
	if src.Type().ConvertibleTo(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
		return nil
	}
	return errors.New("incompatible fake scan types: " + src.Type().String() + " -> " + dst.Type().String())
}

func TestScanStructTagsAndNullToZero(t *testing.T) {
	type user struct {
		ID        int64  `json:"user_id"`
		FirstName string `json:"first_name"`
		Age       int    `json:"age,omitempty"`
		Ignored   string `json:"-"`
	}
	rows := newFakeRows(
		[]string{"user_id", "first_name", "age", "ignored"},
		[]any{int64(42), "Ada", nil, "not mapped"},
	)
	var got user
	if err := Scan(rows, &got); err != nil {
		t.Fatal(err)
	}
	want := user{ID: 42, FirstName: "Ada", Age: 0}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	if !rows.closed {
		t.Fatal("rows were not closed")
	}
}

func TestScanUsesJSONTagsAndIgnoresGORMTags(t *testing.T) {
	type record struct {
		ID       int    `gorm:"column:legacy_id"`
		JSONName string `json:"display_name,omitempty"`
		Default  string
		Ignored  string `json:"-"`
	}
	rows := newFakeRows(
		[]string{"id", "display_name", "default", "ignored"},
		[]any{11, "Ada", "fallback", "not mapped"},
	)
	var got record
	if err := Scan(rows, &got); err != nil {
		t.Fatal(err)
	}
	want := record{ID: 11, JSONName: "Ada", Default: "fallback"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestScanSliceOfStructPointers(t *testing.T) {
	type item struct {
		ItemID int    `json:"item_id"`
		Name   string `json:"name"`
	}
	rows := newFakeRows([]string{"item_id", "name"}, []any{1, "one"}, []any{2, "two"})
	var got []*item
	if err := Scan(rows, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ItemID != 1 || got[1].Name != "two" {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestScanScalarAndScalarSlice(t *testing.T) {
	rows := newFakeRows([]string{"count"}, []any{int64(7)})
	var got int
	if err := Scan(rows, &got); err != nil {
		t.Fatal(err)
	}
	if got != 7 {
		t.Fatalf("got %d, want 7", got)
	}

	rows = newFakeRows([]string{"count"}, []any{int64(3)}, []any{int64(4)})
	var values []int
	if err := Scan(rows, &values); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(values, []int{3, 4}) {
		t.Fatalf("got %#v", values)
	}
}

func TestScanMapAndNull(t *testing.T) {
	rows := newFakeRows([]string{"name", "optional"}, []any{"Ada", nil})
	var got map[string]any
	if err := Scan(rows, &got); err != nil {
		t.Fatal(err)
	}
	if got["name"] != "Ada" {
		t.Fatalf("unexpected name: %#v", got["name"])
	}
	if value, ok := got["optional"]; !ok || value != nil {
		t.Fatalf("NULL should map to nil, got %#v", got)
	}
}

func TestScanValidation(t *testing.T) {
	rows := newFakeRows([]string{"a"}, []any{1})
	var got int
	if err := Scan(rows, got); err == nil {
		t.Fatal("expected pointer validation error")
	}

	rows = newFakeRows([]string{"a", "b"}, []any{1, 2})
	if err := Scan(rows, &got); err == nil {
		t.Fatal("expected scalar column-count error")
	}
}

func TestScanFirstReturnsNotFoundAndScansOnlyFirstRow(t *testing.T) {
	rows := newFakeRows([]string{"value"})
	var got int
	if err := scanFirst(rows, &got); !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}

	rows = newFakeRows([]string{"value"}, []any{int(5)}, []any{int(6)})
	if err := scanFirst(rows, &got); err != nil {
		t.Fatal(err)
	}
	if got != 5 {
		t.Fatalf("got %d, want first row value 5", got)
	}
	if !rows.closed {
		t.Fatal("rows were not closed")
	}
}

func TestScanPointerToStruct(t *testing.T) {
	type item struct {
		ID int `json:"id"`
	}
	rows := newFakeRows([]string{"id"}, []any{7})
	var got *item
	if err := Scan(rows, &got); err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ID != 7 {
		t.Fatalf("unexpected pointer result: %#v", got)
	}
}

func TestScanTimeFieldsWithDifferentPointerDepths(t *testing.T) {
	type record struct {
		Created time.Time   `json:"created"`
		Updated *time.Time  `json:"updated"`
		Deleted **time.Time `json:"deleted"`
	}
	wantTime := time.Date(2026, time.October, 10, 12, 30, 0, 0, time.UTC)
	rows := newFakeRows(
		[]string{"created", "updated", "deleted"},
		[]any{wantTime, wantTime, wantTime},
	)
	var got record
	if err := Scan(rows, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Created.Equal(wantTime) {
		t.Fatalf("Created = %v, want %v", got.Created, wantTime)
	}
	if got.Updated == nil || !got.Updated.Equal(wantTime) {
		t.Fatalf("Updated = %v, want %v", got.Updated, wantTime)
	}
	if got.Deleted == nil || *got.Deleted == nil || !(**got.Deleted).Equal(wantTime) {
		t.Fatalf("Deleted = %v, want pointer depth 2 to %v", got.Deleted, wantTime)
	}
}

func TestScanNullTimeFieldsBecomeZeroValues(t *testing.T) {
	type record struct {
		Created time.Time  `json:"created"`
		Updated *time.Time `json:"updated"`
	}
	rows := newFakeRows([]string{"created", "updated"}, []any{nil, nil})
	var got record
	if err := Scan(rows, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Created.IsZero() {
		t.Fatalf("Created = %v, want zero time", got.Created)
	}
	if got.Updated != nil {
		t.Fatalf("Updated = %v, want nil", got.Updated)
	}
}

func TestScanTimestampIntoStringField(t *testing.T) {
	type record struct {
		Created string `json:"created"`
	}
	stamp := time.Date(2026, time.October, 10, 15, 47, 0, 0, time.UTC)
	rows := newFakeRowsWithOIDs(
		[]string{"created"},
		[]uint32{1114}, // PostgreSQL timestamp without time zone
		[]any{stamp},
	)
	var got record
	if err := Scan(rows, &got); err != nil {
		t.Fatal(err)
	}
	if got.Created != "2026-10-10 15:47:00" {
		t.Fatalf("Created = %q, want %q", got.Created, "2026-10-10 15:47:00")
	}
}

func TestScanTimestamptzIntoStringField(t *testing.T) {
	type record struct {
		Created string `json:"created"`
	}
	stamp := time.Date(2026, time.October, 10, 15, 47, 0, 0, time.FixedZone("UTC+5", 5*60*60))
	rows := newFakeRowsWithOIDs(
		[]string{"created"},
		[]uint32{1184}, // PostgreSQL timestamp with time zone
		[]any{stamp},
	)
	var got record
	if err := Scan(rows, &got); err != nil {
		t.Fatal(err)
	}
	if got.Created != "2026-10-10 15:47:00+05:00" {
		t.Fatalf("Created = %q, want %q", got.Created, "2026-10-10 15:47:00+05:00")
	}
}
