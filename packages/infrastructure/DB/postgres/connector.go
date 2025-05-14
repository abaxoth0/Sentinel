package postgres

import (
	"context"
	"errors"
	"fmt"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type connector struct {
    ctx         context.Context
    pool        *pgxpool.Pool
    isConnected bool
}

func defaultTimeoutContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), time.Second * 5)
}

func (c *connector) Connect() {
    if c.isConnected {
        dbLogger.Panic("DB connection failed", "connection already established")
    }

    c.ctx = context.Background()

    dbLogger.Info("Creating connection pool...")

    config, err := pgxpool.ParseConfig(fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s",
        config.Secret.DatabaseUser,
        config.Secret.DatabasePassword,
        config.Secret.DatabaseHost,
        config.Secret.DatabasePort,
        config.Secret.DatabaseName,
    ))

    config.MinConns = 10
    config.MaxConns = 50
    config.MaxConnIdleTime = time.Minute * 5
    config.MaxConnLifetime = time.Minute * 60

    if err != nil {
        dbLogger.Fatal("Failed to parse DB connection URI", err.Error())
    }

    pool, err := pgxpool.NewWithConfig(c.ctx, config)

    if err != nil {
        dbLogger.Fatal("Failed to create connection pool", err.Error())
    }

    dbLogger.Info("Creating connection pool: OK")

    dbLogger.Info("Ping connection...")

    ctx, cancel := defaultTimeoutContext()

    defer cancel()

    if err = pool.Ping(ctx); err != nil {
        if err == context.DeadlineExceeded {
            dbLogger.Fatal("Failed to ping DB", "Ping timeout")
        }

        dbLogger.Fatal("Failed to ping DB", err.Error())
    }

    dbLogger.Info("Ping connection: OK")

    c.pool = pool

    err = c.postConnection()

    if err != nil {
        dbLogger.Fatal("Post-connection failed", err.Error())
    }

    c.isConnected = true
}

func (c *connector) Disconnect() error {
    if !c.isConnected {
        return errors.New("connection not established")
    }

    dbLogger.Info("Closing connection pool...")

    done := make(chan bool)

    go func() {
        c.pool.Close()
        close(done)
    }()

    select {
    case <-done:
    case <-time.After(time.Second * 10):
        return errors.New("timeout exceeded")
    }

    dbLogger.Info("Closing connection pool: OK")

    c.isConnected = false

    return nil
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

        dbLogger.Error(
            "Failed to acquire connection from pool",
            err.Error(),
        )

        return nil, Error.StatusInternalError
    }

    return connection, nil
}

func (c *connector) postConnection() error {
    dbLogger.Info("Post-connection...")

    if err := c.initializeTable(
        "user",
        `CREATE TABLE IF NOT EXISTS "user" (
            id uuid PRIMARY KEY,
            login VARCHAR(72) UNIQUE NOT NULL,
            password CHAR(60) NOT NULL,
            roles VARCHAR(32)[] NOT NULL,
            deleted_at TIMESTAMP,
            created_at TIMESTAMP NOT NULL DEFAULT NOW()
        );`,
    ); err != nil {
        return err
    }

    if err := c.initializeTable(
        "audit_user",
        `CREATE TABLE IF NOT EXISTS "audit_user" (
            id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
            changed_user_id uuid REFERENCES "user"(id) ON DELETE CASCADE,
            changed_by_user_id uuid REFERENCES "user"(id) ON DELETE CASCADE,
            operation CHAR(1) NOT NULL,
            login VARCHAR(72) NOT NULL,
            password CHAR(60) NOT NULL,
            roles VARCHAR(32)[] NOT NULL,
            deleted_at TIMESTAMP,
            changed_at TIMESTAMP NOT NULL
        );`,
    ); err != nil {
        return err
    }

    dbLogger.Info("Post-connection: OK")

    return nil
}

func (c *connector) initializeTable(tableName string, query string) error {
    con, err := c.getConnection()

    if err != nil {
        return err
    }

    defer con.Release()

    dbLogger.Info("Verifying that table '"+tableName+"' exists...")

    if _, e := con.Exec(c.ctx, query); e != nil {
        return errors.New("Failed to verify that table '"+tableName+"' exists: "+e.Error())
    }

    dbLogger.Info("Verifying that table '"+tableName+"' exists: OK")

    return nil
}

