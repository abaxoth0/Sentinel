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

	if err := parser.Parse(os.Args); err != nil {
		fmt.Println(parser.Usage(err))
	}
}

