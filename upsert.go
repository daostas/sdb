package sdb

type UpsertQuery[Out any] struct {
	query query[Out]
}

func Upsert[Model ModelTable](db Sdb, model Model) (q UpsertQuery[[]*Model]) {
	q.query.setQuery(db, model, upsert)
	return
}

func (q UpsertQuery[Out]) Values(values Map, params ...ValuesParam) UpsertQuery[Out] {
	params = append(params, OptimizeUpdate)
	q.query.setValues(values, params...)
	return q
}

func (q UpsertQuery[Out]) Conflict(conflict interface{}, params ...ConflictParam) UpsertQuery[Out] {
	q.query.setConflict(conflict, params...)
	return q
}

func (q UpsertQuery[Out]) UpdateValues(values Map, params ...ValuesParam) UpsertQuery[Out] {
	params = append(params, OptimizeUpdate)
	q.query.setUpdateValues(values, params...)
	return q
}

func (q UpsertQuery[Out]) Where(where interface{}) UpsertQuery[Out] {
	q.query.setWhere(where)
	return q
}

func (q UpsertQuery[Out]) Exec() (err error) {
	_, err = q.query.exec()
	return
}
