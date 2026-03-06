package sdb

type UpdateQuery[Out any] struct {
	query query[Out]
}

func Update[Model ModelTable](db Sdb, model Model) (q UpdateQuery[Model]) {
	q.query.setQuery(db, model, update)
	return
}

func (q UpdateQuery[Out]) Values(values Map, params ...ValuesParam) UpdateQuery[Out] {
	params = append(params, OptimizeUpdate)
	q.query.setValues(values, params...)
	return q
}

func (q UpdateQuery[Out]) Where(where interface{}) UpdateQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q UpdateQuery[Out]) Exec() (err error) {
	_, err = q.query.exec()
	return
}
