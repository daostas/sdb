package sdb

type CreateSelectQuery[Model ModelTable, Out any] struct {
	query query[Model, Out]
}

func CreateSelect[Model ModelTable](db Sdb, model Model) (q CreateSelectQuery[Model, Model]) {
	q.query.setQuery(db, model, createSelect)
	return
}

func (q CreateSelectQuery[Model, Out]) Fields(fields interface{}) CreateSelectQuery[Model, Out] {
	q.query.setFields(fields)
	return q
}

func (q CreateSelectQuery[Model, Out]) Values(values Map) CreateSelectQuery[Model, Out] {
	q.query.setValues(values)
	return q
}

func (q CreateSelectQuery[Model, Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
