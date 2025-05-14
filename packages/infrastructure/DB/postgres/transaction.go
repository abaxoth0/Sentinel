package postgres

import (
	Error "sentinel/packages/common/errors"
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

    tx, err := driver.pool.Begin(ctx)

    if err != nil {
        dbLogger.Error("Failed to begin transaction", err.Error())
        return Error.StatusInternalError
    }

    defer tx.Rollback(ctx)

    for _, query := range t.queries {
        if _, err := tx.Exec(ctx, query.sql, query.args...); err != nil {
            return query.toStatusError(err)
        }
    }

    if err := tx.Commit(ctx); err != nil {
        dbLogger.Error("Failed to commit transaction", err.Error())
        return Error.StatusInternalError
    }

    return nil
}

