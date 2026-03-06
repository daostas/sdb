package sdb

type UpdateSelectAllMapQuery[Out any] struct {
	query query[Out]
}

func UpdateSelectAllMap[Model any](db Sdb, model Model) (q UpdateSelectAllMapQuery[[]map[string]interface{}]) {
	q.query.setQuery(db, model, updateSelectAll)
	return
}

func (q UpdateSelectAllMapQuery[Out]) Fields(fields interface{}) UpdateSelectAllMapQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q UpdateSelectAllMapQuery[Out]) Values(values Map, params ...ValuesParam) UpdateSelectAllMapQuery[Out] {
	q.query.setValues(values, params...)
	return q
}

func (q UpdateSelectAllMapQuery[Out]) Where(where interface{}) UpdateSelectAllMapQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q UpdateSelectAllMapQuery[Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
