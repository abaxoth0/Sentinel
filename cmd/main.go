package main

import (
	"sentinel/cmd/app"
	"sentinel/packages/common/config"
	"sentinel/packages/common/logger"
	"time"
)

var mainLogger = logger.NewSource("MAIN", logger.Default)

// @title 						Sentinel
// @version 					1.0
// @description 				Authentication/Authorization Service
// @BasePath 					/
// @license.name 				AGPL-3.0 (With additional terms, see NOTICE file)
// @license.url					https://www.gnu.org/licenses/agpl-3.0.html

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description 				Bearer token format: Bearer <access_token>

// @securityDefinitions.apikey 	CSRF_Header
// @in 							header
// @name 						X-CSRF-Token
// @description 				CSRF token in request header. Required for state-changing requests, must match with token in cookie

// @securityDefinitions.apikey 	CSRF_Cookie
// @in							cookie
// @name						_csrf
// @description 				CSRF token in cookie. Required for state-changing requests, must match with token in X-CSRF-Token header
func main() {
	app.Args.Parse()

	app.StartInit()

	app.InitDefault()

	logger.Default.Init(config.App.ServiceID)

	if *app.Args.Debug {
		config.Debug.Enabled = true
	}
	if *app.Args.ShowLogs {
		config.App.ShowLogs = true
	}
	if *app.Args.TraceLogs {
		config.App.TraceLogsEnabled = true
	}

	logger.Debug.Store(config.Debug.Enabled)
	logger.Trace.Store(config.App.TraceLogsEnabled)

	app.InitModules()
	app.InitConnections()

    go func () {
        if err := logger.Default.Start(config.Debug.Enabled); err != nil {
            panic(err.Error())
        }
    }()
    defer func() {
        if err := logger.Default.Stop(); err != nil {
			mainLogger.Error("Failed to stop logger", err.Error(), nil)
        }
    }()

    // Reserve some time for logger to start up
    time.Sleep(time.Millisecond * 50)

	r := app.InitRouter()

	app.EndInit()

    app.Start(r)
}

