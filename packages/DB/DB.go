package DB

import (
	"sentinel/packages/config"
	"strings"

	emongo "github.com/StepanAnanin/EssentialMongoDB"
	"go.mongodb.org/mongo-driver/mongo"
)

var isConnected = false

var UserCollection *mongo.Collection
var DeletedUserCollection *mongo.Collection

// Connect to database. Used to initialize public variables in this package (by default they all are nil)
func Connect() {
	connectionURI := strings.Replace(strings.Replace(config.DB.URI, "<user>", config.DB.Username, 1), "<password>", config.DB.Password, 1)

	emongo.Connect(connectionURI)

	UserCollection = emongo.Client.Database(config.DB.Name).Collection(config.DB.UserCollectionName)
	DeletedUserCollection = emongo.Client.Database(config.DB.Name).Collection(config.DB.DeletedUserCollectionName)

	isConnected = true
}

func Disconnect() {
	emongo.Client.Disconnect(emongo.Context)
}
