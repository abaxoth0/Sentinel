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
    primaryCtx    context.Context
    primaryPool   *pgxpool.Pool
	replicaCtx    context.Context
    replicaPool   *pgxpool.Pool
    isConnected   bool
	primaryConfig *pgxpool.Config
}

func defaultTimeoutContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), time.Second * 5)
}

func newConfig(user, password, host, port, dbName string) *pgxpool.Config {
	dbLogger.Trace("Creating connection config...", nil)

    conConfig, err := pgxpool.ParseConfig(fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s", user, password, host, port, dbName,
    ))

    if err != nil {
        dbLogger.Fatal("Failed to parse connection URI", err.Error(), nil)
    }

	conConfig.MinConns = 10
	conConfig.MaxConns = 50
	conConfig.MaxConnIdleTime = time.Minute * 5
	conConfig.MaxConnLifetime = time.Minute * 60

	dbLogger.Trace("Creating connection config: OK", nil)

	return conConfig
}

func createConnectionPool(poolName string, conConfig *pgxpool.Config, ctx context.Context) *pgxpool.Pool {
	dbLogger.Info("Creating "+poolName+" connection pool...", nil)

    pool, err := pgxpool.NewWithConfig(ctx, conConfig)

    if err != nil {
        dbLogger.Fatal("Failed to create "+poolName+" connection pool", err.Error(), nil)
    }

    dbLogger.Info("Ping "+poolName+" connection...", nil)

    ctx, cancel := defaultTimeoutContext()

    defer cancel()

    if err = pool.Ping(ctx); err != nil {
        if err == context.DeadlineExceeded {
            dbLogger.Fatal("Failed to ping "+poolName+" DB", "Ping timeout", nil)
        }

        dbLogger.Fatal("Failed to ping "+poolName+" DB", err.Error(), nil)
    }

    dbLogger.Info("Ping "+poolName+" connection: OK", nil)

	dbLogger.Info("Creating "+poolName+" connection pool: OK", nil)

	return pool
}

func (c *connector) Connect() {
    if c.isConnected {
        dbLogger.Panic("DB connection failed", "connection already established", nil)
    }

    c.primaryCtx = context.Background()

	primaryConnectionConfig := newConfig(
        config.Secret.PrimaryDatabaseUser,
        config.Secret.PrimaryDatabasePassword,
        config.Secret.PrimaryDatabaseHost,
        config.Secret.PrimaryDatabasePort,
        config.Secret.PrimaryDatabaseName,
	)

	replicaConnectionConfig := newConfig(
        config.Secret.ReplicaDatabaseUser,
        config.Secret.ReplicaDatabasePassword,
        config.Secret.ReplicaDatabaseHost,
        config.Secret.ReplicaDatabasePort,
        config.Secret.ReplicaDatabaseName,
	)

	c.primaryCtx = context.Background()
	c.primaryPool = createConnectionPool("primary", primaryConnectionConfig, c.primaryCtx)
	c.primaryConfig = primaryConnectionConfig

	c.replicaCtx = context.Background()
	c.replicaPool = createConnectionPool("replica", replicaConnectionConfig, c.replicaCtx)

	if err := c.postConnection(); err != nil {
        dbLogger.Fatal("Post-connection failed", err.Error(), nil)
    }

    c.isConnected = true
}

func (c *connector) Disconnect() error {
    if !c.isConnected {
        return errors.New("connection not established")
    }

    dbLogger.Info("Closing connection pool...", nil)

    done := make(chan bool)

    go func() {
        c.primaryPool.Close()
        close(done)
    }()

    select {
    case <-done:
    case <-time.After(time.Second * 10):
        return errors.New("timeout exceeded")
    }

    dbLogger.Info("Closing connection pool: OK", nil)

    c.isConnected = false

    return nil
}

type connectionType byte

const (
	primaryConnection connectionType = 1 << iota
	replicaConnection
)

// Don't forget to release connection
func (c *connector) getConnection(conType connectionType) (*pgxpool.Conn, *Error.Status) {
    ctx, cancel := defaultTimeoutContext()

	var pool *pgxpool.Pool

	switch conType {
	case primaryConnection:
		pool = c.primaryPool
	case replicaConnection:
		pool = c.replicaPool
	default:
		dbLogger.Panic(
			"Failed to acquire connection",
			"Unknown connection type received",
			nil,
		)
	}

    defer cancel()

    connection, err := pool.Acquire(ctx)

    if err != nil {
        if err == context.DeadlineExceeded {
            return nil, Error.StatusTimeout
        }

        dbLogger.Error(
            "Failed to acquire connection from pool",
            err.Error(),
			nil,
        )

        return nil, Error.StatusInternalError
    }

    return connection, nil
}

func (c *connector) postConnection() error {
    dbLogger.Info("Post-connection...", nil)

    if err := c.exec(
		primaryConnection,
		"Verifying that table 'user' exists",
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

    if err := c.exec(
		primaryConnection,
		"Verifying that table 'audit_user' exists",
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

    dbLogger.Info("Post-connection: OK", nil)

    return nil
}

func (c *connector) exec(conType connectionType, logBase string, query string) error {
    con, err := c.getConnection(conType)

    if err != nil {
        return err
    }

    defer con.Release()

	dbLogger.Info(logBase + "...", nil)

    if _, e := con.Exec(c.primaryCtx, query); e != nil {
        return errors.New(logBase+": ERROR"+e.Error())
    }

	dbLogger.Info(logBase + ": OK", nil)

	return nil
}

