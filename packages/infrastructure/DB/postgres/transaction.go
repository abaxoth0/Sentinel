package postgres

import (
	"errors"
	Error "sentinel/packages/common/errors"

	"github.com/jackc/pgx/v5"
)

type transaction struct {
    queries []*query
}

func newTransaction(queries ...*query) *transaction {
    return &transaction{queries}
}

func (t *transaction) Exec() *Error.Status {
    ctx, cancel := defaultTimeoutContext()
    defer cancel()

    tx, err := driver.primaryPool.Begin(ctx)

    if err != nil {
        dbLogger.Error("Failed to begin transaction", err.Error(), nil)
        return Error.StatusInternalError
    }

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			dbLogger.Error("Rollback failed (non-critical)", err.Error(), nil)
		}
	}()

    for _, query := range t.queries {
        if _, err := tx.Exec(ctx, query.sql, query.args...); err != nil {
            return convertQueryError(err, query.sql)
        }
    }

    if err := tx.Commit(ctx); err != nil {
        dbLogger.Error("Failed to commit transaction", err.Error(), nil)
        return Error.StatusInternalError
    }

    return nil
}

