package sdb

type DeleteQuery[Model ModelTable, Out any] struct {
	query query[Model, Out]
}

func Delete[Model ModelTable](db Sdb, model Model) (q DeleteQuery[Model, Model]) {
	q.query.setQuery(db, model, deleteType)
	return
}

func (q DeleteQuery[Model, Out]) Where(where interface{}) DeleteQuery[Model, Out] {
	q.query.setWhere(where)
	return q
}

func (q DeleteQuery[Model, Out]) Exec() (err error) {
	_, err = q.query.exec()
	return
}
