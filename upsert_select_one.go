package sdb

type UpsertSelectOneQuery[Out any] struct {
	query query[Out]
}

func UpsertSelectOne[Model ModelTable](db Sdb, model Model) (q UpsertSelectOneQuery[[]*Model]) {
	q.query.setQuery(db, model, upsertSelectOne)
	return
}

func (q UpsertSelectOneQuery[Out]) Fields(fields interface{}) UpsertSelectOneQuery[Out] {
	q.query.setFields(fields)
	return q
}

func (q UpsertSelectOneQuery[Out]) Values(values Map, params ...ValuesParam) UpsertSelectOneQuery[Out] {
	q.query.setValues(values, params...)
	return q
}

func (q UpsertSelectOneQuery[Out]) Conflict(conflict interface{}, params ...ConflictParam) UpsertSelectOneQuery[Out] {
	q.query.setConflict(conflict, params...)
	return q
}

func (q UpsertSelectOneQuery[Out]) UpdateValues(values Map, params ...ValuesParam) UpsertSelectOneQuery[Out] {
	q.query.setUpdateValues(values, params...)
	return q
}

func (q UpsertSelectOneQuery[Out]) Where(where interface{}) UpsertSelectOneQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q UpsertSelectOneQuery[Out]) Exec() (err error) {
	_, err = q.query.exec()
	return
}
