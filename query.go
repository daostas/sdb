package sdb

import (
	"fmt"
	"gorm.io/gorm"
	"strconv"
)

type (
	query[Model ModelTable, Out any] struct {
		model     Model
		queryType QueryType
		fields    interface{}
		orders    interface{}
		where     interface{}
		limit     interface{}
		offset    interface{}
		values    map[string]string
		db        Sdb
	}

	Map       map[string]interface{}
	QueryType int
)

var (
	queryTypeMap = map[QueryType]string{
		selectAll:       "SELECT ",
		selectOne:       "SELECT ",
		create:          "INSERT INTO ",
		createSelect:    "INSERT INTO ",
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
)

func (q *query[Model, Out]) setQuery(db Sdb, model Model, queryType QueryType) {
	q.db = db
	q.model = model
	q.queryType = queryType
}

func (q *query[Model, Out]) setFields(fields interface{}) *query[Model, Out] {
	q.fields = fields
	return q
}

func (q *query[Model, Out]) setOrders(orders interface{}) *query[Model, Out] {
	q.orders = orders
	return q
}

func (q *query[Model, Out]) setWhere(where interface{}) *query[Model, Out] {
	q.where = where
	return q
}

func (q *query[Model, Out]) setLimit(limit interface{}) *query[Model, Out] {
	q.limit = limit
	return q
}

func (q *query[Model, Out]) setOffset(offset interface{}) *query[Model, Out] {
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

func (q *query[Model, Out]) setValues(values Map) *query[Model, Out] {
	if q.values == nil {
		q.values = make(map[string]string)
	}

	for k, v := range values {
		q.values[k] = parseValue(v)
	}
	return q
}

func (q *query[Model, Out]) exec() (out Out, err error) {
	query, err := q.getQuery()
	if err != nil {
		return
	}

	if q.db.log {
		q.db.logger.Info(query)
	}

	switch q.queryType {
	case create, update, deleteType:
		err = q.db.db.Exec(query).Error
	case selectOne:
		var tempOut []*Out
		err = q.db.db.Raw(query).Scan(&tempOut).Error
		if err != nil {
			return
		}
		if len(tempOut) == 0 {
			err = gorm.ErrRecordNotFound
		} else {
			out = *tempOut[0]
		}
	default:
		err = q.db.db.Raw(query).Scan(&out).Error
	}
	return
}

func (q query[Model, Out]) getQuery() (s string, err error) {
	s = queryTypeMap[q.queryType]

	//Fields
	fields := parseFields(q.fields)
	if queryTypeConst["fields"][q.queryType] {
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
	s += q.model.Table() + " "

	//Values
	if queryTypeConst["values"][q.queryType] {
		i := 0
		if queryTypeConst["set"][q.queryType] {
			s += "SET "
			for k, v := range q.values {
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

	//Where
	if queryTypeConst["where"][q.queryType] {
		where, err := parseWhere(q.db, q.model, fields, q.where)
		if err != nil {
			return s, err
		}
		if where != "" {
			s += "WHERE " + where + " "
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
