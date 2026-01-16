// This packages is used only for logging things inside of postgres module (/packages/infrastructure/DB/postgres)
package dblog

import "sentinel/packages/common/logger"

var Logger = logger.NewSource("DATABASE", logger.Default)
