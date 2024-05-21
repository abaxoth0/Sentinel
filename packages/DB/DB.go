package DB

import (
	"context"
	"log"
	"sentinel/packages/config"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Connect to database. returns pointer to db client and context, used by this connection.
func Connect() (*mongo.Client, context.Context) {
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

	return client, ctx
}
