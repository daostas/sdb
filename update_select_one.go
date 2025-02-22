package sdb

type UpdateSelectOneQuery[Model ModelTable, Out any] struct {
	query query[Model, Out]
}

func UpdateSelectOne[Model ModelTable](db Sdb, model Model) (q UpdateSelectOneQuery[Model, Model]) {
	q.query.setQuery(db, model, updateSelectOne)
	return
}

func (q UpdateSelectOneQuery[Model, Out]) Fields(fields interface{}) UpdateSelectOneQuery[Model, Out] {
	q.query.setFields(fields)
	return q
}

func (q UpdateSelectOneQuery[Model, Out]) Values(values Map) UpdateSelectOneQuery[Model, Out] {
	q.query.setValues(values)
	return q
}

func (q UpdateSelectOneQuery[Model, Out]) Where(where interface{}) UpdateSelectOneQuery[Model, Out] {
	q.query.setWhere(where)
	return q
}

func (q UpdateSelectOneQuery[Model, Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
