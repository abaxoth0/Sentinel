package mongodb

type mongodb struct {
	connector
	seeker
	repository
}

var driver *mongodb

func InitDriver() *mongodb {
	driver = new(mongodb)

	return driver
}
