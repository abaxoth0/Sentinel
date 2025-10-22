package executor

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type executionContext struct {
	Connection *pgxpool.Conn

	context.Context
}

func newExecutionContext(
	parent context.Context,
	timeout time.Duration,
	con *pgxpool.Conn,
) (*executionContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)

	execCtx := &executionContext{
		Connection: con,
		Context:    ctx,
	}

	return execCtx, func() {
		cancel()
		con.Release()
	}
}
