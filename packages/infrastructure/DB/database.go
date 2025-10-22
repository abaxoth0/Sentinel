package DB

import (
	"sentinel/packages/core/location"
	"sentinel/packages/core/session"
	"sentinel/packages/core/user"
	"sentinel/packages/infrastructure/DB/postgres"
)

type database interface {
	connector
	user.Manager
	session.Manager
	location.Manager
}

type connector interface {
	Connect() error
	Disconnect() error
}

// Implemets "Manager" interface of each entity
var Database database = postgres.InitDriver()

type migrate interface {
	Up() error
	Down() error
	Steps(n int) error
}

// Used for applying DB migrations
var Migrate migrate = postgres.Migrate{}
