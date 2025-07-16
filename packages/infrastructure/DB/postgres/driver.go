package postgres

import "sentinel/packages/common/logger"

var dbLogger = logger.NewSource("DB", logger.Default)

type postgers struct {
    connector
    seeker
    repository
	session
	location
}

var driver *postgers

func InitDriver() *postgers {
    driver = new(postgers)

    return driver
}

