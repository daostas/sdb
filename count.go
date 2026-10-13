package sdb

type CountQuery[Out any] struct {
	query query[Out]
}

func Count[Model any](db Sdb, model Model) (q CountQuery[int]) {
	q.query.setQuery(db, model, count)
	return
}

func (q CountQuery[Out]) Fields(fields interface{}) CountQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q CountQuery[Out]) Where(where interface{}, args ...interface{}) CountQuery[Out] {
	q.query.setWhere(where, args...)
	return q
}

func (q CountQuery[Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
