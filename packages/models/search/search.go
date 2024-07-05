package search

import (
	"log"
	"net/http"
	"sentinel/packages/DB"
	"sentinel/packages/cache"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/json"
	"sentinel/packages/models/role"
	"sentinel/packages/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IndexedUser struct {
	ID       string    `bson:"_id" json:"_id"`
	Login    string    `bson:"login" json:"login"`
	Password string    `bson:"password" json:"password"`
	Role     role.Role `bson:"role" json:"role"`
	// If in DB this property is null, then here it will be 0
	DeletedAt int `bson:"deletedAt,omitempty" json:"deletedAt"`
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

func (m *Model) findUserBy(key string, value string, omitDeleted bool) (*IndexedUser, *ExternalError.Error) {
	var user IndexedUser

	cacheKey := util.Ternary(omitDeleted, cache.SoftDeletedUserKeyPrefix, cache.UserKeyPrefix) + value

	if rawCachedUser, hit := cache.Get(cacheKey); hit {
		if cachedUser, ok := json.DecodeString[IndexedUser](rawCachedUser); ok {
			return &cachedUser, nil
		}
	}

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	userFilter := util.Ternary(omitDeleted, bson.D{{key, value}, {"deletedAt", primitive.Null{}}}, bson.D{{key, value}})

	cur, err := m.collection.Find(ctx, userFilter)

	if err != nil {
		log.Fatalln(err)
	}

	if hasResult := cur.Next(ctx); !hasResult {
		return &user, ExternalError.New("user not found", http.StatusNotFound)
	}

	err = cur.Decode(&user)

	if err != nil {
		log.Fatalln(err)
	}

	// This actually not a critical problem, cuz on finishing request processing goroutine will be terminated
	// and garbage collector should kill cursor, but idk how it will work in practice.
	// user will be non-empty, but error will still presence
	if err := cur.Close(ctx); err != nil {
		log.Printf("[ ERROR ] Failed to close cursor. ID: %s, Login:%s\n", user.ID, user.Login)
	}

	if rawUser, ok := json.Encode(user); ok {
		cache.Set(cacheKey, rawUser)
	}

	return &user, nil
}

// Search for not deleted user with given UID
func (m *Model) FindUserByID(uid string) (*IndexedUser, *ExternalError.Error) {
	return m.findUserBy("_id", uid, true)
}

// Search for soft deleted user with given UID
func (m *Model) FindSoftDeletedUserByID(uid string) (*IndexedUser, *ExternalError.Error) {
	return m.findUserBy("_id", uid, false)
}

// Search for user with given UID, regardless of his deletion status
func (m *Model) FindAnyUserByID(uid string) (*IndexedUser, *ExternalError.Error) {
	out, err := m.FindUserByID(uid)

	if err == nil {
		return out, err
	}

	if err.Status != http.StatusNotFound {
		return nil, err
	}

	return m.FindSoftDeletedUserByID(uid)
}

func (m *Model) FindUserByLogin(login string) (*IndexedUser, *ExternalError.Error) {
	return m.findUserBy("login", login, true)
}
