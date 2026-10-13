package sdb

import (
	"reflect"
	"testing"
)

type TestStruct struct {
}

func (TestStruct) Table() string { return "test_table" }
func TestT(t *testing.T) {
	_, err := SelectAllMap(Sdb{}, TestStruct{}).Fields("title, count(*) as total").GroupBy("title").Exec()
	if err != nil {
		t.Fatalf("SelectAllMap().GroupBy().Exec() error = %v", err)
	}
}
func TestParseFieldsWithAliases(t *testing.T) {
	fields, fieldsArray := parseFields("title, count(*) as total")

	if want := `"title", COUNT(*) AS "total"`; fields != want {
		t.Fatalf("parseFields() fields = %q, want %q", fields, want)
	}

	wantArray := []string{`"title"`, "COUNT(*)"}
	if !reflect.DeepEqual(fieldsArray, wantArray) {
		t.Fatalf("parseFields() fieldsArray = %#v, want %#v", fieldsArray, wantArray)
	}
}

func TestParseFieldsWithQuotedAlias(t *testing.T) {
	fields, _ := parseFields(`count(*) AS "total"`)

	if want := `COUNT(*) AS "total"`; fields != want {
		t.Fatalf("parseFields() = %q, want %q", fields, want)
	}
}
