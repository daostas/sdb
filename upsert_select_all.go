package sdb

type UpsertSelectAllQuery[Out any] struct {
	query query[Out]
}

func UpsertSelectAll[Model ModelTable](db Sdb, model Model) (q UpsertSelectAllQuery[[]*Model]) {
	q.query.setQuery(db, model, upsertSelectAll)
	return
}

func (q UpsertSelectAllQuery[Out]) Fields(fields interface{}) UpsertSelectAllQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q UpsertSelectAllQuery[Out]) Values(values Map, params ...ValuesParam) UpsertSelectAllQuery[Out] {
	q.query.setValues(values, params...)
	return q
}

func (q UpsertSelectAllQuery[Out]) Conflict(conflict interface{}, params ...ConflictParam) UpsertSelectAllQuery[Out] {
	q.query.setConflict(conflict, params...)
	return q
}

func (q UpsertSelectAllQuery[Out]) UpdateValues(values Map, params ...ValuesParam) UpsertSelectAllQuery[Out] {
	q.query.setUpdateValues(values, params...)
	return q
}

func (q UpsertSelectAllQuery[Out]) Where(where interface{}) UpsertSelectAllQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q UpsertSelectAllQuery[Out]) Exec() (err error) {
	_, err = q.query.exec()
	return
}
