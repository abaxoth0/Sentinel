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

var log = logger.NewSource("APP", logger.Default)

func Start(Router *echo.Echo) {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	if err := email.Init(); err != nil {
		log.Fatal("Failed to start mailer", err.Error(), nil)
	}

	go func() {
		var err error

		if config.HTTP.Secured {
			err = Router.StartTLS(":"+config.HTTP.Port, "cert.pem", "key.pem")
		} else {
			err = Router.Start(":" + config.HTTP.Port)
		}

		log.Info(err.Error(), nil)
	}()

	printAppInfo()

	sig := <-stop

	println()
	log.Info(sig.String()+" signal received, shutting down...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Router.Shutdown(ctx); err != nil {
		log.Error("Failed to stop HTTP server", err.Error(), nil)
	} else {
		log.Info("HTTP server stopped", nil)
	}

	Shutdown()
}

func Shutdown() {
	log.Info("Shutting down...", nil)

	if err := DB.Database.Disconnect(); err != nil {
		log.Error("Failed to disconnect from DB", err.Error(), nil)
	}

	if err := cache.Client.Close(); err != nil {
		log.Error("Failed to disconnect from DB", err.Error(), nil)
	}

	if err := email.Stop(); err != nil {
		log.Error("Failed to stop mailer", err.Error(), nil)
	}

	log.Info("Shutted down", nil)
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
		log.Warning("Debug mode enabled.", nil)
	}

	if !config.HTTP.Secured {
		log.Warning("HTTPS Disabled", nil)
	}

	print("\n\n")
}
