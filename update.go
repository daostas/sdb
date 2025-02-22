package sdb

type UpdateQuery[Model ModelTable, Out any] struct {
	query query[Model, Out]
}

func Update[Model ModelTable](db Sdb, model Model) (q UpdateQuery[Model, Model]) {
	q.query.setQuery(db, model, update)
	return
}

func (q UpdateQuery[Model, Out]) Values(values Map) UpdateQuery[Model, Out] {
	q.query.setValues(values)
	return q
}

func (q UpdateQuery[Model, Out]) Where(where interface{}) UpdateQuery[Model, Out] {
	q.query.setWhere(where)
	return q
}

func (q UpdateQuery[Model, Out]) Exec() (err error) {
	_, err = q.query.exec()
	return
}
