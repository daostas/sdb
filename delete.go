package sdb

type DeleteQuery[Out any] struct {
	query query[Out]
}

func Delete[Model ModelTable](db Sdb, model Model) (q DeleteQuery[Model]) {
	q.query.setQuery(db, model, deleteType)
	return
}

func (q DeleteQuery[Out]) Where(where interface{}) DeleteQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q DeleteQuery[Out]) Exec() (err error) {
	_, err = q.query.exec()
	return
}
