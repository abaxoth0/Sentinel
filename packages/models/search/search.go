package search

import (
	"log"
	"net/http"
	"sentinel/packages/DB"
	"sentinel/packages/cache"
	Error "sentinel/packages/errs"
	"sentinel/packages/json"
	"sentinel/packages/util"

	emongo "github.com/StepanAnanin/EssentialMongoDB"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IndexedUser struct {
	ID        string   `bson:"_id" json:"_id"`
	Login     string   `bson:"login" json:"login"`
	Password  string   `bson:"password" json:"password"`
	Roles     []string `bson:"roles" json:"roles"`
	DeletedAt int      `bson:"deletedAt,omitmepty" json:"deletedAt"`
}

func findUserBy(key string, value string, deleted bool) (*IndexedUser, *Error.HTTP) {
	var user IndexedUser

	cacheKey := util.Ternary(deleted, cache.DeletedUserKeyPrefix, cache.UserKeyPrefix) + value

	if rawCachedUser, hit := cache.Get(cacheKey); hit {
		if cachedUser, ok := json.DecodeString[IndexedUser](rawCachedUser); ok {
			return &cachedUser, nil
		}
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	// There is a problem with searching user by ID, it works correctly only with primitive.ObjectID, idk why
	var uid primitive.ObjectID

	isKeyID := key == "_id"

	if isKeyID {
		objectID, e := primitive.ObjectIDFromHex(value)

		if e != nil {
			return &user, Error.NewHTTP("Invalid UID", http.StatusBadRequest)
		}

		uid = objectID
	}

	filter := util.Ternary(isKeyID, bson.D{{key, uid}}, bson.D{{key, value}})

	var cur *mongo.Cursor
	var err error

	if deleted {
		cur, err = DB.DeletedUserCollection.Find(ctx, filter)
	} else {
		cur, err = DB.UserCollection.Find(ctx, filter)
	}

	if err != nil {
		log.Fatalln(err)
	}

	if hasResult := cur.Next(ctx); !hasResult {
		return &user, Error.NewHTTP("Пользователь не был найден", http.StatusNotFound)
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
func FindUserByID(uid string) (*IndexedUser, *Error.HTTP) {
	return findUserBy("_id", uid, false)
}

// Search for soft deleted user with given UID
func FindSoftDeletedUserByID(uid string) (*IndexedUser, *Error.HTTP) {
	return findUserBy("_id", uid, true)
}

// Search for user with given UID, regardless of his deletion status
func FindAnyUserByID(uid string) (*IndexedUser, *Error.HTTP) {
	out, err := FindUserByID(uid)

	if err == nil {
		return out, err
	}

	if err.Status != http.StatusNotFound {
		return nil, err
	}

	return FindSoftDeletedUserByID(uid)
}

func FindUserByLogin(login string) (*IndexedUser, *Error.HTTP) {
	return findUserBy("login", login, false)
}
