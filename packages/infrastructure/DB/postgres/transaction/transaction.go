package transaction

import (
	"context"
	"errors"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	log "sentinel/packages/infrastructure/DB/postgres/logger"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"time"

	"github.com/jackc/pgx/v5"
)

var conManager *connection.Manager

func Init(manager *connection.Manager) {
	if manager == nil {
		log.DB.Panic(
			"Failed to initlized DB transaction module",
			"Connetion manager can't be nil",
			nil,
		)
	}
	conManager = manager
}

type Transaction struct {
    queries []*query.Query
}

func New(queries ...*query.Query) *Transaction {
    return &Transaction{queries}
}

func (t *Transaction) Exec(conType connection.Type) *Error.Status {
	log.DB.Trace("Executing transaction...", nil)

	if len(t.queries) == 0 {
		log.DB.Warning("Transaction has no queries, execution will be skipped", nil)
		return nil
	}

	for _, query := range t.queries {
		if query == nil {
			log.DB.Panic("Failed to run transaction", "At least one query is nil", nil)
			return Error.StatusInternalError
		}
	}

    ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
	defer cancel()

	var tx pgx.Tx
	var err error
	switch conType {
	case connection.Primary:
		tx, err = conManager.PrimaryPool.Begin(ctx)
	case connection.Replica:
		tx, err = conManager.ReplicaPool.Begin(ctx)
	default:
		log.DB.Panic(
			"Failed to run DB transaction",
			"Unknown connection type",
			nil,
		)
	}

    if err != nil {
        log.DB.Error("Failed to begin transaction", err.Error(), nil)
        return Error.StatusInternalError
    }

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.DB.Error("Rollback failed (non-critical)", err.Error(), nil)
		}
	}()

    for _, query := range t.queries {
        if _, err := tx.Exec(ctx, query.SQL, query.Args...); err != nil {
			log.DB.Error("Transaction failed", err.Error(), nil)
            return query.ConvertAndLogError(err)
		}
    }

    if err := tx.Commit(ctx); err != nil {
        log.DB.Error("Failed to commit transaction", err.Error(), nil)
        return Error.StatusInternalError
    }

	log.DB.Trace("Executing transaction: OK", nil)

    return nil
}

