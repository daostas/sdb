package sdb

import (
	"fmt"
	"strings"
)

func Columns[Model ModelTable](sdb Sdb, model Model) (res []string, err error) {
	return ColumnsByTable(sdb, model.Table())
}

func ColumnsByTable(sdb Sdb, table string) (res []string, err error) {
	dotIndex := strings.Index(table, ".")
	schema := table[:dotIndex]
	table = table[dotIndex+1:]

	query := fmt.Sprintf("SELECT column_name FROM %s WHERE table_schema='%s' and table_name='%s'", InformationSchemaColumns, schema, table)
	if sdb.log {
		sdb.logger.Info(query)
	}

	err = sdb.db.Raw(query).Scan(&res).Error
	return
}
