package DB

import (
	"sentinel/packages/core/activation"
	"sentinel/packages/core/user"
	"sentinel/packages/infrastructure/DB/postgres"
)

type database interface {
	connector
	user.Repository
    activation.Repository
}

type connector interface {
	Connect()
	Disconnect()
}

// Implemets all entities "Repository" interfaces
var Database database = postgres.InitDriver()

