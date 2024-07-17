package DB

import (
	"context"
	"log"
	"net/http"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var isConnected = false

var Client *mongo.Client
var Context context.Context

var UserCollection *mongo.Collection
var DeletedUserCollection *mongo.Collection

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
	DeletedUserCollection = Client.Database(config.DB.Name).Collection(config.DB.DeletedUserCollectionName)

	isConnected = true
}

type bsonIndexed struct {
	ID string `bson:"_id"`
}

// `errCallback` (can be nil) will be used if transfer failed,
// and it will be called only on error of inserting `document` back into the `source`.
func CollectionTransfer(document any, source *mongo.Collection, target *mongo.Collection, errCallback func()) *ExternalError.Error {
	indexedDocument, ok := document.(bsonIndexed)

	if !ok {
		return ExternalError.New("Internal Server Error (type assertation failed)", http.StatusInternalServerError)
	}

	ctx, cancel := DefaultTimeoutContext()

	defer cancel()

	if _, err := source.DeleteOne(ctx, bson.D{{"_id", indexedDocument.ID}}); err != nil {
		return ExternalError.New("Transfer failed", http.StatusInternalServerError)
	}

	if _, err := target.InsertOne(ctx, document); err != nil {
		if errCallback != nil {
			errCallback()
		}

		if _, err := source.InsertOne(ctx, document); err != nil {
			log.Fatalln("[ CRITICAL ERROR] Transfer failed, user data lost")
		}

		return ExternalError.New("Transfer failed", http.StatusInternalServerError)
	}

	return nil
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
