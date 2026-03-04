package sdb

import "reflect"

func getModelTable(v any) ModelTable {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	ptr := reflect.New(t).Interface()
	if model, ok := ptr.(ModelTable); ok {
		return model
	} else {
		return nil
	}
}
