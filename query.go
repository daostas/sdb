package sdb

import (
	"fmt"
	"reflect"
	"strconv"
)

type (
	query[Out any] struct {
		table            string
		queryType        QueryType
		fields           interface{}
		orders           interface{}
		where            interface{}
		limit            interface{}
		offset           interface{}
		values           map[string]string
		updateValues     map[string]string
		conflict         interface{}
		doNothing        bool
		exclude          bool
		optimizeUpdate   bool
		ignoreNullValues bool
		db               Sdb
	}

	Map           map[string]interface{}
	QueryType     int
	ConflictParam int
	ValuesParam   int
)

var (
	queryTypeMap = map[QueryType]string{
		selectAll:       "SELECT ",
		selectOne:       "SELECT ",
		create:          "INSERT INTO ",
		createSelect:    "INSERT INTO ",
		upsert:          "INSERT INTO ",
		upsertSelectAll: "INSERT INTO ",
		upsertSelectOne: "INSERT INTO ",
		update:          "UPDATE ",
		updateSelectAll: "UPDATE ",
		updateSelectOne: "UPDATE ",
		deleteType:      "DELETE ",
		count:           "SELECT COUNT(%s)",
	}
	queryTypeConst = map[string]map[QueryType]bool{
		"fields": {
			selectAll: true,
			selectOne: true,
			count:     true,
		},
		"from": {
			selectAll:  true,
			selectOne:  true,
			deleteType: true,
			count:      true,
		},
		"orders": {
			selectAll: true,
		},
		"where": {
			selectAll:       true,
			selectOne:       true,
			update:          true,
			updateSelectAll: true,
			updateSelectOne: true,
			deleteType:      true,
			count:           true,
			upsert:          true,
			upsertSelectAll: true,
			upsertSelectOne: true,
		},
		"limit": {
			selectAll: true,
		},
		"offset": {
			selectAll: true,
		},
		"values": {
			create:          true,
			createSelect:    true,
			update:          true,
			updateSelectAll: true,
			updateSelectOne: true,
			upsert:          true,
		},
		"set": {
			update:          true,
			updateSelectAll: true,
			updateSelectOne: true,
		},
		"returning": {
			createSelect:    true,
			updateSelectAll: true,
			updateSelectOne: true,
		},
		"on conflict": {
			upsert: true,
		},
		"set1": {
			upsert: true,
		},
	}
)

const (
	selectAll QueryType = iota
	selectOne
	create
	createSelect
	update
	updateSelectAll
	updateSelectOne
	deleteType
	count
	routine
	upsert
	upsertSelectAll
	upsertSelectOne
)

const (
	DoNothing ConflictParam = iota
	Excluded
)

const (
	IgnoreNullValues ValuesParam = iota
	OptimizeUpdate
)

func (q *query[Out]) setQuery(db Sdb, table any, queryType QueryType) {
	if model, ok := table.(ModelTable); ok {
		q.table = model.Table()
	} else {
		switch t := table.(type) {
		case string:
			q.table = t
		}
	}
	q.db = db
	q.queryType = queryType
}

func (q *query[Out]) setFields(fields interface{}) *query[Out] {
	q.fields = fields
	return q
}

func (q *query[Out]) setOrders(orders interface{}) *query[Out] {
	q.orders = orders
	return q
}

func (q *query[Out]) setWhere(where interface{}) *query[Out] {
	q.where = where
	return q
}

func (q *query[Out]) setLimit(limit interface{}) *query[Out] {
	q.limit = limit
	return q
}

func (q *query[Out]) setOffset(offset interface{}) *query[Out] {
	if offset == nil {
		q.offset = 0
	} else {
		switch v := offset.(type) {
		case int:
			q.offset = v
		case int32:
			q.offset = int(v)
		}
	}
	return q
}

func (q *query[Out]) setValues(values Map, params ...ValuesParam) *query[Out] {
	if q.values == nil {
		q.values = make(map[string]string)
	}
	q.setValuesParams(params...)

	q._setValues(q.values, values)
	return q
}

func (q *query[Out]) setUpdateValues(updateValues Map, params ...ValuesParam) *query[Out] {
	if q.updateValues == nil {
		q.updateValues = make(map[string]string)
	}

	q._setValues(q.updateValues, updateValues)

	return q
}

func (q *query[Out]) _setValues(dest map[string]string, values Map) {

	for k, value := range values {
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Chan, reflect.Func:
		case reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice, reflect.Array:
			if value != nil && !v.IsNil() {
				dest[k] = ValueToPostgresValue(reflect.Indirect(v).Interface())
			} else {
				if !q.ignoreNullValues {
					dest[k] = "null"
				}
			}
		default:
			s := fmt.Sprintf("%v", value)
			if value != nil && s != "<nil>" {
				dest[k] = ValueToPostgresValue(reflect.Indirect(v).Interface())
			} else {
				if !q.ignoreNullValues {
					dest[k] = "null"
				}
			}
		}
	}
}

func (q *query[Out]) setConflict(conflict interface{}, params ...ConflictParam) *query[Out] {
	q.conflict = conflict
	q.setConflictParams(params...)
	return q
}

func (q *query[Out]) setValuesParams(params ...ValuesParam) *query[Out] {
	for _, v := range params {
		switch v {
		case IgnoreNullValues:
			q.ignoreNullValues = true
		case OptimizeUpdate:
			q.optimizeUpdate = true
		}
	}

	return q
}

func (q *query[Out]) setConflictParams(params ...ConflictParam) *query[Out] {
	for _, v := range params {
		switch v {
		case DoNothing:
			q.doNothing = true
		case Excluded:
			q.exclude = true
		}
	}

	return q
}

func (q *query[Out]) exec() (out Out, err error) {
	query, err := q.getQuery()
	if err != nil {
		return
	}

	if q.db.log {
		if q.db.logInFile {
			q.db.logger.InfoFl(query)
		} else {
			q.db.logger.Info(query)
		}
	}

	switch q.queryType {
	case create, update, deleteType:
		err = q.db.db.Exec(query).Error
	case selectOne:
		t := reflect.TypeOf(out)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		if t.Kind() == reflect.Map {
			var tempOut []Out
			err = q.db.db.Raw(query).Scan(&tempOut).Error
			if err != nil {
				return
			}
			if len(tempOut) == 0 {
				err = ErrRecordNotFound
			} else {
				out = tempOut[0]
			}
		} else {
			var tempOut []*Out
			err = q.db.db.Raw(query).Scan(&tempOut).Error
			if err != nil {
				return
			}
			if len(tempOut) == 0 {
				err = ErrRecordNotFound
			} else {
				out = *tempOut[0]
			}
		}
	default:
		err = q.db.db.Raw(query).Scan(&out).Error
	}

	if err != nil {
		if q.db.log {
			if q.db.logInFile {
				q.db.logger.ErrorFl(err)
			} else {
				q.db.logger.Err(err)
			}
		}
	}
	return
}

func (q query[Out]) getQuery() (s string, err error) {
	s = queryTypeMap[q.queryType]

	//Fields
	fields, fieldsArray := parseFields(q.fields)
	if queryTypeConst["fields"][q.queryType] {
		fields := fields
		if q.queryType == count {
			if fields != "*" {
				fields = fmt.Sprintf("(%s)", fields)
			}
			s = fmt.Sprintf(s+" ", fields)
		} else {
			s += fields + " "
		}
	}

	//From and table name
	if queryTypeConst["from"][q.queryType] {
		s += "FROM "
	}
	s += q.table + " "

	conflict, conflictArray := parseFields(q.conflict)
	//Values
	for round := 0; round < 2; round++ {
		if queryTypeConst["on conflict"][q.queryType] {
			if round == 1 {
				s += fmt.Sprintf("ON CONFLICT (%s) ", conflict)
				if q.doNothing {
					s += "DO NOTHING "
					continue
				}
			}
		} else if round == 1 {
			continue
		}

		if queryTypeConst["values"][q.queryType] {
			i := 0
			if queryTypeConst["set"][q.queryType] || (queryTypeConst["set1"][q.queryType] && round == 1) {

				if round == 1 {
					s += "DO UPDATE "
					if q.updateValues != nil {
						q.values = q.updateValues
					}
				}

				s += "SET "
			valueLoop:
				for k, v := range q.values {
					if round == 1 {
						for _, field := range conflictArray {
							if field == fmt.Sprintf(`"%s"`, k) {
								delete(q.values, k)
								continue valueLoop
							}
						}

						if q.exclude {
							v = fmt.Sprintf("EXCLUDED.%s", k)
						}

					}
					if i != 0 {
						s += ", "
					}
					s += fmt.Sprintf("%s = %s", k, v)
					i++
				}
				s += " "
			} else {
				var (
					template = "("
					values   = "("
				)
				for k, v := range q.values {
					if i != 0 {
						template += ", "
						values += ", "
					}
					template += k
					values += v
					i++
				}
				template += ")"
				values += ")"
				s += fmt.Sprintf("%s VALUES %s ", template, values)
			}
		}

	}

	//Where
	if queryTypeConst["where"][q.queryType] {
		where, err := parseWhere(q.db, q.table, fieldsArray, q.where)
		if err != nil {
			return s, err
		}
		if where != "" {
			if q.optimizeUpdate && (queryTypeConst["set"][q.queryType] || queryTypeConst["set1"][q.queryType]) {
				where = fmt.Sprintf("(%s)", where)
			}
			s += "WHERE " + where + " "
		}

		if (queryTypeConst["set"][q.queryType] || queryTypeConst["set1"][q.queryType]) && len(q.values) != 0 && q.optimizeUpdate {
			keys, values := "(", "("
			i := 0
			for k, v := range q.values {
				if q.exclude {
					v = fmt.Sprintf("EXCLUDED.%s", k)
				}

				if i != 0 {
					keys += ", "
					values += ", "
				}

				if queryTypeConst["set1"][q.queryType] {
					keys += fmt.Sprintf("%s.%s", q.table, k)
				} else {
					keys += k
				}

				values += v
				i++
			}
			keys += ")"
			values += ")"

			if where == "" {
				s += "WHERE "
			} else {
				s += "and "
			}

			s += fmt.Sprintf("(%s IS DISTINCT FROM %s) ", keys, values)
		}

	}

	//Orders
	if queryTypeConst["orders"][q.queryType] {
		orders := parseOrders(q.orders)
		if orders != "" {
			s += "ORDER BY " + orders + " "
		}
	}

	//Limit
	if queryTypeConst["limit"][q.queryType] {
		limit := parseLimit(q.limit)
		if limit != 0 && limit != -1 {
			s += "LIMIT " + strconv.Itoa(limit) + " "
		}
	}

	//Offset
	if queryTypeConst["offset"][q.queryType] {
		offset := parseOffset(q.offset)
		if offset != 0 && offset != -1 {
			s += "OFFSET " + strconv.Itoa(offset) + " "
		}
	}

	//Returning
	if queryTypeConst["returning"][q.queryType] {
		s += "RETURNING " + fields
	}
	return
}
