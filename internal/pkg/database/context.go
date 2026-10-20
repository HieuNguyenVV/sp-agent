package database

import (
	"sync"

	"github.com/jmoiron/sqlx"
)

type TransactionCtxKeyType string

type CustomerSettingCtx struct {
	IsJobAfterTxCommit bool
}

type CustomerSettingCtxKeyType string

const (
	TransactionCtxKey     TransactionCtxKeyType     = "transaction_ctx"
	CustomerSettingCtxKey CustomerSettingCtxKeyType = "customer_setting_ctx"
)

type TransactionCtx struct {
	Mu   sync.Mutex
	Conn *sqlx.Tx
}

func (t *TransactionCtx) Commit() error {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	if t.Conn != nil {
		err := t.Conn.Commit()
		if err == nil {
			t.Conn = nil
		}
		return err
	}
	return nil
}

func (t *TransactionCtx) Rollback() error {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	if t.Conn != nil {
		err := t.Conn.Rollback()
		if err == nil {
			t.Conn = nil
		}
		return err
	}
	return nil
}
