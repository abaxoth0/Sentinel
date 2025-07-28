package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sentinel/packages/common/config"
	"sentinel/packages/common/logger"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/infrastructure/email"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

var appLogger = logger.NewSource("APP", logger.Default)

func Start(Router *echo.Echo) {
    stop := make(chan os.Signal, 1)

    signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

    if config.App.IsLoginEmail {
        email.Run()
    }

    go func(){
        err := Router.StartTLS(":" + config.HTTP.Port, "cert.pem", "key.pem")

        appLogger.Info(err.Error(), nil)
    }()

    printAppInfo()

    sig := <-stop

    println()
    appLogger.Info(sig.String() + " signal received, shutting down...", nil)

    ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancel()

    if err := Router.Shutdown(ctx); err != nil {
        appLogger.Error("Failed to stop HTTP server", err.Error(), nil)
    } else {
        appLogger.Info("HTTP server stopped", nil)
    }

	Shutdown()
}

func Shutdown() {
    appLogger.Info("Shutting down...", nil)

    if err := DB.Database.Disconnect(); err != nil {
        appLogger.Error("Failed to disconnect from DB", err.Error(), nil)
    }

    if err := cache.Client.Close(); err != nil {
        appLogger.Error("Failed to disconnect from DB", err.Error(), nil)
    }

    if config.App.IsLoginEmail {
        if err := email.Stop(); err != nil {
            appLogger.Error("Failed to stop mailer", err.Error(), nil)
        }
    }

    appLogger.Info("Shutted down", nil)
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
        appLogger.Warning("Debug mode enabled.", nil)
        print("\n\n")
    }
}

