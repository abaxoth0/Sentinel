package main

import (
	"fmt"
	"os"
	"sentinel/cmd/app"
	"sentinel/packages/common/config"
	"sentinel/packages/common/logger"
	"sentinel/packages/infrastructure/DB"
	"strconv"
	"time"

	"github.com/akamensky/argparse"
)

var migrateLogger = logger.NewSource("MIGRATE", logger.Default)

var args = new(migrateArgs)

func main() {
	args.Parse()

	app.StartInit()

	app.InitDefault()

	logger.Default.Init(config.App.ServiceID)

	if *args.Debug {
		config.Debug.Enabled = true
	}
	if *args.ShowLogs {
		config.App.ShowLogs = true
	}
	if *args.TraceLogs {
		config.App.TraceLogsEnabled = true
	}

	logger.Debug.Store(config.Debug.Enabled)
	logger.Trace.Store(config.App.TraceLogsEnabled)

	DB.Database.Connect()

	go func() {
		if err := logger.Default.Start(config.Debug.Enabled); err != nil {
			panic(err.Error())
		}
	}()
	defer func() {
		if err := logger.Default.Stop(); err != nil {
			migrateLogger.Error("Failed to stop logger", err.Error(), nil)
		}
	}()

	// Reserve some time for logger to start up
	time.Sleep(time.Millisecond * 50)

	app.EndInit()

	migrateDB(*args.Steps)
}

func migrateDB(steps string) {
	var err error

	switch steps {
	case "Up", "up":
		err = DB.Migrate.Up()
	case "Down", "down":
		err = DB.Migrate.Down()
	default:
		n, e := strconv.Atoi(steps)
		if e != nil {
			println("Invalid 'migrate-db' argument value. Expected: number or 'Up' or 'Down'. Got: " + steps)
			os.Exit(1)
		}

		err = DB.Migrate.Steps(n)
	}

	if err != nil {
		println("Failed to apply migration.\n" + err.Error())
		os.Exit(1)
	}

	if err := DB.Database.Disconnect(); err != nil {
		migrateLogger.Error("Failed to disconnect from DB", err.Error(), nil)
	}
}

type migrateArgs struct {
	Debug     *bool
	ShowLogs  *bool
	TraceLogs *bool
	Steps     *string
}

func (a *migrateArgs) Parse() {
	parser := argparse.NewParser("sentinel-migrate", "Application for applying database migrations to Sentinel DB")

	args.Debug = parser.Flag("d", "debug", &argparse.Options{
		Help: "Enable debug mode",
	})
	args.ShowLogs = parser.Flag("l", "show-logs", &argparse.Options{
		Help: "Show logs in terminal",
	})
	args.TraceLogs = parser.Flag("t", "trace-logs", &argparse.Options{
		Help: "Enable trace logs",
	})
	args.Steps = parser.String("s", "steps", &argparse.Options{
		Required: true,
		Help: "(Required) Amount of database migration steps. Valid values:\n" +
			"\t\t\t- Up: Migrate forward on 1 version\n" +
			"\t\t\t- Down: Migrate back on 1 version\n" +
			"\t\t\t- N: Number, if N > 0 then will migrate forward on N versions, if N < 0 then will migrate back on N versions",
	})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Println(parser.Usage(err))
		os.Exit(1)
	}
}
