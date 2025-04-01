package main

import (
	"fmt"
	"log"
	"runtime"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/common/config"
	"sentinel/packages/presentation/api/http/router"
	"sentinel/packages/common/util"
)

var logo = `

  ███████╗ ███████╗ ███╗   ██╗ ████████╗ ██╗ ███╗   ██╗ ███████╗ ██╗
  ██╔════╝ ██╔════╝ ████╗  ██║ ╚══██╔══╝ ██║ ████╗  ██║ ██╔════╝ ██║
  ███████╗ █████╗   ██╔██╗ ██║    ██║    ██║ ██╔██╗ ██║ █████╗   ██║
  ╚════██║ ██╔══╝   ██║╚██╗██║    ██║    ██║ ██║╚██╗██║ ██╔══╝   ██║
  ███████║ ███████╗ ██║ ╚████║    ██║    ██║ ██║ ╚████║ ███████╗ ███████╗
  ╚══════╝ ╚══════╝ ╚═╝  ╚═══╝    ╚═╝    ╚═╝ ╚═╝  ╚═══╝ ╚══════╝ ╚══════╝

`

func main() {
	// Program wasn't run and/or tested on Windows and MacOS.
	// (Probably it will work, but required minor code modifications)
	if runtime.GOOS != "linux" {
		log.Fatalln("[ CRITICAL ERROR ] OS is not supported. This program can be used only on Linux.")
	}

    config.Init()

    authorization.Init()

    cache.Client.Init()

	DB.Database.Connect()

	defer DB.Database.Disconnect()

	log.Println("[ SERVER ] Initializng router...")

	Router := router.Create()

	log.Println("[ SERVER ] Initializng router: OK")

    if !config.Debug.Enabled {
	    util.ClearTerminal()
    }

	fmt.Print(logo)

	fmt.Println("  Authentication/authorization service")

	fmt.Println("  Mady by Stepan Ananin (xrf848@gmail.com)")

	fmt.Printf("  Listening on port: %s\n\n", config.HTTP.Port)

	if config.Debug.Enabled {
		fmt.Printf("[ WARNING ] Debug mode enabled. Some functions may work different and return unexpected results. \n\n")
	}

    err := Router.Start(":" + config.HTTP.Port)

    Router.Logger.Fatal(err)
}

