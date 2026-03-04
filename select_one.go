package sdb

type SelectOneQuery[Out any] struct {
	query query[Out]
}

func SelectOne[Model ModelTable](db Sdb, model Model) (q SelectOneQuery[Model]) {
	q.query.setQuery(db, model, selectOne)
	q.query.setLimit(1)
	return
}

func (q SelectOneQuery[Out]) Fields(fields interface{}) SelectOneQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q SelectOneQuery[Out]) Where(where interface{}) SelectOneQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q SelectOneQuery[Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
