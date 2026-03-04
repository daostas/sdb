package sdb

type CreateSelectQuery[Out any] struct {
	query query[Out]
}

func CreateSelect[Model ModelTable](db Sdb, model Model) (q CreateSelectQuery[Model]) {
	q.query.setQuery(db, model, createSelect)
	return
}

func (q CreateSelectQuery[Out]) Fields(fields interface{}) CreateSelectQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q CreateSelectQuery[Out]) Values(values Map) CreateSelectQuery[Out] {
	q.query.setValues(values)
	return q
}

func (q CreateSelectQuery[Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
