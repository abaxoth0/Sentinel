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
	ver := "0.9.9.7"

	// Program wasn't run and/or tested on Windows.
	// (Probably it will work, but required minor code modifications)
	if runtime.GOOS != "linux" {
		log.Fatalln("[ CRITICAL ERROR ] OS is not supported. This program can be used only on Linux.")
	}

	DB.Connect()

	defer DB.Client.Disconnect(DB.Context)

	cache.Init()

	log.Println("[ SERVER ] Initializng router...")

	Router := router.Create()

	http.Handle("/", Router)

	log.Println("[ SERVER ] Initializng router: OK")

	util.ClearTerminal()

	fmt.Print(logo)

	fmt.Printf("  Authentication/authorization service (v%s)\n", ver)

	fmt.Printf("  Listening on port: %s\n\n", config.HTTP.Port)

	if config.Debug.Enabled {
		fmt.Printf("[ WARNING ] Debug mode enabled. Some functions may work different and return unexpected values. \n\n")
	}

	if err := http.ListenAndServe(":"+config.HTTP.Port, Router); err != nil {
		log.Println("[ CRITICAL ERROR ] Server error has occurred, the program will stop")

		DB.Client.Disconnect(DB.Context)

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
