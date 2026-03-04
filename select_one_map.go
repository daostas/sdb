package sdb

type SelectOneMapQuery[Out any] struct {
	query query[Out]
}

func SelectOneMap[Model any](db Sdb, model Model) (q SelectOneMapQuery[map[string]interface{}]) {
	q.query.setQuery(db, model, selectOne)
	q.query.setLimit(1)
	return
}

func (q SelectOneMapQuery[Out]) Fields(fields interface{}) SelectOneMapQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q SelectOneMapQuery[Out]) Where(where interface{}) SelectOneMapQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q SelectOneMapQuery[Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
