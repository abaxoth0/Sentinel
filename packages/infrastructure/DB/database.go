package DB

import (
	"sentinel/packages/core/user"
	"sentinel/packages/infrastructure/DB/postgres"
)

type database interface {
	connector
	user.Repository
}

type connector interface {
	Connect()
	Disconnect() error
}

// Implemets all entities "Repository" interfaces
var Database database = postgres.InitDriver()

