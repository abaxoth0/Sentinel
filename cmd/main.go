package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/infrastructure/config"
	"sentinel/packages/presentation/api/http/router"
	"sentinel/packages/util"

	"github.com/labstack/echo/v4/middleware"
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
	ver := "1.2.0.0"

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

    cors := middleware.CORSConfig{
        Skipper:      middleware.DefaultSkipper,
        AllowOrigins: config.HTTP.AllowedOrigins,
        AllowCredentials: true,
        AllowMethods: []string{
            http.MethodGet,
            http.MethodHead,
            http.MethodPut,
            http.MethodPatch,
            http.MethodPost,
            http.MethodDelete,
        },
    }

    Router.Use(middleware.CORSWithConfig(cors))
    Router.Use(middleware.Recover())
    // Router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10_000)))

    if config.Debug.Enabled {
        Router.Use(middleware.Logger())
    }

	log.Println("[ SERVER ] Initializng router: OK")

    if !config.Debug.Enabled {
	    util.ClearTerminal()
    }

	fmt.Print(logo)

	fmt.Printf("  Authentication/authorization service (v%s)\n", ver)

	fmt.Println("  Mady by Stepan Ananin (xrf844@gmail.com)")

	fmt.Printf("  Listening on port: %s\n\n", config.HTTP.Port)

	if config.Debug.Enabled {
		fmt.Printf("[ WARNING ] Debug mode enabled. Some functions may work different and return unexpected results. \n\n")
	}

    err := Router.Start(":" + config.HTTP.Port)

    Router.Logger.Fatal(err)
}

