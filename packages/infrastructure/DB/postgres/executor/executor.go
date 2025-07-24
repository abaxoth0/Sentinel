package executor

import (
	"context"
	"errors"
	"net"
	"reflect"
	"sentinel/packages/common/config"
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/logger"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var executorLogger = logger.NewSource("EXECUTOR", logger.Default)

var conManager *connection.Manager

func Init(manager *connection.Manager) {
	if manager == nil {
		executorLogger.Panic(
			"Failed to initlized DB executor module",
			"Connetion manager can't be nil",
			nil,
		)
	}
	conManager = manager
}

// Prepares query for execution by acquiring connection and creating contex.
// Also initializes q.free(), which is used to release connection and close context.
//
// Will cause panic if query was already prepared.
func prepare(conType connection.Type, q *query.Query) (*executionContext, context.CancelFunc, *Error.Status) {
    con, err := conManager.GetConnection(conType)
    if err != nil {
        return nil, nil, err
    }

	if config.Debug.Enabled && config.Debug.LogDbQueries {
		args := make([]string, len(q.Args))

		for i, arg := range q.Args {
			switch a := arg.(type) {
			case string:
				args[i] = a
			case []string:
				args[i] = strings.Join(a, ", ")
			case int:
				args[i] = strconv.FormatInt(int64(a), 10)
			case int64:
				args[i] = strconv.FormatInt(a, 10)
			case int32:
				args[i] = strconv.FormatInt(int64(a), 10)
			case float32:
				args[i] = strconv.FormatFloat(float64(a), 'f', 8, 32)
			case float64:
				args[i] = strconv.FormatFloat(float64(a), 'f', 11, 64)
			case time.Time:
				args[i] = a.String()
			case *time.Time:
				args[i] = a.String()
			case bool:
				args[i] = strconv.FormatBool(a)
			case net.IP:
				args[i] = a.To4().String()
			}
		}

		executorLogger.Debug("Running query:\n" + q.SQL + "\n * Query args: " + strings.Join(args, "; "), nil)
	}

	ctx, cancel := newExecutionContext(context.Background(), time.Second * 5, con)

	return ctx, cancel, nil
}

func Rows(conType connection.Type, query *query.Query) (pgx.Rows, *Error.Status) {
	ctx, cancel, err := prepare(conType, query)
	if err != nil {
		return nil, err
	}
	defer cancel()

	r, e := ctx.Connection.Query(ctx, query.SQL, query.Args...)
	if e != nil {
		return nil, query.ConvertError(err)
	}

	return r, nil
}

// Scans a row into the given destinations.
// All dests must be pointers.
// By default, dests validation is disabled,
// to enable this add "debug-safe-db-scans: true" to the config.
// (works only if app launched in debug mode)
type rowScanner = func(dests ...any) *Error.Status

// Wrapper for '*pgxpool.Con.QueryRow'
func Row(conType connection.Type, query *query.Query) (rowScanner, *Error.Status) {
	ctx, cancel, err := prepare(conType, query)
	if err != nil {
		return nil, err
	}
	defer cancel()

	row := ctx.Connection.QueryRow(ctx, query.SQL, query.Args...)

    return func (dests ...any) *Error.Status {
		if config.Debug.Enabled && config.Debug.SafeDatabaseScans {
			for _, dest := range dests {
				typeof := reflect.TypeOf(dest)

				if typeof.Kind() != reflect.Ptr {
					executorLogger.Panic(
						"Query scan failed",
						"Destination for scanning must be a pointer, but got '"+typeof.String()+"'",
						nil,
					)
				}
			}
		}

		if e := row.Scan(dests...); e != nil {
			if errors.Is(e, pgx.ErrNoRows) {
				return Error.StatusNotFound
			}
			return query.ConvertError(e)
		}

		return nil
	}, nil
}

// Wrapper for '*pgxpool.Con.Exec'
func Exec(conType connection.Type, query *query.Query) (*Error.Status) {
	ctx, cancel, err := prepare(conType, query)
	if err != nil {
		return err
	}
	defer cancel()

	if _, err := ctx.Connection.Exec(ctx, query.SQL, query.Args...); err != nil {
		return query.ConvertError(err)
	}

    return nil
}

