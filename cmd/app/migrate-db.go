package app

import (
	"os"
	"sentinel/packages/infrastructure/DB"
)

func MigrateDB(migrateSteps string) {
	var err error
	valid := false

	if migrateSteps == "Up"{
		valid = true
		err = DB.Migrate.Up()
	}
	if migrateSteps == "Down" {
		valid = true
		err = DB.Migrate.Down()
	}

	Shutdown()

	if !valid {
		println("Invalid 'migrate-db' argument value. Expected: number or 'Up' or 'Down'. Got: " + migrateSteps)
		os.Exit(1)
	}
	if err != nil {
		println("Failed to apply migration.\n" + err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

