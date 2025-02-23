package sdb

import (
	"fmt"
	"github.com/daostas/slogger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
	"time"
)

const (
	InformationSchemaColumns = "information_schema.columns"
)

type ModelTable interface {
	Table() string
}

type Sdb struct {
	db     *gorm.DB
	logger slogger.Logger
	log    bool
}

type DbConfig struct {
	Server                 string `yaml:"server" json:"server"`
	Port                   int    `yaml:"port" json:"port"`
	Username               string `yaml:"username" json:"username"`
	Password               string `yaml:"password" json:"password"`
	Name                   string `yaml:"name" json:"name"`
	Ssl                    bool   `yaml:"ssl" json:"ssl"`
	SkipDefaultTransaction bool   `yaml:"skip_default_transaction" json:"skip-default-transaction"`
	MaxIdleConnections     int    `yaml:"max_idle_connections"`
	MaxOpenConnections     int    `yaml:"max_open_connections"`
	MaxLifeTime            int    `yaml:"max_life_time"`
	MaxIdleLifeTime        int    `yaml:"max_idle_life_time"`
	Log                    bool   `yaml:"log" json:"log"`
}

func ConnectDb(config DbConfig, prefix string) (db Sdb, err error) {

	dataSource := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d",
		config.Server,
		config.Username,
		config.Password,
		config.Name,
		config.Port,
	)

	if !config.Ssl {
		dataSource += " sslmode=disable"
	}

	db.log = config.Log
	db.db, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dataSource,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		SkipDefaultTransaction: config.SkipDefaultTransaction,
	})
	if err != nil {
		return
	}

	sqlDb, err := db.db.DB()
	if err != nil {
		return
	}

	sqlDb.SetMaxIdleConns(config.MaxIdleConnections)
	sqlDb.SetMaxOpenConns(config.MaxOpenConnections)
	sqlDb.SetConnMaxLifetime(time.Duration(config.MaxLifeTime) * time.Millisecond)
	sqlDb.SetConnMaxIdleTime(time.Duration(config.MaxIdleLifeTime) * time.Millisecond)

	db.logger = slogger.NewLogger(prefix)
	return
}

func (sdb *Sdb) Close() error {
	if db, err := sdb.db.DB(); err == nil {
		return db.Close()
	} else {
		return err
	}
}

func (s *Sdb) Set(log bool, logger ...slogger.Logger) *Sdb {
	if len(logger) >= 1 {
		s.logger = logger[0]
	}
	s.log = log || s.log
	return s
}

func (s Sdb) Copy(log bool, logger ...slogger.Logger) (copy Sdb) {
	copy = s
	if len(logger) >= 1 {
		copy.logger = logger[0]
	}
	copy.log = log || copy.log
	return
}

func (sdb Sdb) Begin() Sdb {
	return Sdb{
		db:     sdb.db.Begin(),
		logger: sdb.logger,
		log:    sdb.log,
	}
}

func (sdb Sdb) Rollback() Sdb {
	return Sdb{
		db:     sdb.db.Rollback(),
		logger: sdb.logger,
		log:    sdb.log,
	}
}

func (sdb Sdb) Commit() Sdb {
	return Sdb{
		db:     sdb.db.Commit(),
		logger: sdb.logger,
		log:    sdb.log,
	}
}

func (sdb Sdb) Scan(db *gorm.DB, out interface{}) error {
	rows, err := db.Rows()
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		err = db.ScanRows(rows, &out)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}

func CallRoutine(sdb Sdb, funcName string, args ...any) (m map[string]interface{}, err error) {
	query := fmt.Sprintf("SELECT * FROM %s(", funcName)
	//Формируем аргументы из args
	for i, arg := range args {
		query += ValueToPostgresValue(arg)
		if i < len(args)-1 {
			query += ", "
		}
	}
	query += ")"

	// Логирование и выполнение функции
	if sdb.log {
		sdb.logger.Info(query)
	}

	err = sdb.db.Raw(query).Scan(&m).Error
	return
}

func Columns[Model ModelTable](sdb Sdb, model Model) (res []string, err error) {
	tableName := model.Table()
	dotIndex := strings.Index(tableName, ".")
	schema := tableName[:dotIndex]
	tableName = tableName[dotIndex+1:]

	query := fmt.Sprintf("SELECT column_name FROM %s WHERE table_schema='%s' and table_name='%s'", InformationSchemaColumns, schema, tableName)
	if sdb.log {
		sdb.logger.Info(query)
	}

	err = sdb.db.Raw(query).Scan(&res).Error
	return
}
