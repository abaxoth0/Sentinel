package app

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
)

type appArgs struct {
	Debug     *bool
	ShowLogs  *bool
	TraceLogs *bool
	MigrateDB *string
}

var Args = new(appArgs)

func (a *appArgs) Parse() {
	parser := argparse.NewParser(
		"Sentinel",
		"Authentication/authorization service. Made by Stepan Ananin (xrf848@gmail.com)",
	)

	Args.Debug = parser.Flag("d", "debug", &argparse.Options{
		Help: "Enable debug mode",
	})
	Args.ShowLogs = parser.Flag("l", "show-logs", &argparse.Options{
		Help: "Show logs in terminal",
	})
	Args.TraceLogs = parser.Flag("t", "trace-logs", &argparse.Options{
		Help: "Enable trace logs",
	})
	Args.MigrateDB = parser.String("M", "migrate-db", &argparse.Options{
		Help: "Apply DB migrations, valid values:\n"+
			  "\t\t\tUp - Migrate forward on 1 version\n"+
			  "\t\t\tDown - Migrate back on 1 version\n"+
			  "\t\t\tN - Number, if N > 0 then will migrate forward on N versions, if N < 0 then will migrate back on N versions",
	})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Println(parser.Usage(err))
	}
}

