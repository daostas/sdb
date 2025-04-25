package sdb

import (
	"encoding/json"
	"fmt"
	qp "github.com/daostas/query_parser"
	"reflect"
	"strings"
)

func parseQueryParam[Model ModelTable](sdb Sdb, model Model, fields string, param qp.QueryParam) (result string, err error) {
	if param.Key == "*" {
		if fields == "*" {

			columns, err := Columns(sdb, model)
			if err != nil {
				return "", err
			}

			param.Type = "::text"

			for i, column := range columns {
				if i == 0 {
					result += "("
				} else if i < len(columns) {
					result += " or "
				}
				result += fmt.Sprintf("%s%s %s %s", column, param.Type, param.Sign, param.Value)
			}
			result += ")"

		} else {
			arr := strings.Split(fields, ",")
			for i, v := range arr {
				if i == 0 {
					result += "("
				} else if i < len(arr) {
					result += " or "
				}
				result += fmt.Sprintf("%s%s %s %s", v, param.Type, param.Sign, param.Value)
			}
			result += ")"
		}
	} else {
		result += param.String()
	}

	return
}

//func parseQueryParam[Model ModelTable](sdb Sdb, model Model, fields string, param qp.QueryParam) (result string, err error) {
//	if param.Key == "*" {
//		if fields == "*" {
//			columns, err := Columns(sdb, model)
//			if err != nil {
//				return "", err
//			}
//
//			for i, name := range columns {
//				if i == 0 {
//					result += "("
//				} else if i < len(columns) {
//					result += ", "
//				}
//				result += name
//			}
//			result += ")"
//		} else {
//			result += fmt.Sprintf("(%s)", fields)
//		}
//
//		param.Type = "::text"
//		result += fmt.Sprintf("%s %s %s", param.Type, param.Sign, param.Value)
//
//	} else {
//		result += param.String()
//	}
//	return
//}

func parseQueryParams[Model ModelTable](sdb Sdb, model Model, fields string, params qp.QueryParams) (result string, err error) {
	for _, param := range params {
		str, err := parseQueryParam(sdb, model, fields, param)
		if err != nil {
			return "", err
		}

		if result != "" {
			result += " and "
		}
		result += str
	}
	return
}

func ReplaceForSqlQuery(str string, isArray ...bool) string {
	if len(isArray) != 0 && isArray[0] {
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(str, ",", `\,`))
	}
	str = strings.ReplaceAll(str, "'", "''")
	str = fmt.Sprintf("'%s'", str)
	return str
}

func MakeSqlWhereFromApiWhere[Model ModelTable](sdb Sdb, model Model, fields string, where interface{}) (result string, err error) {
	switch t := where.(type) {
	case qp.QueryParams:
		return parseQueryParams(sdb, model, fields, t)
	case *qp.QueryParams:
		return parseQueryParams(sdb, model, fields, *t)
	case qp.QueryParam:
		return parseQueryParam(sdb, model, fields, t)
	case *qp.QueryParam:
		return parseQueryParam(sdb, model, fields, *t)
	}
	return
}

func ValueToPostgresValue(arg interface{}, isArray ...bool) (str string) {
	switch t := arg.(type) {
	case nil:
		str = "null"
	case string:
		if t == "null" || t == "default" || t == "true" || t == "false" {
			str = t
		} else {
			str = ReplaceForSqlQuery(t, isArray...)
		}
	}

	if str == "" {
		v := reflect.ValueOf(arg)
		switch v.Kind() {
		case reflect.Struct, reflect.Map:
			data, _ := json.Marshal(arg)
			str = ReplaceForSqlQuery(string(data), isArray...)
		case reflect.Slice, reflect.Array:
			str += "'{"
			for i := 0; i < v.Len(); i++ {
				temp := ValueToPostgresValue(v.Index(i).Interface(), true)
				str += temp + ","
			}
			str = str[:len(str)-1] + "}'"
		default:
			str = fmt.Sprint(v)
		}
	}
	return
}

func parseFields(fields interface{}) (res string) {
	if fields != nil {
		switch t := fields.(type) {
		case []string:
			for _, v := range t {
				res += fmt.Sprintf("%v, ", v)
			}
			if res != "" {
				res = res[:len(res)-2]
			}
		case string:
			res = t
		}
	}
	if fields == nil || len(res) == 0 {
		res = "*"
	}
	return
}

func parseOrders(orders interface{}) (res string) {
	if orders != nil {
		switch t := orders.(type) {
		case []string:
			for _, v := range t {
				res += fmt.Sprintf("%v, ", v)
			}
			if res != "" {
				res = res[:len(res)-2]
			}
		case string:
			res = t
		}
	}
	return
}

func NewWhere(where string, args ...interface{}) string {
	for _, v := range args {
		where = strings.Replace(where, "?", parseValue(v), 1)
	}
	return where
}

func parseWhere[Model ModelTable](sdb Sdb, model Model, fields string, where interface{}) (_where string, err error) {
	switch t := where.(type) {
	case string:
		if len(t) >= 2 {
			if t[0] == '[' && t[len(t)-1] == ']' {
				arr := qp.QueryParams{}
				err = arr.Unmarshal([]byte(t))
				if err != nil {
					return
				}

				_where, err = MakeSqlWhereFromApiWhere(sdb, model, fields, arr)
			} else {
				_where = t
			}
		}
	case qp.QueryParams, qp.QueryParam, *qp.QueryParams, *qp.QueryParam:
		_where, err = MakeSqlWhereFromApiWhere(sdb, model, fields, t)
		if err != nil {
			return
		}
	}
	return
}

func parseValue(value interface{}) (s string) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func:
	case reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice, reflect.Array:
		if value != nil && !v.IsNil() {
			s = ValueToPostgresValue(reflect.Indirect(v).Interface())
		} else {
			s = "null"
		}
	default:
		s = fmt.Sprintf("%v", value)
		if value != nil && s != "<nil>" {
			s = ValueToPostgresValue(reflect.Indirect(v).Interface())
		} else {
			s = "null"
		}
	}
	return
}

func parseLimit(limit interface{}) (i int) {
	if limit == nil {
		i = 0
	} else {
		switch v := limit.(type) {
		case int:
			i = v
		case int32:
			i = int(v)
		}
	}
	return
}

func parseOffset(offset interface{}) (i int) {
	if offset == nil {
		i = 0
	} else {
		switch v := offset.(type) {
		case int:
			i = v
		case int32:
			i = int(v)
		}
	}
	return
}
