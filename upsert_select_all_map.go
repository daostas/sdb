package sdb

type UpsertSelectAllMapQuery[Out any] struct {
	query query[Out]
}

func UpsertSelectAllMap[Model ModelTable](db Sdb, model Model) (q UpsertSelectAllMapQuery[[]map[string]interface{}]) {
	q.query.setQuery(db, model, upsertSelectAll)
	return
}

func (q UpsertSelectAllMapQuery[Out]) Fields(fields interface{}) UpsertSelectAllMapQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q UpsertSelectAllMapQuery[Out]) Values(values Map, params ...ValuesParam) UpsertSelectAllMapQuery[Out] {
	q.query.setValues(values, params...)
	return q
}

func (q UpsertSelectAllMapQuery[Out]) Conflict(conflict interface{}, params ...ConflictParam) UpsertSelectAllMapQuery[Out] {
	q.query.setConflict(conflict, params...)
	return q
}

func (q UpsertSelectAllMapQuery[Out]) UpdateValues(values Map, params ...ValuesParam) UpsertSelectAllMapQuery[Out] {
	q.query.setUpdateValues(values, params...)
	return q
}

func (q UpsertSelectAllMapQuery[Out]) Where(where interface{}) UpsertSelectAllMapQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q UpsertSelectAllMapQuery[Out]) Exec() (err error) {
	_, err = q.query.exec()
	return
}
