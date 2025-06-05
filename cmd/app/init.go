package app

import (
	"os"
	"runtime"
	"sentinel/packages/common/config"
	"sentinel/packages/common/logger"
	"sentinel/packages/infrastructure/DB"
	"sentinel/packages/infrastructure/auth/authz"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/presentation/api/http/router"

	"github.com/labstack/echo/v4"
)

func StartInit() {
	// All init logs will be shown anyway
	if err := logger.Default.NewTransmission(logger.Stdout); err != nil {
		panic(err.Error())
	}
}

func EndInit() {
	if !config.App.ShowLogs {
		if err := logger.Default.RemoveTransmission(logger.Stdout); err != nil {
			panic(err.Error())
		}
	}
}

func InitDefault() {
	// Program wasn't tested on OS other than Linux.
	if runtime.GOOS != "linux" {
		println("[ CRITICAL ERROR ] OS is not supported. This program can be used only on Linux-based OS.")
		os.Exit(1)
	}

	config.Init()
}

func InitModules() {
	appLogger.Info("Initializng modules...", nil)

    authz.Init()

	appLogger.Info("Initializng modules: OK", nil)
}

func InitConnections() {
	appLogger.Info("Initializng connections...", nil)

    cache.Client.Connect()
	DB.Database.Connect()

	appLogger.Info("Initializng connections: OK", nil)
}

func InitRouter() *echo.Echo {
	appLogger.Info("Initializng router...", nil)

	Router := router.Create()

	appLogger.Info("Initializng router: OK", nil)

    return Router
}

