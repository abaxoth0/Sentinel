package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sentinel/packages/common/config"
	"sentinel/packages/common/logger"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/infrastructure/email"
	"sentinel/packages/presentation/api/http/router"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
    Router := initialize()

    start(Router)
}

var appLogger = logger.NewSource("APP", logger.Default)

func initialize() *echo.Echo {
	// Program wasn't tested on OS other than Linux.
	if runtime.GOOS != "linux" {
		log.Fatalln("[ CRITICAL ERROR ] OS is not supported. This program can be used only on Linux-based OS.")
	}

    // Make logs also appear in terminal
    if err := logger.Default.NewTransmission(logger.Stdout); err != nil {
        log.Fatalln(err.Error())
    }

    // Reserve some time for logger to start up
    time.Sleep(time.Millisecond * 50)

    config.Init()
    authorization.Init()
    cache.Client.Connect()
	DB.Database.Connect()

	appLogger.Info("Initializng router...")

	Router := router.Create()

	appLogger.Info("Initializng router: OK")

    return Router
}

func start(Router *echo.Echo) {
    stop := make(chan os.Signal, 1)

    signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

    go func () {
        if err := logger.Default.Start(config.Debug.Enabled); err != nil {
            panic(err.Error())
        }
    }()
    defer func() {
        if err := logger.Default.Stop(); err != nil {
            entry := logger.NewLogEntry(
                logger.ErrorLogLevel,
                "APP",
                "Failed to stop logger",
                err.Error(),
                )
            logger.Stderr.Log(&entry)
        }
    }()

    // Reserve some time for logger to start up
    time.Sleep(time.Millisecond * 50)

    // Currently email module used only to send activation emails,
    // so there are no point to run/stop it if login isn't email.
    // (cuz in this case activation emails not sends and all users are active by default)
    if config.App.IsLoginEmail {
        email.Run()
    }

    go func(){
        err := Router.Start(":" + config.HTTP.Port)

        appLogger.Info(err.Error())
    }()

    printAppInfo()

    sig := <-stop

    println()
    appLogger.Info(sig.String() + " signal received, shutting down...")

    appLogger.Info("Stopping...")

    ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancel()

    if err := Router.Shutdown(ctx); err != nil {
        appLogger.Error("Failed to stop HTTP server", err.Error())
    } else {
        appLogger.Info("HTTP server stopped")
    }

    if err := DB.Database.Disconnect(); err != nil {
        appLogger.Error("Failed to disconnect from DB", err.Error())
    }

    if err := cache.Client.Close(); err != nil {
        appLogger.Error("Failed to disconnect from DB", err.Error())
    }

    if config.App.IsLoginEmail {
        if err := email.Stop(); err != nil {
            appLogger.Error("Failed to stop mailer", err.Error())
        }
    }

    appLogger.Info("Shutted down")
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
        appLogger.Warning("Debug mode enabled.")
        print("\n\n")
    }
}

