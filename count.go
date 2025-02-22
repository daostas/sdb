package sdb

type CountQuery[Model ModelTable, Out any] struct {
	query query[Model, Out]
}

func Count[Model ModelTable](db Sdb, model Model) (q CountQuery[Model, int]) {
	q.query.setQuery(db, model, count)
	return
}

func (q CountQuery[Model, Out]) Fields(fields interface{}) CountQuery[Model, Out] {
	q.query.setFields(fields)
	return q
}

func (q CountQuery[Model, Out]) Where(where interface{}) CountQuery[Model, Out] {
	q.query.setWhere(where)
	return q
}

func (q CountQuery[Model, Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
