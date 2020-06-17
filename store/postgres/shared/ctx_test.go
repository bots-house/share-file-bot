package shared

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxCtx(t *testing.T) {
	ctx := context.TODO()

	assert.Nil(t, GetTx(ctx))

	tx := &sql.Tx{}
	ctx = WithTx(ctx, tx)

	assert.Equal(t, tx, GetTx(ctx))
}
