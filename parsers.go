package sdb

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	qp "github.com/daostas/query_parser"
)

func parseQueryParam(sdb Sdb, table string, fields []string, param qp.QueryParam) (result string, err error) {
	if param.Key == "*" {
		if fields[0] == "*" {

			columns, err := ColumnsByTable(sdb, table)
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
				result += fmt.Sprintf(`"%s"%s %s %s`, column, param.Type, param.Sign, param.Value)
			}
			result += ")"

		} else {
			for i, v := range fields {
				if i == 0 {
					result += "("
				} else {
					result += " or "
				}
				result += fmt.Sprintf(`%s%s %s %s`, v, param.Type, param.Sign, param.Value)
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

func parseQueryParams(sdb Sdb, table string, fields []string, params qp.QueryParams) (result string, err error) {
	for _, param := range params {
		str, err := parseQueryParam(sdb, table, fields, param)
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

func MakeSqlWhereFromApiWhere(sdb Sdb, table string, fields []string, where interface{}) (result string, err error) {
	switch t := where.(type) {
	case qp.QueryParams:
		return parseQueryParams(sdb, table, fields, t)
	case *qp.QueryParams:
		return parseQueryParams(sdb, table, fields, *t)
	case qp.QueryParam:
		return parseQueryParam(sdb, table, fields, t)
	case *qp.QueryParam:
		return parseQueryParam(sdb, table, fields, *t)
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
				if i != 0 {
					str += ","
				}
				str += ValueToPostgresValue(v.Index(i).Interface(), true)
			}
			str += "}'"
		default:
			str = fmt.Sprint(v)
		}
	}
	return
}

func _parseFields(fields []string) (res string, resArray []string) {
	for i, v := range fields {
		if i != 0 {
			res += ", "
		}

		r := regexp.MustCompile(`(?i)^(SUM|AVG|MIN|MAX|COUNT)\((.+)\)$`)

		if r.MatchString(v) {
			match := r.FindStringSubmatch(v)
			match[2] = fmt.Sprintf(`"%s"`, strings.ReplaceAll(match[2], `"`, `""`))
			v = fmt.Sprintf("%s(%s)", strings.ToUpper(match[1]), match[2])
		} else {
			v = fmt.Sprintf(`"%s"`, strings.ReplaceAll(v, `"`, `""`))
		}

		res += v
		resArray = append(resArray, v)
	}
	return
}

func parseFields(fields interface{}) (res string, resArray []string) {
	if fields != nil {
		switch t := fields.(type) {
		case []string:
			res, resArray = _parseFields(t)
		case string:
			reg := regexp.MustCompile(`\s*,\s*`)
			matches := reg.Split(t, -1)
			if matches != nil {
				res, resArray = _parseFields(matches)
			}
		}
	}

	if fields == nil || len(resArray) == 0 {
		res = "*"
		resArray = append(resArray, "*")
	}
	return
}

func parseDistinct(distinct interface{}) (res string) {
	if distinct != nil {
		switch t := distinct.(type) {
		case []string:
			res, _ = _parseFields(t)
		case string:
			reg := regexp.MustCompile(`\s*,\s*`)
			matches := reg.Split(t, -1)
			if matches != nil {
				res, _ = _parseFields(matches)
			}
		}
	}

	return
}

func parseOrders(orders interface{}) (res string) {
	if orders != nil {
		r := regexp.MustCompile(`^(\w+)\s*(desc|asc|)$`)
		switch t := orders.(type) {
		case []string:
			for i, v := range t {
				if r.MatchString(v) {
					matches := r.FindStringSubmatch(v)
					if i != 0 {
						res += ", "
					}
					res += fmt.Sprintf(`"%s" %s`, strings.ReplaceAll(matches[1], `"`, `""`), matches[2])
				}
			}
		case string:
			if r.MatchString(t) {
				matches := r.FindStringSubmatch(t)
				res = fmt.Sprintf(`"%s" %s`, strings.ReplaceAll(matches[1], `"`, `""`), matches[2])
			}
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

func parseWhere(sdb Sdb, table string, fields []string, where interface{}) (_where string, err error) {
	switch t := where.(type) {
	case string:
		if len(t) >= 2 {
			if t[0] == '[' && t[len(t)-1] == ']' {
				arr := qp.QueryParams{}
				err = arr.Unmarshal([]byte(t))
				if err != nil {
					return
				}

				_where, err = MakeSqlWhereFromApiWhere(sdb, table, fields, arr)
			} else {
				_where = t
			}
		}
	case qp.QueryParams, qp.QueryParam, *qp.QueryParams, *qp.QueryParam:
		_where, err = MakeSqlWhereFromApiWhere(sdb, table, fields, t)
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
