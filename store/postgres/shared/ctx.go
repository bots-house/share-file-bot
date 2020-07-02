package shared

import (
	"context"
	"database/sql"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

type ctxKey string

const ctxKeyTx = ctxKey("tx")

func GetTx(ctx context.Context) *sql.Tx {
	v := ctx.Value(ctxKeyTx)
	if v != nil {
		return v.(*sql.Tx)
	}
	return nil
}

func GetExecutorOrDefault(ctx context.Context, d boil.ContextExecutor) boil.ContextExecutor {
	tx := GetTx(ctx)

	if tx != nil {
		return tx
	}

	return d
}

func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, ctxKeyTx, tx)
}
