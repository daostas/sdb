package sdb

type SelectAllQuery[Out any] struct {
	query query[Out]
}

func SelectAll[Model ModelTable](db Sdb, model Model) (q SelectAllQuery[[]*Model]) {
	q.query.setQuery(db, model, selectAll)
	return
}

func (q SelectAllQuery[Out]) Distinct(distinct interface{}) SelectAllQuery[Out] {
	q.query.setDistinct(distinct)
	return q
}

func (q SelectAllQuery[Out]) Fields(fields interface{}) SelectAllQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q SelectAllQuery[Out]) Orders(orders interface{}) SelectAllQuery[Out] {
	q.query.setOrders(orders)
	return q
}

func (q SelectAllQuery[Out]) Where(where interface{}) SelectAllQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q SelectAllQuery[Out]) Limit(limit interface{}) SelectAllQuery[Out] {
	q.query.setLimit(limit)
	return q
}

func (q SelectAllQuery[Out]) Offset(offset interface{}) SelectAllQuery[Out] {
	q.query.setOffset(offset)
	return q
}

func (q SelectAllQuery[Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
