package router

import (
	"os"

	"github.com/labstack/echo/v4"
)

func catchError(next echo.HandlerFunc) echo.HandlerFunc {
	return func (ctx echo.Context) error {
		defer func(){
			if r := recover(); r != nil {
				os.Exit(1)
			}
		}()
		return next(ctx)
	}
}
