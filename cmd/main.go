package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sentinel/packages/DB"
	"sentinel/packages/cache"
	"sentinel/packages/config"
	"sentinel/packages/router"
	"sentinel/packages/util"
)

func main() {
	// Program wasn't run and/or tested on Windows.
	// (Probably it will work, but required minor code modifications)
	if runtime.GOOS != "linux" {
		log.Fatalln("[ CRITICAL ERROR ] OS is not supported. This program can be used only on Linux.")
	}

	dbClient, ctx := DB.Connect()

	defer dbClient.Disconnect(ctx)

	cache.Init()

	log.Println("[ SERVER ] Initializng router...")

	Router := router.Init(dbClient)

	http.Handle("/", Router)

	log.Println("[ SERVER ] Initializng router: OK")

	util.ClearTerminal()

	fmt.Print(logo)

	fmt.Printf("  Authentication/authorization service (v%s)\n\n", config.AppVersion)

	if config.Debug.Enabled {
		fmt.Printf("[ CRITICAL WARNING ] Debug mode enabled! Some functions may work different and return unexpected values. Builded program with enabled debug mode is not intended for production deployment! \n\n")
	}

	// Starting HTTP server
	if err := http.ListenAndServe(":"+config.HTTP.Port, Router); err != nil {
		log.Println("[ CRITICAL ERROR ] Server error has occurred, the program will stop")

		dbClient.Disconnect(ctx)

		panic(err)
	}
}

var logo string = `
  ███████╗ ███████╗ ███╗   ██╗ ████████╗ ██╗ ███╗   ██╗ ███████╗ ██╗     
  ██╔════╝ ██╔════╝ ████╗  ██║ ╚══██╔══╝ ██║ ████╗  ██║ ██╔════╝ ██║     
  ███████╗ █████╗   ██╔██╗ ██║    ██║    ██║ ██╔██╗ ██║ █████╗   ██║     
  ╚════██║ ██╔══╝   ██║╚██╗██║    ██║    ██║ ██║╚██╗██║ ██╔══╝   ██║     
  ███████║ ███████╗ ██║ ╚████║    ██║    ██║ ██║ ╚████║ ███████╗ ███████╗
  ╚══════╝ ╚══════╝ ╚═╝  ╚═══╝    ╚═╝    ╚═╝ ╚═╝  ╚═══╝ ╚══════╝ ╚══════╝
                                                                
`
