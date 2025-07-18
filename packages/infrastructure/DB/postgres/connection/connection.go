package connection

import (
	"context"
	"errors"
	"fmt"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var connectionLogger = logger.NewSource("CONNECTION", logger.Default)

type Manager struct {
    primaryCtx    context.Context
    PrimaryPool   *pgxpool.Pool
	replicaCtx    context.Context
    ReplicaPool   *pgxpool.Pool
    isConnected   bool
	PrimaryConfig *pgxpool.Config
}

type Type byte

const (
	Primary Type = 1 << iota
	Replica
)

func newConfig(user, password, host, port, dbName string) *pgxpool.Config {
	connectionLogger.Trace("Creating connection config...", nil)

    conConfig, err := pgxpool.ParseConfig(fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s", user, password, host, port, dbName,
    ))

    if err != nil {
        connectionLogger.Fatal("Failed to parse connection URI", err.Error(), nil)
    }

	conConfig.MinConns = 10
	conConfig.MaxConns = 50
	conConfig.MaxConnIdleTime = time.Minute * 5
	conConfig.MaxConnLifetime = time.Minute * 60

	connectionLogger.Trace("Creating connection config: OK", nil)

	return conConfig
}

func createConnectionPool(poolName string, conConfig *pgxpool.Config, ctx context.Context) *pgxpool.Pool {
	connectionLogger.Info("Creating "+poolName+" connection pool...", nil)

    pool, err := pgxpool.NewWithConfig(ctx, conConfig)

    if err != nil {
        connectionLogger.Fatal("Failed to create "+poolName+" connection pool", err.Error(), nil)
    }

    connectionLogger.Info("Ping "+poolName+" connection...", nil)

    ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)

    defer cancel()

    if err = pool.Ping(ctx); err != nil {
        if err == context.DeadlineExceeded {
            connectionLogger.Fatal("Failed to ping "+poolName+" DB", "Ping timeout", nil)
        }

        connectionLogger.Fatal("Failed to ping "+poolName+" DB", err.Error(), nil)
    }

    connectionLogger.Info("Ping "+poolName+" connection: OK", nil)

	connectionLogger.Info("Creating "+poolName+" connection pool: OK", nil)

	return pool
}

func (m *Manager) IsConnected() bool {
	return m.isConnected
}

func (m *Manager) Connect() {
    if m.isConnected {
        connectionLogger.Panic("DB connection failed", "connection already established", nil)
    }

    m.primaryCtx = context.Background()

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

	m.primaryCtx = context.Background()
	m.PrimaryPool = createConnectionPool("primary", primaryConnectionConfig, m.primaryCtx)
	m.PrimaryConfig = primaryConnectionConfig

	m.replicaCtx = context.Background()
	m.ReplicaPool = createConnectionPool("replica", replicaConnectionConfig, m.replicaCtx)

	if err := m.postConnection(); err != nil {
        connectionLogger.Fatal("Post-connection failed", err.Error(), nil)
    }

    m.isConnected = true
}

func (m *Manager) Disconnect() error {
    if !m.isConnected {
        return errors.New("connection not established")
    }

    connectionLogger.Info("Closing connection pool...", nil)

    done := make(chan bool)

    go func() {
        m.PrimaryPool.Close()
        close(done)
    }()

    select {
    case <-done:
    case <-time.After(time.Second * 10):
        return errors.New("timeout exceeded")
    }

    connectionLogger.Info("Closing connection pool: OK", nil)

    m.isConnected = false

    return nil
}

type Getter = func (conType Type) (*pgxpool.Conn, *Error.Status)

// Don't forget to release connection
func (m *Manager) GetConnection(conType Type) (*pgxpool.Conn, *Error.Status) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)

	var pool *pgxpool.Pool

	switch conType {
	case Primary:
		pool = m.PrimaryPool
	case Replica:
		pool = m.ReplicaPool
	default:
		connectionLogger.Panic(
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

        connectionLogger.Error(
            "Failed to acquire connection from pool",
            err.Error(),
			nil,
        )

        return nil, Error.StatusInternalError
    }

    return connection, nil
}

func (m *Manager) postConnection() error {
    connectionLogger.Info("Post-connection...", nil)

    if err := m.exec(
		Primary,
		"Verifying that table 'user' exists",
        `CREATE TABLE IF NOT EXISTS "user" (
            id uuid PRIMARY KEY,
            login VARCHAR(72) UNIQUE NOT NULL,
            password CHAR(60) NOT NULL,
            roles VARCHAR(32)[] NOT NULL,
            deleted_at TIMESTAMP,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			version INT DEFAULT 1
        );`,
    ); err != nil {
        return err
    }

    if err := m.exec(
		Primary,
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
            changed_at TIMESTAMP NOT NULL,
			version INT DEFAULT 1
        );`,
    ); err != nil {
        return err
    }

    if err := m.exec(
		Primary,
		"Verifying that table 'user_session' exists",
		`CREATE TABLE IF NOT EXISTS "user_session" (
			id                  UUID PRIMARY KEY,
			user_id             UUID REFERENCES "user"(id) ON DELETE CASCADE,
			user_agent          TEXT NOT NULL,
			ip_address          INET,
			device_id           TEXT,
			device_type         TEXT NOT NULL,
			os                  TEXT NOT NULL,
			os_version          TEXT,
			browser             TEXT NOT NULL,
			browser_version     TEXT,
			location            TEXT,
			created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
			last_used_at        TIMESTAMP,
			expires_at          TIMESTAMP NOT NULL,
			revoked             BOOL NOT NULL DEFAULT FALSE
    	);`,
    ); err != nil {
        return err
    }
    if err := m.exec(
		Primary,
		"Verifying that table 'location' exists",
		`CREATE TABLE IF NOT EXISTS location (
 	       id          UUID PRIMARY KEY,
 	       ip          INET NOT NULL,
 	       session_id  UUID REFERENCES "user_session"(id) ON DELETE SET NULL,
 	       country     VARCHAR(2) NOT NULL,
 	       region      VARCHAR(3),
 	       city        VARCHAR(100),
 	       latitude    REAL,
 	       longitude   REAL,
 	       isp         VARCHAR(100),
 	       deleted_at  TIMESTAMPTZ,
 	       created_at  TIMESTAMPTZ DEFAULT NOW() NOT NULL
 	   );`,
	); err != nil {
	return err
	}
    connectionLogger.Info("Post-connection: OK", nil)

    return nil
}

func (m *Manager) exec(conType Type, logBase string, query string) error {
    con, err := m.GetConnection(conType)

    if err != nil {
        return err
    }

    defer con.Release()

	connectionLogger.Info(logBase + "...", nil)

    if _, e := con.Exec(m.primaryCtx, query); e != nil {
        return errors.New(logBase+": ERROR"+e.Error())
    }

	connectionLogger.Info(logBase + ": OK", nil)

	return nil
}

