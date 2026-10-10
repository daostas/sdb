package sdb

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/daostas/slogger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const InformationSchemaColumns = "information_schema.columns"

type DigitConstraints interface {
	int | int8 | int16 | int32 | int64
}

type ModelTable interface {
	Table() string
}

// dbConn is implemented by both pgxpool.Pool and pgx.Tx.
type dbConn interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Sdb struct {
	db        dbConn
	pool      *pgxpool.Pool
	tx        pgx.Tx
	txErr     error
	logger    slogger.Logger
	log       bool
	logInFile bool
}

type Config struct {
	Server                 string `yaml:"server" json:"server"`
	Port                   int    `yaml:"port" json:"port"`
	Username               string `yaml:"username" json:"username"`
	Password               string `yaml:"password" json:"password"`
	Name                   string `yaml:"name" json:"name"`
	Ssl                    bool   `yaml:"ssl" json:"ssl"`
	SkipDefaultTransaction bool   `yaml:"skip_default_transaction" json:"skip-default-transaction"`
	MaxIdleConnections     int    `yaml:"max_idle_connections" json:"max_idle_connections"`
	MaxOpenConnections     int    `yaml:"max_open_connections" json:"max_open_connections"`
	MaxLifeTime            int    `yaml:"max_life_time" json:"max_life_time"`
	MaxIdleLifeTime        int    `yaml:"max_idle_life_time" json:"max_idle_life_time"`
	Log                    bool   `yaml:"log" json:"log"`
	LogInFile              bool   `yaml:"log_in_file" json:"log_in_file"`
}

func Connect(config Config, logger slogger.Logger, loggerPrefix string) (Sdb, error) {
	dsnURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(config.Username, config.Password),
		Host:   net.JoinHostPort(config.Server, strconv.Itoa(config.Port)),
		Path:   "/" + config.Name,
	}
	if !config.Ssl {
		query := dsnURL.Query()
		query.Set("sslmode", "disable")
		dsnURL.RawQuery = query.Encode()
	}

	poolConfig, err := pgxpool.ParseConfig(dsnURL.String())
	if err != nil {
		return Sdb{}, fmt.Errorf("sdb: parse PostgreSQL config: %w", err)
	}
	if config.MaxOpenConnections > 0 {
		poolConfig.MaxConns = int32(config.MaxOpenConnections)
	} else {
		// database/sql interpreted zero as unlimited. pgxpool requires a
		// positive limit, so use the largest supported int32 value.
		poolConfig.MaxConns = int32(^uint32(0) >> 1)
	}
	const effectivelyUnlimited = time.Duration(1<<63 - 1)
	poolConfig.MaxConnLifetime = effectivelyUnlimited
	if config.MaxLifeTime > 0 {
		poolConfig.MaxConnLifetime = time.Duration(config.MaxLifeTime) * time.Millisecond
	}
	poolConfig.MaxConnIdleTime = effectivelyUnlimited
	if config.MaxIdleLifeTime > 0 {
		poolConfig.MaxConnIdleTime = time.Duration(config.MaxIdleLifeTime) * time.Millisecond
	}
	// pgxpool has MinConns but no direct equivalent of database/sql's
	// MaxIdleConns. Keep the pool's default minimum rather than reserving
	// connections just to emulate MaxIdleConnections.
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return Sdb{}, fmt.Errorf("sdb: connect to PostgreSQL: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return Sdb{}, fmt.Errorf("sdb: ping PostgreSQL: %w", err)
	}

	logger.AddToPrefix(loggerPrefix)
	return Sdb{
		db:        pool,
		pool:      pool,
		logger:    logger,
		log:       config.Log,
		logInFile: config.LogInFile,
	}, nil
}

func (sdb *Sdb) Close() error {
	if sdb == nil {
		return nil
	}
	if sdb.tx != nil {
		return fmt.Errorf("sdb: cannot close pool from a transaction handle; commit or rollback the transaction")
	}
	if sdb.pool != nil {
		sdb.pool.Close()
		return nil
	}
	return sdb.txErr
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

// Begin starts a transaction. Errors are retained on the returned Sdb and
// can be checked with Err(); use BeginContext when explicit error handling is preferred.
func (sdb Sdb) Begin() Sdb {
	txDB, err := sdb.BeginContext(context.Background())
	if err != nil {
		txDB.txErr = err
	}
	return txDB
}

// BeginContext starts a transaction and returns any startup error immediately.
func (sdb Sdb) BeginContext(ctx context.Context) (Sdb, error) {
	if sdb.txErr != nil {
		return sdb, sdb.txErr
	}
	if sdb.tx != nil {
		return sdb, fmt.Errorf("sdb: nested transactions are not supported")
	}
	if sdb.db == nil {
		return sdb, fmt.Errorf("sdb: database is not connected")
	}
	tx, err := sdb.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	}).Begin(ctx)
	if err != nil {
		return sdb, fmt.Errorf("sdb: begin transaction: %w", err)
	}
	sdb.db = tx
	sdb.tx = tx
	return sdb, nil
}

func (sdb Sdb) Rollback() Sdb {
	if sdb.tx == nil {
		sdb.txErr = fmt.Errorf("sdb: rollback called outside a transaction")
		return sdb
	}
	if err := sdb.tx.Rollback(context.Background()); err != nil {
		sdb.txErr = fmt.Errorf("sdb: rollback transaction: %w", err)
	}
	sdb.tx = nil
	if sdb.pool != nil {
		sdb.db = sdb.pool
	}
	return sdb
}

func (sdb Sdb) Commit() Sdb {
	if sdb.tx == nil {
		sdb.txErr = fmt.Errorf("sdb: commit called outside a transaction")
		return sdb
	}
	if err := sdb.tx.Commit(context.Background()); err != nil {
		sdb.txErr = fmt.Errorf("sdb: commit transaction: %w", err)
	}
	sdb.tx = nil
	if sdb.pool != nil {
		sdb.db = sdb.pool
	}
	return sdb
}

// Err returns a transaction lifecycle error retained by Begin, Commit or Rollback.
func (sdb Sdb) Err() error {
	return sdb.txErr
}

func CallRoutine(sdb Sdb, funcName string, args ...any) (m map[string]interface{}, err error) {
	query := fmt.Sprintf("SELECT * FROM %s(", funcName)
	for i, arg := range args {
		query += ValueToPostgresValue(arg)
		if i < len(args)-1 {
			query += ", "
		}
	}
	query += ")"

	if sdb.log {
		if sdb.logInFile {
			sdb.logger.InfoFl(query)
		} else {
			sdb.logger.Info(query)
		}
	}
	if sdb.txErr != nil {
		return nil, sdb.txErr
	}
	if sdb.db == nil {
		return nil, fmt.Errorf("sdb: database is not connected")
	}
	rows, err := sdb.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	err = Scan(rows, &m)
	return
}
