package sdb

type UpdateSelectOneQuery[Out any] struct {
	query query[Out]
}

func UpdateSelectOne[Model ModelTable](db Sdb, model Model) (q UpdateSelectOneQuery[Model]) {
	q.query.setQuery(db, model, updateSelectOne)
	return
}

func (q UpdateSelectOneQuery[Out]) Fields(fields interface{}) UpdateSelectOneQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q UpdateSelectOneQuery[Out]) Values(values Map, params ...ValuesParam) UpdateSelectOneQuery[Out] {
	q.query.setValues(values, params...)
	return q
}

func (q UpdateSelectOneQuery[Out]) Where(where interface{}) UpdateSelectOneQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q UpdateSelectOneQuery[Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
