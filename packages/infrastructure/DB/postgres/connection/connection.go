package connection

import (
	"context"
	"errors"
	"fmt"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"sentinel/packages/common/util"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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
	if config.DB.SkipPostConnection {
		connectionLogger.Warning("Post-connection skipped", nil)
		return nil
	}

    connectionLogger.Info("Post-connection...", nil)

	connectionLogger.Info("Verifying that all tables exists in Primary DB...", nil)

	if err := m.checkTables(Primary); err != nil {
		connectionLogger.Fatal("Post-connection failed", err.Error(), nil)
	}

	connectionLogger.Info("Verifying that all tables exists in Primary DB: OK", nil)

	connectionLogger.Info("Verifying that all tables exists in Replica DB...", nil)

	if err := m.checkTables(Replica); err != nil {
		connectionLogger.Fatal("Post-connection failed", err.Error(), nil)
	}

	connectionLogger.Info("Verifying that all tables exists in Replica DB: OK", nil)

    connectionLogger.Info("Post-connection: OK", nil)

    return nil
}

func (m *Manager) checkTables(conType Type) error {
    con, err := m.GetConnection(conType)

    if err != nil {
        return err
    }

    defer con.Release()

	sql := `WITH tables_to_check(table_name) AS (VALUES ('user'), ('audit_user'), ('user_session'), ('audit_user_session'), ('location'), ('audit_location'))
	SELECT t.table_name, EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = t.table_name
	) AS table_exists FROM tables_to_check t;`

	ctx := util.Ternary(conType == Primary, m.primaryCtx, m.replicaCtx)

    rows, e := con.Query(ctx, sql)
	if e != nil {
        return e
    }

	type table struct {
		name 	string
		exists  bool
	}

	tables, e := pgx.CollectRows(rows, func (row pgx.CollectableRow) (*table, error) {
		table := new(table)

		if err := row.Scan(&table.name, &table.exists); err != nil {
			return nil, err
		}

		return table, nil
	})
	if e != nil {
		return e
	}

	nonExistingTables := []string{}
	for _, table := range tables {
		if !table.exists {
			nonExistingTables = append(nonExistingTables, table.name)
		}
	}

	if len(nonExistingTables) != 0 {
		return errors.New("ERROR: Following table(-s) does not exists: " + strings.Join(nonExistingTables, ", "))
	}

	return nil
}

