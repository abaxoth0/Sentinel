// This packages is used only for logging things inside of postgres module (/packages/infrastructure/DB/postgres)
package log

import "sentinel/packages/common/logger"

var (
	DB 		  = logger.NewSource("DATABASE", logger.Default)
	Migration = logger.NewSource("MIGRATION", logger.Default)
)

