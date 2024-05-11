package sdb

import (
	"fmt"
	"github.com/daostas/slogger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	InformationSchemaColumns = "information_schema.columns"
)

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
	Log                    bool   `yaml:"log" json:"log"`
	SkipDefaultTransaction bool   `yaml:"skip_default_transaction" json:"skip-default-transaction"`
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
	db.logger = slogger.NewLogger(prefix)
	return
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
