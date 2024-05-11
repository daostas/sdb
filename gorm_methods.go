package sdb

import (
	"fmt"
	"github.com/daostas/slogger"
	"gorm.io/gorm"
	"strings"
)

type Table struct {
	Name    string
	db      *gorm.DB
	logger  slogger.Logger
	log     bool
	selects []string
	orders  []string
	limit   int
	offset  int
}

func (s Sdb) Table(name string) *Table {
	return &Table{
		db:     s.db,
		logger: s.logger,
		log:    s.log,
		Name:   strings.ToLower(name),
	}
}

// Select Выбор отображаемых полей при загрузке данных
func (t *Table) Select(selects interface{}) *Table {
	switch arg := selects.(type) {
	case []string:
		t.selects = append(t.selects, arg...)
	case string:
		arg = strings.Trim(arg, " ")
		if arg == "" {
			arg = "*"
		}
		t.selects = append(t.selects, arg)
	}
	return t
}

// OrderBy Указываются поля для сортировки
func (t *Table) OrderBy(orders interface{}) *Table {
	switch arg := orders.(type) {
	case []string:
		t.orders = append(t.orders, arg...)
	case string:
		arg = strings.Trim(arg, " ")
		if orders != "" {
			t.orders = append(t.orders, arg)
		}
	}
	return t
}

// Limit Установка лимита возвращаемых записей
func (t *Table) Limit(l int) *Table {
	if l == 0 {
		l = -1
	}
	t.limit = l
	return t
}

// Offset Установка строки с какой возвращать записи
func (t *Table) Offset(o int) *Table {
	t.offset = o
	return t
}

// GetRecord Возвращаем одну запись
func (t *Table) GetRecord(where string, v interface{}) error {
	if t.log {
		sql := t.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Table(t.Name).Select(t.selects).Where(where).Order(strings.Join(t.orders, ",")).Limit(t.limit).Offset(t.offset).Take(v)
		})
		t.logger.Info(sql)
	}
	return t.db.Table(t.Name).Select(t.selects).Where(where).Order(strings.Join(t.orders, ",")).Limit(t.limit).Offset(t.offset).Take(v).Error
}

// GetRecords Возвращаем всю таблицу
func (t *Table) GetRecords(where string, v interface{}) error {
	if t.log {
		sql := t.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Table(t.Name).Select(t.selects).Where(where).Order(strings.Join(t.orders, ",")).Limit(t.limit).Offset(t.offset).Find(v)
		})
		t.logger.Info(sql)
	}

	return t.db.Table(t.Name).Select(t.selects).Where(where).Order(strings.Join(t.orders, ",")).Limit(t.limit).Offset(t.offset).Find(v).Error
}

func (t *Table) Count(where string) (num int, err error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s ", strings.ToLower(t.Name))
	if where != "" {
		query += fmt.Sprintf("WHERE %s", where)
	}
	if t.log {
		t.logger.Info(query)
	}
	err = t.db.Raw(query).Scan(&num).Error
	return
}

// Create Создать запись
func (t *Table) Create(v interface{}) error {
	return t.db.Table(t.Name).Create(v).Error
}

// CreateAndSelect Создать и вернуть последнюю запись
func (t *Table) CreateAndSelect(input, output interface{}) error {
	err := t.Create(input)
	if err != nil {
		return err
	}
	return t.OrderBy("id desc").GetRecord("", output)
}

//// Update Обновить запись, можно указать Select для полей, которые хотим обновить, * - все поля
//func (t *Table) Update(v interface{}, where string) error {
//	return t.db.Table(t.Name).Where(where).Updates(v).Error
//}
//
//// UpdateAndSelect Обновить и вернуть запись по критериям обновления
//func (t *Table) UpdateAndSelect(input, output interface{}, where string) error {
//	err := t.Update(input, where)
//	if err != nil {
//		return err
//	}
//	return t.GetRecord(where, output)
//}

// Delete Удалить запись
//
//	func (t *Table) Delete(v interface{}, where string) error {
//		return t.db.Table(t.Name).Where(where).Delete(v).Error
//	}

// UpdateOrCreate Обновить или создать запись на основе проверки наличия записи
func (t *Table) UpdateOrCreate(v interface{}, where string /*, selects ...interface{}*/) error {
	tx := t.db.Table(t.Name).Where(where).Updates(v)
	if tx.RowsAffected == 0 {
		return t.Create(v)
	}
	return tx.Error
}

//// GetMapRecords Возвращаем всю таблицу
//func (t *Table) GetMapRecords(query interface{}, args ...interface{}) ([]map[string]interface{}, error) {
//	var m []map[string]interface{}
//	sql := t.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
//		return tx.Table(t.Name).Select(t.selects).Where(query, args...).Order(strings.Join(t.orders, ",")).Limit(t.limit).Offset(t.offset).Find(&m)
//	})
//	logger.Info(sql)
//	err := t.db.Table(t.Name).Select(t.selects).Where(query, args...).Order(strings.Join(t.orders, ",")).Limit(t.limit).Offset(t.offset).Find(&m).Error
//	if err == gorm.ErrRecordNotFound {
//		m = nil
//	}
//	return m, err
//}
