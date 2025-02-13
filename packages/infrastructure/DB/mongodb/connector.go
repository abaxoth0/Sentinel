package mongodb

import (
	"sentinel/packages/infrastructure/config"
	"strings"

	emongo "github.com/StepanAnanin/EssentialMongoDB"
	"go.mongodb.org/mongo-driver/mongo"
)

type connector struct {
	isConnected           bool
	UserCollection        *mongo.Collection
	DeletedUserCollection *mongo.Collection
}

// Connect to database. Used to initialize public variables in this package (by default they all are nil)
func (DB *connector) Connect() {
	connectionURI := strings.Replace(strings.Replace(config.DB.URI, "<user>", config.DB.Username, 1), "<password>", config.DB.Password, 1)

	emongo.Connect(connectionURI)

	DB.UserCollection = emongo.Client.Database(config.DB.Name).Collection(config.DB.UserCollectionName)
	DB.DeletedUserCollection = emongo.Client.Database(config.DB.Name).Collection(config.DB.DeletedUserCollectionName)

	DB.isConnected = true
}

func (DB *connector) Disconnect() {
	emongo.Client.Disconnect(emongo.Context)
}
