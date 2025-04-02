package main

import (
	"fmt"
	"log"
	"runtime"
	"sentinel/packages/common/config"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/presentation/api/http/router"
)

func main() {
	// Program wasn't tested on OS other than Linux.
	if runtime.GOOS != "linux" {
		log.Fatalln("[ CRITICAL ERROR ] OS is not supported. This program can be used only on Linux-based OS.")
	}

    config.Init()
    authorization.Init()
    cache.Client.Init()
	DB.Database.Connect()
	defer DB.Database.Disconnect()

	log.Println("[ SERVER ] Initializng router...")

	Router := router.Create()

	log.Println("[ SERVER ] Initializng router: OK")

    printAppInfo()

    err := Router.Start(":" + config.HTTP.Port)

    Router.Logger.Fatal(err)
}

func printAppInfo() {
    fmt.Print(`

  ███████╗ ███████╗ ███╗   ██╗ ████████╗ ██╗ ███╗   ██╗ ███████╗ ██╗
  ██╔════╝ ██╔════╝ ████╗  ██║ ╚══██╔══╝ ██║ ████╗  ██║ ██╔════╝ ██║
  ███████╗ █████╗   ██╔██╗ ██║    ██║    ██║ ██╔██╗ ██║ █████╗   ██║
  ╚════██║ ██╔══╝   ██║╚██╗██║    ██║    ██║ ██║╚██╗██║ ██╔══╝   ██║
  ███████║ ███████╗ ██║ ╚████║    ██║    ██║ ██║ ╚████║ ███████╗ ███████╗
  ╚══════╝ ╚══════╝ ╚═╝  ╚═══╝    ╚═╝    ╚═╝ ╚═╝  ╚═══╝ ╚══════╝ ╚══════╝

`)

    fmt.Println("  Authentication/authorization service")

    fmt.Println("  Mady by Stepan Ananin (xrf848@gmail.com)")

    fmt.Printf("  Listening on port: %s\n\n", config.HTTP.Port)

    if config.Debug.Enabled {
        fmt.Printf("[ WARNING ] Debug mode enabled.\n\n")
    }
}

