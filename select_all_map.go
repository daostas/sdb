package sdb

type SelectAllMapQuery[Out any] struct {
	query query[Out]
}

func SelectAllMap[Model any](db Sdb, model Model) (q SelectAllMapQuery[[]map[string]interface{}]) {
	q.query.setQuery(db, model, selectAll)
	return
}

func (q SelectAllMapQuery[Out]) Fields(fields interface{}) SelectAllMapQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q SelectAllMapQuery[Out]) Orders(orders interface{}) SelectAllMapQuery[Out] {
	q.query.setOrders(orders)
	return q
}

func (q SelectAllMapQuery[Out]) Where(where interface{}) SelectAllMapQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q SelectAllMapQuery[Out]) Limit(limit interface{}) SelectAllMapQuery[Out] {
	q.query.setLimit(limit)
	return q
}

func (q SelectAllMapQuery[Out]) Offset(offset interface{}) SelectAllMapQuery[Out] {
	q.query.setOffset(offset)
	return q
}

func (q SelectAllMapQuery[Out]) Exec() (out Out, err error) {
	out, err = q.query.exec()
	return
}
