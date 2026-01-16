package connection

import (
	"context"
	"errors"
	"fmt"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/util"
	"sentinel/packages/infrastructure/DB/postgres/dblog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Manager struct {
	primaryCtx    context.Context
	PrimaryPool   *pgxpool.Pool
	replicaCtx    context.Context
	ReplicaPool   *pgxpool.Pool
	isConnected   bool
	PrimaryConfig *pgxpool.Config
}

// Type of connection, must be either Primary, either Replica
type Type byte

const (
	Primary Type = 1 << iota
	Replica
)

func newConfig(user, password, host, port, dbName string) *pgxpool.Config {
	dblog.Logger.Trace("Creating connection config...", nil)

	conConfig, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s", user, password, host, port, dbName,
	))
	if err != nil {
		dblog.Logger.Fatal("Failed to parse connection URI", err.Error(), nil)
	}

	conConfig.MinConns = 10
	conConfig.MaxConns = 50
	conConfig.MaxConnIdleTime = time.Minute * 5
	conConfig.MaxConnLifetime = time.Minute * 60

	dblog.Logger.Trace("Creating connection config: OK", nil)

	return conConfig
}

func createConnectionPool(poolName string, conConfig *pgxpool.Config, ctx context.Context) *pgxpool.Pool {
	dblog.Logger.Info("Creating "+poolName+" connection pool...", nil)

	pool, err := pgxpool.NewWithConfig(ctx, conConfig)
	if err != nil {
		dblog.Logger.Fatal("Failed to create "+poolName+" connection pool", err.Error(), nil)
	}

	dblog.Logger.Info("Ping "+poolName+" connection...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	if err = pool.Ping(ctx); err != nil {
		if err == context.DeadlineExceeded {
			dblog.Logger.Fatal("Failed to ping "+poolName+" DB", "Ping timeout", nil)
		}

		dblog.Logger.Fatal("Failed to ping "+poolName+" DB", err.Error(), nil)
	}

	dblog.Logger.Info("Ping "+poolName+" connection: OK", nil)

	dblog.Logger.Info("Creating "+poolName+" connection pool: OK", nil)

	return pool
}

func (m *Manager) IsConnected() bool {
	return m.isConnected
}

func (m *Manager) Connect() error {
	dblog.Logger.Info("Connecting...", nil)

	if m.isConnected {
		errMsg := "connection already established"
		dblog.Logger.Error("Connection failed", errMsg, nil)
		return errors.New(errMsg)
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

	dblog.Logger.Info("Connecting: OK", nil)

	if err := m.postConnection(); err != nil {
		dblog.Logger.Fatal("Post-connection failed", err.Error(), nil)
	}

	m.isConnected = true

	return nil
}

func (m *Manager) Disconnect() error {
	dblog.Logger.Info("Disconnecting...", nil)

	if !m.isConnected {
		errMsg := "connection not established"
		dblog.Logger.Error("Failed to disconnect", errMsg, nil)
		return errors.New(errMsg)
	}

	done := make(chan bool)

	go func() {
		dblog.Logger.Info("Closing primary connection pool...", nil)
		m.PrimaryPool.Close()
		dblog.Logger.Info("Closing primary connection pool: OK", nil)

		dblog.Logger.Info("Closing replica connection pool...", nil)
		m.ReplicaPool.Close()
		dblog.Logger.Info("Closing replica connection pool: OK", nil)

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 10):
		errMsg := "timeout exceeded"
		dblog.Logger.Error("Failed to disconnect", errMsg, nil)
		return errors.New(errMsg)
	}

	dblog.Logger.Info("Disconnecting: OK", nil)

	m.isConnected = false

	return nil
}

// Don't forget to release connection
func (m *Manager) AcquireConnection(conType Type) (*pgxpool.Conn, *Error.Status) {
	dblog.Logger.Trace("Acquiring connection...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var pool *pgxpool.Pool

	switch conType {
	case Primary:
		pool = m.PrimaryPool
	case Replica:
		pool = m.ReplicaPool
	default:
		dblog.Logger.Panic("Failed to acquire connection", "Unknown connection type received", nil)
	}

	connection, err := pool.Acquire(ctx)
	if err != nil {
		if err == context.DeadlineExceeded {
			dblog.Logger.Error("Failed to acquire connection", Error.StatusTimeout.Error(), nil)
			return nil, Error.StatusTimeout
		}

		dblog.Logger.Error("Failed to acquire connection", err.Error(), nil)

		return nil, Error.StatusInternalError
	}

	dblog.Logger.Trace("Acquiring connection: OK", nil)

	return connection, nil
}

func (m *Manager) postConnection() error {
	if config.DB.SkipPostConnection {
		dblog.Logger.Warning("Post-connection skipped", nil)
		return nil
	}

	dblog.Logger.Info("Post-connection...", nil)

	dblog.Logger.Info("Verifying that all tables exists in Primary DB...", nil)

	if err := m.checkTables(Primary); err != nil {
		dblog.Logger.Fatal("Post-connection failed", err.Error(), nil)
	}

	dblog.Logger.Info("Verifying that all tables exists in Primary DB: OK", nil)

	dblog.Logger.Info("Verifying that all tables exists in Replica DB...", nil)

	if err := m.checkTables(Replica); err != nil {
		dblog.Logger.Fatal("Post-connection failed", err.Error(), nil)
	}

	dblog.Logger.Info("Verifying that all tables exists in Replica DB: OK", nil)

	dblog.Logger.Info("Post-connection: OK", nil)

	return nil
}

func (m *Manager) checkTables(conType Type) error {
	con, err := m.AcquireConnection(conType)
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
		name   string
		exists bool
	}

	tables, e := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*table, error) {
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
