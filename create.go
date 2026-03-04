package sdb

type CreateQuery[Out any] struct {
	query query[Out]
}

func Create[Model ModelTable](db Sdb, model Model) (q CreateQuery[Model]) {
	q.query.setQuery(db, model, create)
	return
}

func (q CreateQuery[Out]) Values(values Map) CreateQuery[Out] {
	q.query.setValues(values)
	return q
}

func (q CreateQuery[Out]) Exec() (err error) {
	_, err = q.query.exec()
	return
}
