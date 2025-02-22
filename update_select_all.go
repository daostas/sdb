package sdb

type UpdateSelectAllQuery[Model ModelTable, Out any] struct {
	query query[Model, Out]
}

func UpdateSelectAll[Model ModelTable](db Sdb, model Model) (q UpdateSelectAllQuery[Model, []*Model]) {
	q.query.setQuery(db, model, updateSelectAll)
	return
}

func (q UpdateSelectAllQuery[Model, Out]) Fields(fields interface{}) UpdateSelectAllQuery[Model, Out] {
	q.query.setFields(fields)
	return q
}

func (q UpdateSelectAllQuery[Model, Out]) Values(values Map) UpdateSelectAllQuery[Model, Out] {
	q.query.setValues(values)
	return q
}

func (q UpdateSelectAllQuery[Model, Out]) Where(where interface{}) UpdateSelectAllQuery[Model, Out] {
	q.query.setWhere(where)
	return q
}

func (q UpdateSelectAllQuery[Model, Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
