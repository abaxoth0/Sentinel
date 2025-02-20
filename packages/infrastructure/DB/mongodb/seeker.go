package mongodb

import (
	"log"
	"net/http"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errs"
	"sentinel/packages/infrastructure/cache"
	datamodel "sentinel/packages/presentation/data"
	"sentinel/packages/util"

	emongo "github.com/StepanAnanin/EssentialMongoDB"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type seeker struct {
	//
}

func (_ *seeker) findUserBy(key string, value string, deleted bool) (*UserDTO.Indexed, *Error.Status) {
	var user UserDTO.Indexed

	cacheKey := util.Ternary(deleted, cache.DeletedUserKeyPrefix, cache.UserKeyPrefix) + value

	if rawCachedUser, hit:= cache.Get(cacheKey); hit {
		if cachedUser, err:= datamodel.DecodeString[UserDTO.Indexed](rawCachedUser); err != nil {
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
			return &user, Error.NewStatusError("Invalid UID", http.StatusBadRequest)
		}

		uid = objectID
	}

	// TODO test queries (changed from D to M)
    filter := util.Ternary(isKeyID, bson.D{{key, uid}}, bson.D{{key, value}})

	var cur *mongo.Cursor
	var err error

	if deleted {
		cur, err = driver.DeletedUserCollection.Find(ctx, filter)
	} else {
		cur, err = driver.UserCollection.Find(ctx, filter)
	}

	if err != nil {
		log.Fatalln(err)
	}

	if hasResult := cur.Next(ctx); !hasResult {
		return &user, Error.NewStatusError("Пользователь не был найден", http.StatusNotFound)
	}

	err = cur.Decode(&user)

	if err != nil {
		log.Fatalln(err)
	}

	// This actually not a critical problem, cuz on finishing request processing goroutine will be terminated
	// and garbage collector should kill cursor, but idk how it will work in practice.
	// user will be non-empty, but error will still presence
	if err := cur.Close(ctx); err != nil {
		log.Printf("[ ERROR ] Failed to close cursor. ID: %s, Login: %s\n", user.ID, user.Login)
	}

	if rawUser, err := datamodel.Encode(user); err != nil {
		cache.Set(cacheKey, rawUser)
	}

	return &user, nil
}

// Search for not deleted user with given UID
func (seeker *seeker) FindUserByID(uid string) (*UserDTO.Indexed, *Error.Status) {
	return seeker.findUserBy("_id", uid, false)
}

// Search for soft deleted user with given UID
func (seeker *seeker) FindSoftDeletedUserByID(uid string) (*UserDTO.Indexed, *Error.Status) {
	return seeker.findUserBy("_id", uid, true)
}

// Search for user with given UID, regardless of his deletion status
func (seeker *seeker) FindAnyUserByID(uid string) (*UserDTO.Indexed, *Error.Status) {
	out, err := seeker.FindUserByID(uid)

	if err == nil {
		return out, err
	}

	if err.Status != http.StatusNotFound {
		return nil, err
	}

	return seeker.FindSoftDeletedUserByID(uid)
}

func (seeker *seeker) FindUserByLogin(login string) (*UserDTO.Indexed, *Error.Status) {
	return seeker.findUserBy("login", login, false)
}

// TODO add cache
func (seeker *seeker) IsLoginExists(login string) (bool, *Error.Status) {
	if _, err := seeker.FindUserByLogin(login); err != nil {
		if err.Status == http.StatusNotFound {
			return false, nil
		}

		return true, err
	}

	return true, nil
}
