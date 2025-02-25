package postgres

import (
	"context"
	"fmt"
	"os"
	Error "sentinel/packages/errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type connector struct {
    ctx context.Context
    pool  *pgxpool.Pool
}

func defaultTimeoutContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), time.Second * 5)
}

func (c *connector) Connect() {
    c.ctx = context.Background()

    fmt.Println("[ DATABASE ] Creating connection pool...")

    config, err := pgxpool.ParseConfig("postgres://postgres:1234@localhost:5432/sentinel")

    config.MinConns = 10
    config.MaxConns = 50
    config.MaxConnIdleTime = time.Minute * 5
    config.MaxConnLifetime = time.Minute * 60

    if err != nil {
        fmt.Printf("Unable to parse DB connection string: %v\n", err.Error())
        os.Exit(1)
    }

    pool, err := pgxpool.NewWithConfig(c.ctx, config)

    if err != nil {
        fmt.Printf("Failed to create connection pool: %v\n", err.Error())
        os.Exit(1)
    }

    fmt.Println("[ DATABASE ] Creating connection pool: OK")

    fmt.Println("[ DATABASE ] Ping connection...")

    ctx, cancel := defaultTimeoutContext()

    defer cancel()

    if err = pool.Ping(ctx); err != nil {
        if err == context.DeadlineExceeded {
            fmt.Printf("[ DATABASE ] Error: Ping timeout")
            os.Exit(1)
        }

        fmt.Printf("[ DATABASE ] Failed to ping: %v\n", err.Error())
        os.Exit(1)
    }

    fmt.Println("[ DATABASE ] Ping connection: OK")

    c.pool = pool
}

func (c *connector) Disconnect() {
    c.pool.Close()
}

// Don't forget to release connection
func (c *connector) getConnection() (*pgxpool.Conn, *Error.Status) {
    ctx, cancel := defaultTimeoutContext()

    defer cancel()

    connection, err := c.pool.Acquire(ctx)

    if err != nil {
        if err == context.DeadlineExceeded {
            return nil, Error.StatusTimeout
        }

        fmt.Printf(
            "[ ERROR ] Failed to acquire connection from pool: %v\n",
            err.Error(),
        )

        return nil, Error.StatusInternalError
    }

    return connection, nil
}

