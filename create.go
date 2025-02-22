package sdb

type CreateQuery[Model ModelTable, Out any] struct {
	query query[Model, Out]
}

func Create[Model ModelTable](db Sdb, model Model) (q CreateQuery[Model, Model]) {
	q.query.setQuery(db, model, create)
	return
}

func (q CreateQuery[Model, Out]) Values(values Map) CreateQuery[Model, Out] {
	q.query.setValues(values)
	return q
}

func (q CreateQuery[Model, Out]) Exec() (err error) {
	_, err = q.query.exec()
	return
}
