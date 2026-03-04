package sdb

type CreateSelectMapQuery[Out any] struct {
	query query[Out]
}

func CreateSelectMap[Model any](db Sdb, model Model) (q CreateSelectMapQuery[map[string]interface{}]) {
	q.query.setQuery(db, model, createSelect)
	return
}

func (q CreateSelectMapQuery[Out]) Fields(fields interface{}) CreateSelectMapQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q CreateSelectMapQuery[Out]) Values(values Map) CreateSelectMapQuery[Out] {
	q.query.setValues(values)
	return q
}

func (q CreateSelectMapQuery[Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
