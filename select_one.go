package sdb

type SelectOneQuery[Model ModelTable, Out any] struct {
	query query[Model, Out]
}

func SelectOne[Model ModelTable](db Sdb, model Model) (q SelectOneQuery[Model, Model]) {
	q.query.setQuery(db, model, selectOne)
	q.query.setLimit(1)
	return
}

func (q SelectOneQuery[Model, Out]) Fields(fields interface{}) SelectOneQuery[Model, Out] {
	q.query.setFields(fields)
	return q
}

func (q SelectOneQuery[Model, Out]) Where(where interface{}) SelectOneQuery[Model, Out] {
	q.query.setWhere(where)
	return q
}

func (q SelectOneQuery[Model, Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
