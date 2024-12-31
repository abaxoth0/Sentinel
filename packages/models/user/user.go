package user

import (
	"context"
	"log"
	"net/http"
	"sentinel/packages/DB"
	"sentinel/packages/cache"
	Error "sentinel/packages/errs"
	"sentinel/packages/models/auth"
	"sentinel/packages/models/search"
	"sentinel/packages/util"

	emongo "github.com/StepanAnanin/EssentialMongoDB"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type Filter struct {
	TargetUID      string
	RequesterUID   string
	RequesterRoles []string
}

type user struct {
	Login    string
	Password string
	Roles    []string
}

func Create(login string, password string) (primitive.ObjectID, error) {
	var uid primitive.ObjectID

	if err := verifyPassword(password); err != nil {
		return uid, err
	}

	if _, err := search.FindUserByLogin(login); err == nil {
		return uid, Error.NewHTTP("Пользователь с таким логином уже существует.", http.StatusConflict)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)

	if err != nil {
		return uid, Error.NewHTTP("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	usr := user{
		Login:    login,
		Password: string(hashedPassword),
		Roles:    []string{auth.Host.OriginRoleName},
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	result, err := DB.UserCollection.InsertOne(ctx, usr)

	if err != nil {
		return uid, Error.NewHTTP("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	uid = result.InsertedID.(primitive.ObjectID)

	return uid, nil
}

func update(filter *Filter, upd *primitive.E, deleted bool) *Error.HTTP {
	if deleted {
		if _, err := search.FindSoftDeletedUserByID(filter.TargetUID); err != nil {
			return err
		}
	} else {
		_, err := search.FindUserByID(filter.TargetUID)

		if err != nil {
			return err
		}
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	update := bson.D{{"$set", bson.D{*upd}}}

	_, updError := DB.UserCollection.UpdateByID(ctx, emongo.ObjectIDFromHex(filter.TargetUID), update)

	if updError != nil {
		log.Println("[ ERROR ] Failed to update user (query error) \"" + filter.TargetUID + "\" - " + updError.Error())

		return Error.NewHTTP("Внутренняя ошибка сервера", http.StatusInternalServerError)
	}

	cache.Delete(cache.UserKeyPrefix + filter.TargetUID)

	return nil
}

func SoftDelete(filter *Filter) *Error.HTTP {
	if err := auth.Authorize(auth.Action.SoftDelete, auth.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := search.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewHTTP("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	user, err := search.FindUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	user.DeletedAt = int(util.UnixTimeNow())

	if e := emongo.DocumentTransfer(user, DB.UserCollection, DB.DeletedUserCollection, func() { user.DeletedAt = 0 }); e != nil {
		err = Error.NewHTTP(e.Error(), http.StatusInternalServerError)
	}

	if err != nil {
		return err
	}

	return nil
}

func Restore(filter *Filter) *Error.HTTP {
	if err := auth.Authorize(auth.Action.Restore, auth.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := search.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewHTTP("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	user, err := search.FindSoftDeletedUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	deletedAtTimestamp := user.DeletedAt
	user.DeletedAt = 0

	if e := emongo.DocumentTransfer(user, DB.DeletedUserCollection, DB.UserCollection, func() { user.DeletedAt = deletedAtTimestamp }); e != nil {
		err = Error.NewHTTP(e.Error(), http.StatusInternalServerError)
	}

	if err != nil {
		return err
	}

	return nil
}

// Hard delete
func Drop(filter *Filter) *Error.HTTP {
	if err := auth.Authorize(auth.Action.Drop, auth.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := search.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewHTTP("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	user, err := search.FindAnyUserByID(filter.TargetUID)
	deleted := user.DeletedAt == 0

	if err != nil {
		return err
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	collection := util.Ternary(deleted, DB.DeletedUserCollection, DB.UserCollection)

	if _, e := collection.DeleteOne(ctx, bson.D{{"_id", filter.TargetUID}}); e != nil {
		return Error.NewHTTP("Не удалось удалить пользователя", http.StatusInternalServerError)
	}

	cacheKeyPrefix := util.Ternary(deleted, cache.DeletedUserKeyPrefix, cache.UserKeyPrefix)
	cache.Delete(cacheKeyPrefix + filter.TargetUID)

	return nil
}

func DropAllDeleted(requesterRoles []string) *Error.HTTP {
	if err := auth.Authorize(auth.Action.DropAllDeleted, auth.Resource.User, requesterRoles); err != nil {
		return err
	}

	_, err := DB.DeletedUserCollection.DeleteMany(context.TODO(), bson.D{})

	if err != nil {
		return Error.NewHTTP("Operation failed (Internal Server Error)", http.StatusInternalServerError)
	}

	return nil
}

func ChangeLogin(filter *Filter, newlogin string) *Error.HTTP {
	if err := auth.Authorize(auth.Action.ChangeLogin, auth.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := search.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewHTTP("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	_, err := search.FindUserByLogin(newlogin)

	// user with new login was found
	if err == nil {
		return Error.NewHTTP("Данный логин уже занят", http.StatusConflict)
	}

	upd := &primitive.E{"login", newlogin}

	return update(filter, upd, true)
}

func ChangePassword(filter *Filter, newPassword string) *Error.HTTP {
	if err := auth.Authorize(auth.Action.ChangePassword, auth.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := search.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewHTTP("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	if err := verifyPassword(newPassword); err != nil {
		return err
	}

	hashedPassword, e := bcrypt.GenerateFromPassword([]byte(newPassword), 12)

	if e != nil {
		return Error.NewHTTP("Не удалось изменить пароль: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	upd := &primitive.E{"password", hashedPassword}

	return update(filter, upd, true)
}

func ChangeRole(filter *Filter, newRole string) *Error.HTTP {
	if err := auth.Authorize(auth.Action.ChangeRole, auth.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := search.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewHTTP("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	upd := &primitive.E{"role", newRole}

	return update(filter, upd, true)
}

func CheckIsLoginExists(login string) (bool, *Error.HTTP) {
	if _, err := search.FindUserByLogin(login); err != nil {
		if err.Status == http.StatusNotFound {
			return false, nil
		}

		return true, err
	}

	return true, nil
}

func GetRoles(filter *Filter) ([]string, *Error.HTTP) {
	if err := auth.Authorize(auth.Action.GetRole, auth.Resource.User, filter.RequesterRoles); err != nil {
		return []string{}, err
	}

	user, err := search.FindUserByID(filter.TargetUID)

	if err != nil {
		return []string{}, Error.NewHTTP("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	return user.Roles, nil
}
