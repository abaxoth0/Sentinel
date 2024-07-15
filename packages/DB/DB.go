package DB

import (
	"context"
	"log"
	"sentinel/packages/config"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var isConnected = false

var Client *mongo.Client
var Context context.Context

var UserCollection *mongo.Collection

// Connect to database. Used to initialize public variables in this package (by default they all are nil)
func Connect() {
	if isConnected {
		log.Fatalln("[ DATABASE ] Critical error: connection already established")
	}

	ctx := context.Background()

	log.Print("[ DATABASE ] Connecting...")

	connectionURI := strings.Replace(strings.Replace(config.DB.URI, "<user>", config.DB.Username, 1), "<password>", config.DB.Password, 1)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionURI))

	if err != nil {
		panic(err)
	}

	log.Print("[ DATABASE ] Connecting: OK")

	log.Print("[ DATABASE ] Checking connection...")

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalln("[ CRITICAL ERROR ] Failed to ensure DB connection. Error:\n" + err.Error())
	}

	log.Print("[ DATABASE ] Checking connection: OK")

	Client = client
	Context = ctx
	UserCollection = Client.Database(config.DB.Name).Collection(config.DB.UserCollectionName)

	isConnected = true
}

func ObjectIDFromHex(hex string) primitive.ObjectID {
	objectID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		log.Fatal(err)
	}
	return objectID
}

func DefaultTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.TODO(), config.DB.QueryDefaultTimeout)
}
