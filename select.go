package sdb

func selectAllFrom[Model ModelTable, digit DigitConstraints](sdb Sdb, model Model, fields, orders interface{}, where interface{}, limit, offset digit) (out []*Model, err error) {
	field := parseFields(fields)
	order := parseOrders(orders)
	_where, err := parseWhere(sdb, model, where)
	if err != nil {
		return
	}

	//Достаем необходимые данные
	err = sdb.Table(model.Table()).Select(field).OrderBy(order).Limit(int(limit)).Offset(int(offset)).GetRecords(_where, &out)
	return
}

func SelectAllFrom[Model ModelTable, digit DigitConstraints](sdb Sdb, model Model, fields, orders interface{}, where interface{}, limit, offset digit) (out []*Model, err error) {
	return selectAllFrom(sdb, model, fields, orders, where, limit, offset)
}

func selectOneFrom[Model ModelTable](sdb Sdb, model Model, fields interface{}, where interface{}) (out Model, err error) {
	field := parseFields(fields)
	_where, err := parseWhere(sdb, model, where)
	if err != nil {
		return
	}
	err = sdb.Table(model.Table()).Select(field).GetRecord(_where, &out)
	return
}

func SelectOneFrom[Model ModelTable](sdb Sdb, model Model, fields interface{}, where interface{}) (out Model, err error) {
	return selectOneFrom(sdb, model, fields, where)
}
