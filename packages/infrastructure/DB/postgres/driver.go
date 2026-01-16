package postgres

import (
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/executor"
	LocationTable "sentinel/packages/infrastructure/DB/postgres/table/location"
	SessionTable "sentinel/packages/infrastructure/DB/postgres/table/session"
	UserTable "sentinel/packages/infrastructure/DB/postgres/table/user"
	"sentinel/packages/infrastructure/DB/postgres/transaction"
)

type (
	ConnectionManager = *connection.Manager
	UserManager       = *UserTable.Manager
	SessionManager    = *SessionTable.Manager
	LocationManager   = *LocationTable.Manager
)

type postgers struct {
	ConnectionManager
	UserManager
	SessionManager
	LocationManager
}

var driver *postgers

func InitDriver() *postgers {
	session := new(SessionTable.Manager)
	location := new(LocationTable.Manager)
	connection := new(connection.Manager)

	user := UserTable.NewManager(session)

	driver = &postgers{
		ConnectionManager: ConnectionManager(connection),
		UserManager:       UserManager(user),
		SessionManager:    SessionManager(session),
		LocationManager:   LocationManager(location),
	}

	executor.Init(connection)
	transaction.Init(connection)

	return driver
}
