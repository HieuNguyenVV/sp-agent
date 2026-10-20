package tx

import (
	"context"

	"github.com/HieuNguyenVV/sp-agent/internal/log"
	"github.com/HieuNguyenVV/sp-agent/internal/pkg/database"
)

func WithTransaction(ctx context.Context, logger log.Logger, fu func(ctx context.Context) error) error {
	trxCtx := new(database.TransactionCtx)
	ctx = context.WithValue(ctx, database.TransactionCtxKey, trxCtx)
	err := fu(ctx)
	if err != nil {
		if rollbackErr := trxCtx.Rollback(); rollbackErr != nil {
			logger.Errorf("failed to rollback transaction: %v", rollbackErr)
		}
		return err
	}
	if err := trxCtx.Commit(); err != nil {
		if rollbackErr := trxCtx.Rollback(); rollbackErr != nil {
			logger.Errorf("failed to rollback transaction: %v", rollbackErr)
		}
		return err
	}
	return nil
}
