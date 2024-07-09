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
	ID        string    `bson:"_id" json:"_id"`
	Login     string    `bson:"login" json:"login"`
	Password  string    `bson:"password" json:"password"`
	Role      role.Role `bson:"role" json:"role"`
	DeletedAt int       `bson:"deletedAt" json:"deletedAt"`
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

func (m *Model) findUserBy(key string, value string, deleted bool) (*IndexedUser, *ExternalError.Error) {
	var user IndexedUser

	cacheKey := cache.UserKeyPrefix + value

	if rawCachedUser, hit := cache.Get(cacheKey); hit {
		if cachedUser, ok := json.DecodeString[IndexedUser](rawCachedUser); ok {
			return &cachedUser, nil
		}
	}

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	// There is a problem with searching user by ID, it works correctly only with primitive.ObjectID, idk why
	var uid primitive.ObjectID

	isKeyID := key == "_id"

	if isKeyID {
		u, e := primitive.ObjectIDFromHex(value)

		if e != nil {
			return &user, ExternalError.New("Invalid UID", http.StatusBadRequest)
		}

		uid = u
	}

	filter := util.Ternary(isKeyID, bson.D{{key, uid}}, bson.D{{key, value}})

	cur, err := m.collection.Find(ctx, filter)

	if err != nil {
		log.Fatalln(err)
	}

	if hasResult := cur.Next(ctx); !hasResult {
		return &user, ExternalError.New("Пользователь не был найден", http.StatusNotFound)
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

	if deleted && user.DeletedAt == 0 {
		return &IndexedUser{}, ExternalError.New("Пользователь не был найден", http.StatusNotFound)
	}

	if !deleted && user.DeletedAt != 0 {
		return &IndexedUser{}, ExternalError.New("Пользователь не был найден", http.StatusNotFound)
	}

	if rawUser, ok := json.Encode(user); ok {
		cache.Set(cacheKey, rawUser)
	}

	return &user, nil
}

// Search for not deleted user with given UID
func (m *Model) FindUserByID(uid string) (*IndexedUser, *ExternalError.Error) {
	return m.findUserBy("_id", uid, false)
}

// Search for soft deleted user with given UID
func (m *Model) FindSoftDeletedUserByID(uid string) (*IndexedUser, *ExternalError.Error) {
	return m.findUserBy("_id", uid, true)
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
	return m.findUserBy("login", login, false)
}
