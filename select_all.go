package sdb

type SelectQuery[Model ModelTable, Out any] struct {
	query query[Model, Out]
}

func SelectAll[Model ModelTable](db Sdb, model Model) (q SelectQuery[Model, []*Model]) {
	q.query.setQuery(db, model, selectAll)
	return
}

func (q SelectQuery[Model, Out]) Fields(fields interface{}) SelectQuery[Model, Out] {
	q.query.setFields(fields)
	return q
}

func (q SelectQuery[Model, Out]) Orders(orders interface{}) SelectQuery[Model, Out] {
	q.query.setOrders(orders)
	return q
}

func (q SelectQuery[Model, Out]) Where(where interface{}) SelectQuery[Model, Out] {
	q.query.setWhere(where)
	return q
}

func (q SelectQuery[Model, Out]) Limit(limit interface{}) SelectQuery[Model, Out] {
	q.query.setLimit(limit)
	return q
}

func (q SelectQuery[Model, Out]) Offset(offset interface{}) SelectQuery[Model, Out] {
	q.query.setOffset(offset)
	return q
}

func (q SelectQuery[Model, Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
