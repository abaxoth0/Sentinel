package search

import (
	"log"
	"net/http"
	"sentinel/packages/DB"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// type user struct {
// 	Email    string
// 	Password string
// 	Role     string
// }

type IndexedUser struct {
	ID       string `bson:"_id"`
	Email    string `bson:"email"`
	Password string `bson:"password"`
	Role     string `bson:"role"`
	// If in DB this property will be nil, then here it will be 0
	DeletedAt int `bson:"deletedAt,omitempty"`
}

type Model struct {
	dbClient   *mongo.Client
	collection *mongo.Collection
}

func New(dbClient *mongo.Client) *Model {
	return &Model{
		dbClient:   dbClient,
		collection: dbClient.Database(config.DB.Name).Collection(config.DB.UserCollectionName),
	}
}

func (m Model) findUserBy(key string, value any) (IndexedUser, error) {
	var user IndexedUser

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	userFilter := bson.D{{key, value}, {"deletedAt", primitive.Null{}}}

	cur, err := m.collection.Find(ctx, userFilter)

	if err != nil {
		log.Fatalln(err)
	}

	if hasResult := cur.Next(ctx); !hasResult {
		return user, ExternalError.New("user not found", http.StatusNotFound)
	}

	err = cur.Decode(&user)

	if err != nil {
		log.Fatalln(err)
	}

	// This actually not a critical problem, cuz on finishing request processing goroutine will be terminated
	// and garbage collector should kill cursor, but idk how it will work in practice.
	if err := cur.Close(ctx); err != nil {
		log.Printf("[ ERROR ] Failed to close cursor. ID: %s, E-Mail:%s\n", user.ID, user.Email)

		// user will be non-empty, but error will still presence
		return user, err
	}

	return user, nil
}

func (m Model) FindUserByID(uid string) (IndexedUser, error) {
	return m.findUserBy("_id", DB.ObjectIDFromHex(uid))
}

func (m Model) FindUserByEmail(email string) (IndexedUser, error) {
	return m.findUserBy("email", email)
}
