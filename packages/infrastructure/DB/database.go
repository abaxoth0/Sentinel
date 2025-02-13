package DB

import (
	"sentinel/packages/core/user"
	"sentinel/packages/infrastructure/DB/mongodb"
)

type database interface {
	connector
	user.Repository
}

type connector interface {
	Connect()
	Disconnect()
}

var Database database = mongodb.InitDriver()
