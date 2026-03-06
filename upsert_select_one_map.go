package sdb

type UpsertSelectOneMapQuery[Out any] struct {
	query query[Out]
}

func UpsertSelectOneMap[Model ModelTable](db Sdb, model Model) (q UpsertSelectOneMapQuery[[]map[string]interface{}]) {
	q.query.setQuery(db, model, upsertSelectOne)
	return
}

func (q UpsertSelectOneMapQuery[Out]) Fields(fields interface{}) UpsertSelectOneMapQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q UpsertSelectOneMapQuery[Out]) Values(values Map, params ...ValuesParam) UpsertSelectOneMapQuery[Out] {
	q.query.setValues(values, params...)
	return q
}

func (q UpsertSelectOneMapQuery[Out]) Conflict(conflict interface{}, params ...ConflictParam) UpsertSelectOneMapQuery[Out] {
	q.query.setConflict(conflict, params...)
	return q
}

func (q UpsertSelectOneMapQuery[Out]) UpdateValues(values Map, params ...ValuesParam) UpsertSelectOneMapQuery[Out] {
	q.query.setUpdateValues(values, params...)
	return q
}

func (q UpsertSelectOneMapQuery[Out]) Where(where interface{}) UpsertSelectOneMapQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q UpsertSelectOneMapQuery[Out]) Exec() (err error) {
	_, err = q.query.exec()
	return
}
