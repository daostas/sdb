package sdb

type UpdateSelectAllQuery[Out any] struct {
	query query[Out]
}

func UpdateSelectAll[Model ModelTable](db Sdb, model Model) (q UpdateSelectAllQuery[[]*Model]) {
	q.query.setQuery(db, model, updateSelectAll)
	return
}

func (q UpdateSelectAllQuery[Out]) Fields(fields interface{}) UpdateSelectAllQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q UpdateSelectAllQuery[Out]) Values(values Map, ignoreNull ...bool) UpdateSelectAllQuery[Out] {
	q.query.setValues(values, ignoreNull...)
	return q
}

func (q UpdateSelectAllQuery[Out]) Where(where interface{}) UpdateSelectAllQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q UpdateSelectAllQuery[Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
