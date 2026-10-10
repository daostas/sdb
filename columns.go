package sdb

import (
	"context"
	"fmt"
	"strings"
)

func Columns[Model ModelTable](sdb Sdb, model Model) (res []string, err error) {
	return ColumnsByTable(sdb, model.Table())
}

func ColumnsByTable(sdb Sdb, table string) (res []string, err error) {
	if sdb.txErr != nil {
		return nil, sdb.txErr
	}
	if sdb.db == nil {
		return nil, fmt.Errorf("sdb: database is not connected")
	}
	dotIndex := strings.Index(table, ".")
	if dotIndex < 0 {
		return nil, fmt.Errorf("sdb: table must be schema-qualified, got %q", table)
	}
	schema := table[:dotIndex]
	table = table[dotIndex+1:]

	query := fmt.Sprintf("SELECT column_name FROM %s WHERE table_schema='%s' and table_name='%s'", InformationSchemaColumns, schema, table)
	if sdb.log {
		sdb.logger.Info(query)
	}

	rows, queryErr := sdb.db.Query(context.Background(), query)
	if queryErr != nil {
		return nil, queryErr
	}
	err = Scan(rows, &res)
	return
}
