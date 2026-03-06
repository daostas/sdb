package sdb

type UpdateSelectOneMapQuery[Out any] struct {
	query query[Out]
}

func UpdateSelectOneMap[Model any](db Sdb, model Model) (q UpdateSelectOneMapQuery[map[string]interface{}]) {
	q.query.setQuery(db, model, updateSelectOne)
	return
}

func (q UpdateSelectOneMapQuery[Out]) Fields(fields interface{}) UpdateSelectOneMapQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q UpdateSelectOneMapQuery[Out]) Values(values Map, params ...ValuesParam) UpdateSelectOneMapQuery[Out] {
	q.query.setValues(values, params...)
	return q
}

func (q UpdateSelectOneMapQuery[Out]) Where(where interface{}) UpdateSelectOneMapQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q UpdateSelectOneMapQuery[Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
