package DB

import (
	"sentinel/packages/core/location"
	"sentinel/packages/core/session"
	"sentinel/packages/core/user"
	"sentinel/packages/infrastructure/DB/postgres"
)

type database interface {
	connector
	user.Repository
	session.Repository
	location.Repository
}

type connector interface {
	Connect()
	Disconnect() error
}

// Implemets all entities "Repository" interfaces
var Database database = postgres.InitDriver()

type migrate interface {
	Up() 		 error
	Down() 		 error
	Steps(n int) error
}

// Used for applying DB migrations
var Migrate migrate = postgres.Migrate{}

