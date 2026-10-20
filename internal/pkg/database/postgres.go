package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HieuNguyenVV/sp-agent/internal/pkg/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Postgres holds the master and slave database connections.
type Postgres struct {
	writeDB *sqlx.DB
	readDB  *sqlx.DB
}

type MasterDB *sqlx.DB

type SlaveDB *sqlx.DB

// Conn is the subset of sqlx used by repositories.
type Conn interface {
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
	DriverName() string
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	MustExec(query string, args ...interface{}) sql.Result
	MustExecContext(ctx context.Context, query string, args ...interface{}) sql.Result
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	Preparex(query string) (*sqlx.Stmt, error)
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	Rebind(query string) string
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func NewMasterDB(cfg config.PostgresNodeConfig) (MasterDB, error) {
	sqlDB, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	db := sqlx.NewDb(sqlDB, "postgres")
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func NewSlaveDB(cfg config.PostgresNodeConfig) (SlaveDB, error) {
	sqlDB, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	db := sqlx.NewDb(sqlDB, "postgres")
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func NewPostgres(cfg config.PostgresConfig) (*Postgres, error) {
	masterDB, err := NewMasterDB(cfg.Master)
	if err != nil {
		return nil, err
	}
	slaveDB, err := NewSlaveDB(cfg.Slaves)
	if err != nil {
		return nil, err
	}
	return &Postgres{writeDB: masterDB, readDB: slaveDB}, nil
}

func (p *Postgres) Ping() error {
	if p.writeDB != nil {
		if err := p.writeDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping write database: %w", err)
		}
	}
	if p.readDB != nil {
		if err := p.readDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping read database: %w", err)
		}
	}
	return nil
}

func (p *Postgres) Shutdown() {
	if p.writeDB != nil {
		_ = p.writeDB.Close()
	}
	if p.readDB != nil {
		_ = p.readDB.Close()
	}
}

func (p *Postgres) MasterDB() MasterDB {
	return p.writeDB
}

func (p *Postgres) SlaveDB() SlaveDB {
	return p.readDB
}

func (p *Postgres) GetReadConnection(ctx context.Context) (Conn, error) {
	if transactionCtx, ok := ctx.Value(TransactionCtxKey).(*TransactionCtx); ok {
		if transactionCtx.Conn != nil {
			return transactionCtx.Conn, nil
		}
	}
	if customerSettingCtx, ok := ctx.Value(CustomerSettingCtxKey).(*CustomerSettingCtx); ok {
		if customerSettingCtx.IsJobAfterTxCommit {
			return p.writeDB, nil
		}
	}
	return p.readDB, nil
}

func (p *Postgres) GetWriteConnection(ctx context.Context) (Conn, error) {
	if transactionCtx, ok := ctx.Value(TransactionCtxKey).(*TransactionCtx); ok {
		if transactionCtx.Conn == nil {
			transactionCtx.Mu.Lock()
			defer transactionCtx.Mu.Unlock()

			if transactionCtx.Conn == nil {
				conn, err := p.writeDB.Beginx()
				if err != nil {
					return nil, err
				}
				transactionCtx.Conn = conn
			}
		}
		return transactionCtx.Conn, nil
	}
	return p.writeDB, nil
}
