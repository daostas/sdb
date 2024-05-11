package sdb

import (
	"errors"
	"fmt"
	qp "github.com/daostas/query_parser"

	"strconv"

	"net/http"
	"reflect"
	"strings"
)

type DigitConstraints interface {
	int | int8 | int16 | int32 | int64
}

type ModelTable interface {
	Table() string
}

type SqlMap map[string]string

func (m SqlMap) AddToMap(key string, value interface{}) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		if value != nil && !v.IsNil() {
			m[key], _ = ValueToPostgresValue(reflect.Indirect(v).Interface(), false)
		}
	default:
		str := fmt.Sprintf("%v", value)
		if value != nil && str != "<nil>" {
			m[key], _ = ValueToPostgresValue(reflect.Indirect(v).Interface(), false)
		}
	}

}

//func (m SqlMap) AddToMap(key string, value interface{}) {
//	str := fmt.Sprintf("%v", value)
//	if value != nil && str != "<nil>" {
//		v := reflect.ValueOf(value)
//		m[key] = ValueToPostgresValue(reflect.Indirect(v).Interface())
//	}
//}

func delete[Model ModelTable](sdb Sdb, model Model, where interface{}) (err error) {
	_where, err := parseWhere(sdb, model, where)
	if err != nil {
		return
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", model.Table(), _where)
	if sdb.log {
		sdb.logger.Info(query)
	}
	return sdb.db.Exec(query).Error
}

func Delete[Model ModelTable](sdb Sdb, model Model, where interface{}) (err error) {
	return delete(sdb, model, where)
}

func update[Model ModelTable](sdb Sdb, model Model, set SqlMap, where interface{}) (err error) {
	_where, err := parseWhere(sdb, model, where)
	if err != nil {
		return
	}

	if len(set) == 0 {
		return fmt.Errorf("cannot set empty values")
	}
	_set := ""
	for key, value := range set {
		_set += fmt.Sprintf("%s = %s, ", key, value)
	}
	_set = _set[:len(_set)-2]

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", model.Table(), _set, _where)
	if sdb.log {
		sdb.logger.Info(query)
	}
	return sdb.db.Exec(query).Error
}

func Update[Model ModelTable](sdb Sdb, model Model, set SqlMap, where interface{}) (err error) {
	return update(sdb, model, set, where)
}

func updateSelectOne[Model ModelTable](sdb Sdb, model Model, set SqlMap, where interface{}) (out Model, err error) {
	_where, err := parseWhere(sdb, model, where)
	if err != nil {
		return
	}

	if len(set) == 0 {
		return out, fmt.Errorf("cannot set empty values")
	}
	_set := ""
	for key, value := range set {
		_set += fmt.Sprintf("%s = %s, ", key, value)
	}
	_set = _set[:len(_set)-2]

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s returning *", model.Table(), _set, _where)
	if sdb.log {
		sdb.logger.Info(query)
	}
	err = sdb.db.Raw(query).Scan(&out).Error
	return
}

func UpdateSelectOne[Model ModelTable](sdb Sdb, model Model, set SqlMap, where interface{}) (out Model, err error) {
	return updateSelectOne(sdb, model, set, where)
}

func updateSelectAll[Model ModelTable](sdb Sdb, model Model, set SqlMap, where interface{}) (out []*Model, err error) {
	_where, err := parseWhere(sdb, model, where)
	if err != nil {
		return
	}

	if len(set) == 0 {
		return out, fmt.Errorf("cannot set empty values")
	}
	_set := ""
	for key, value := range set {
		_set += fmt.Sprintf("%s = %s, ", key, value)
	}
	_set = _set[:len(_set)-2]

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s returning *", model.Table(), _set, _where)
	if sdb.log {
		sdb.logger.Info(query)
	}
	err = sdb.db.Raw(query).Scan(&out).Error
	return
}

func UpdateSelectAll[Model ModelTable](sdb Sdb, model Model, set SqlMap, where interface{}) (out []*Model, err error) {
	return updateSelectAll(sdb, model, set, where)
}

func create[Model ModelTable](sdb Sdb, model Model, template string, args ...interface{}) (err error) {
	if template != "" {
		count := len(strings.Split(template, ","))
		if count != len(args) {
			return fmt.Errorf("number of args not match to amount of arguments in template")
		}
		template = fmt.Sprintf("(%s)", template)
	}

	query := fmt.Sprintf("INSERT INTO %s %s VAlUES(", model.Table(), template)
	for _, v := range args {
		temp, err := ValueToPostgresValue(v, false)
		if err != nil {
			return err
		}
		query += temp + ", "
	}
	query = query[:len(query)-2] + ")"
	if sdb.log {
		sdb.logger.Info(query)
	}
	return sdb.db.Exec(query).Error
}

func Create[Model ModelTable](sdb Sdb, model Model, template string, args ...interface{}) (err error) {
	return create(sdb, model, template, args...)
}

func createSelect[Model ModelTable](sdb Sdb, model Model, template string, args ...interface{}) (out Model, err error) {
	if template != "" {
		count := len(strings.Split(template, ","))
		if count != len(args) {
			return out, fmt.Errorf("number of args not match to amount of arguments in template")
		}
		template = fmt.Sprintf("(%s)", template)
	}

	query := fmt.Sprintf("INSERT INTO %s %s VAlUES(", model.Table(), template)
	for _, v := range args {
		temp, err := ValueToPostgresValue(v, false)
		if err != nil {
			return out, err
		}
		query += temp + ", "
	}
	query = query[:len(query)-2] + ") returning *"

	if sdb.log {
		sdb.logger.Info(query)
	}

	err = sdb.db.Raw(query).Scan(&out).Error
	return
}

func CreateSelect[Model ModelTable](sdb Sdb, model Model, template string, args ...interface{}) (out Model, err error) {
	return createSelect(sdb, model, template, args...)
}

func count[Model ModelTable](sdb Sdb, model Model, where interface{}) (total int, err error) {
	_where, err := parseWhere(sdb, model, where)
	if err != nil {
		return
	}

	//Достаем кол-во записей из базы
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s ", strings.ToLower(model.Table()))
	if _where != "" {
		query += fmt.Sprintf("WHERE %s", _where)
	}

	if sdb.log {
		sdb.logger.Info(query)
	}
	err = sdb.db.Raw(query).Scan(&total).Error
	return
}

func Count[Model ModelTable](sdb Sdb, model Model, where interface{}) (total int, err error) {
	return count(sdb, model, where)
}

func callRoutine(sdb Sdb, funcName string, args ...any) (statusCode int, err error) {
	query := fmt.Sprintf("SELECT * FROM %s(", funcName)
	//Формируем аргументы из args
	for i, arg := range args {
		temp, err := ValueToPostgresValue(arg, false)
		if err != nil {
			return http.StatusUnprocessableEntity, err
		}
		query += temp
		if i < len(args)-1 {
			query += ", "
		} else {
			query += ")"
		}
	}
	// Логирование и выполнение функции
	str := ""
	if sdb.log {
		sdb.logger.Info(query)
	}
	err = sdb.db.Raw(query).Scan(&str).Error

	if str != "" {
		code, err := strconv.Atoi(str[:3])
		if err != nil {
			return http.StatusUnprocessableEntity, fmt.Errorf(str)
		}
		return code, fmt.Errorf(str[4:])
	}
	return http.StatusOK, nil
}

func CallRoutine(sdb Sdb, funcName string, args ...any) (statusCode int, err error) {
	return callRoutine(sdb, funcName, args...)
}

func columns[Model ModelTable](sdb Sdb, model Model) ([]string, error) {
	tableName := model.Table()
	dotIndex := strings.Index(tableName, ".")
	schema := tableName[:dotIndex]
	tableName = tableName[dotIndex+1:]

	var res []string
	err := sdb.Table(InformationSchemaColumns).Select("column_name").Limit(-1).GetRecords(fmt.Sprintf("table_schema='%s' and table_name='%s'", schema, tableName), &res)
	return res, err
}

func Columns[Model ModelTable](sdb Sdb, model Model) ([]string, error) {
	return columns(sdb, model)
}

func ReplaceForSqlQuery(str string, isArray bool) string {
	if isArray {
		return strings.ReplaceAll(str, ",", `\,`)
	}
	return fmt.Sprintf("'%s'", strings.ReplaceAll(str, "'", "''"))
}

func MakeSqlWhereFromApiWhere[Model ModelTable](sdb Sdb, model Model, where interface{}) (result string, err error) {

	switch t := where.(type) {
	case qp.QueryParams:
		for _, q := range t {
			if result != "" {
				result += " and "
			}
			if q.Key == "*" {

				names, err := columns(sdb, model)
				if err != nil {
					return "", err
				}

				for i, name := range names {
					if i == 0 {
						result += "("
					} else if i < len(names) {
						result += ", "
					}
					result += name
				}

				q.Type = "::text"
				result += fmt.Sprintf(")%s %s %s", q.Type, q.Sign, q.Value)

			} else {
				result += q.String()
			}
		}
	case qp.QueryParam:
		if t.Key == "*" {

			names, err := columns(sdb, model)
			if err != nil {
				return "", err
			}

			for i, name := range names {
				if i == 0 {
					result += "("
				} else if i < len(names) {
					result += ", "
				}
				result += name
			}

			t.Type = "::text"
			result += fmt.Sprintf(")%s %s %s", t.Type, t.Sign, t.Value)

		} else {
			result += t.String()
		}
	}
	return
}

func ValueToPostgresValue(arg interface{}, isArray bool) (str string, err error) {
	switch t := arg.(type) {
	case nil:
		str = "null"
	case string:
		if t == "null" {
			str = "null"
		} else {
			str = ReplaceForSqlQuery(t, isArray)
		}
	}

	if str == "" {
		v := reflect.ValueOf(arg)
		switch v.Kind() {
		case reflect.Struct, reflect.Func, reflect.Map, reflect.Chan:
			err = errors.New("unsupported type")
		case reflect.Slice, reflect.Array:
			str += "'{"
			for i := 0; i < v.Len(); i++ {
				temp, err := ValueToPostgresValue(v.Index(i).Interface(), true)
				if err != nil {
					return "", err
				}
				str += temp + ","
			}
			str = str[:len(str)-1] + "}'"
		default:
			str = fmt.Sprint(v)
		}
	}
	return
}

func parseFields(fields interface{}) string {
	field := ""
	if fields != nil {
		switch t := fields.(type) {
		case []string:
			field = SliceToString(t)
		case string:
			field = t
		}
	}
	return field
}

func parseOrders(orders interface{}) string {
	order := ""
	if orders != nil {
		switch t := orders.(type) {
		case []string:
			order = SliceToString(t)
		case string:
			order = t
		}
	}
	return order
}

func parseWhere[Model ModelTable](sdb Sdb, model Model, where interface{}) (_where string, err error) {
	switch t := where.(type) {
	case string:
		if len(t) != 0 && len(t) >= 2 {
			if t[0] == '[' && t[len(t)-1] == ']' {
				arr := qp.QueryParams{}
				err = arr.Unmarshal([]byte(t))
				if err != nil {
					return
				}

				_where, err = MakeSqlWhereFromApiWhere(sdb, model, arr)
			} else {
				_where = t
			}
		}
	case qp.QueryParams, qp.QueryParam:
		_where, err = MakeSqlWhereFromApiWhere(sdb, model, t)
		if err != nil {
			return
		}
	}
	return
}

func SliceToString[T any](arr []T) (str string) {
	for _, v := range arr {
		str += fmt.Sprintf("%v,", v)
	}
	if str != "" {
		str = str[:len(str)-1]
	}
	return str
}
