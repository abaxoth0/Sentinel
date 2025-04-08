package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sentinel/packages/common/config"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/presentation/api/http/router"
	"syscall"
	"time"
)

func main() {
	// Program wasn't tested on OS other than Linux.
	if runtime.GOOS != "linux" {
		log.Fatalln("[ CRITICAL ERROR ] OS is not supported. This program can be used only on Linux-based OS.")
	}

    config.Init()
    authorization.Init()
    cache.Client.Connect()
	DB.Database.Connect()

	log.Println("[ SERVER ] Initializng router...")

	Router := router.Create()

	log.Println("[ SERVER ] Initializng router: OK")

    stop := make(chan os.Signal, 1)

    signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

    go func(){
        err := Router.Start(":" + config.HTTP.Port)

        log.Printf("[ SERVER ] Stopped: %s\n", err.Error())
    }()

    printAppInfo()

    sig := <-stop

    println()
    log.Printf("[ APP ] '%s' signal received, shutting down...\n", sig.String())

    log.Println("[ SERVER ] Stopping...")

    ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancel()

    if err := Router.Shutdown(ctx); err != nil {
        log.Printf("[ SERVER ] Failed to stop server: %v\n", err)
    } else {
        log.Println("[ SERVER ] Stopping: OK")
    }

    if err := DB.Database.Disconnect(); err != nil {
        log.Printf("[ DATABASE ] Failed to disconnect from DB: %s\n", err.Error())
    }

    if err := cache.Client.Close(); err != nil {
        log.Printf("[ CACHE ] Failed to disconnect from DB: %s\n", err.Error())
    }

    log.Println("[ APP ] Shutted down")
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

