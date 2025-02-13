package mongodb

import (
	"context"
	"log"
	"net/http"
	"sentinel/packages/core/user"
	userdto "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errs"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/util"

	emongo "github.com/StepanAnanin/EssentialMongoDB"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// TODO a lot of warnings on queries, need to check them

type repository struct {
	//
}

func (repo *repository) Create(login string, password string) (string, error) {
	// var uid primitive.ObjectID

	if err := user.VerifyPassword(password); err != nil {
		return "", err
	}

	if _, err := driver.FindUserByLogin(login); err == nil {
		return "", Error.NewStatusError("Пользователь с таким логином уже существует.", http.StatusConflict)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)

	if err != nil {
		return "", Error.NewStatusError("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	user := user.RawSecured{
		Password: string(hashedPassword),
		Raw: user.Raw{
			Login: login,
			Roles: []string{authorization.Host.OriginRoleName},
		},
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	result, err := driver.UserCollection.InsertOne(ctx, user)

	if err != nil {
		return "", Error.NewStatusError("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	// TODO test this func
	uid := result.InsertedID.(string)

	return uid, nil
}

func (repo *repository) update(filter *userdto.Filter, upd *primitive.E, deleted bool) *Error.Status {
	if deleted {
		if _, err := driver.FindSoftDeletedUserByID(filter.TargetUID); err != nil {
			return err
		}
	} else {
		if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
			return err
		}
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	update := bson.D{{"$set", bson.D{*upd}}}

	_, updError := driver.UserCollection.UpdateByID(ctx, emongo.ObjectIDFromHex(filter.TargetUID), update)

	if updError != nil {
		log.Println("[ ERROR ] Failed to update user (query error) \"" + filter.TargetUID + "\" - " + updError.Error())

		return Error.NewStatusError("Внутренняя ошибка сервера", http.StatusInternalServerError)
	}

	cache.Delete(cache.UserKeyPrefix + filter.TargetUID)

	return nil
}

func (repo *repository) SoftDelete(filter *userdto.Filter) *Error.Status {
	if err := authorization.Authorize(authorization.Action.SoftDelete, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewStatusError("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	user, err := driver.FindUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	user.DeletedAt = int(util.UnixTimeNow())

	if e := emongo.DocumentTransfer(user, driver.UserCollection, driver.DeletedUserCollection, func() { user.DeletedAt = 0 }); e != nil {
		err = Error.NewStatusError(e.Error(), http.StatusInternalServerError)
	}

	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) Restore(filter *userdto.Filter) *Error.Status {
	if err := authorization.Authorize(authorization.Action.Restore, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewStatusError("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	user, err := driver.FindSoftDeletedUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	deletedAtTimestamp := user.DeletedAt
	user.DeletedAt = 0

	if e := emongo.DocumentTransfer(user, driver.DeletedUserCollection, driver.UserCollection, func() { user.DeletedAt = deletedAtTimestamp }); e != nil {
		err = Error.NewStatusError(e.Error(), http.StatusInternalServerError)
	}

	if err != nil {
		return err
	}

	return nil
}

// Hard delete
func (repo *repository) Drop(filter *userdto.Filter) *Error.Status {
	if err := authorization.Authorize(authorization.Action.Drop, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewStatusError("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	user, err := driver.FindAnyUserByID(filter.TargetUID)
	deleted := user.DeletedAt == 0

	if err != nil {
		return err
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	collection := util.Ternary(deleted, driver.DeletedUserCollection, driver.UserCollection)

	if _, e := collection.DeleteOne(ctx, bson.D{{"_id", filter.TargetUID}}); e != nil {
		return Error.NewStatusError("Не удалось удалить пользователя", http.StatusInternalServerError)
	}

	cacheKeyPrefix := util.Ternary(deleted, cache.DeletedUserKeyPrefix, cache.UserKeyPrefix)
	cache.Delete(cacheKeyPrefix + filter.TargetUID)

	return nil
}

func (repo *repository) DropAllSoftDeleted(requesterRoles []string) *Error.Status {
	if err := authorization.Authorize(authorization.Action.DropAllSoftDeleted, authorization.Resource.User, requesterRoles); err != nil {
		return err
	}

	_, err := driver.DeletedUserCollection.DeleteMany(context.TODO(), bson.D{})

	if err != nil {
		return Error.NewStatusError("Operation failed (Internal Server Error)", http.StatusInternalServerError)
	}

	return nil
}

func (repo *repository) ChangeLogin(filter *userdto.Filter, newlogin string) *Error.Status {
	if err := authorization.Authorize(authorization.Action.ChangeLogin, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewStatusError("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	_, err := driver.FindUserByLogin(newlogin)

	// user with new login was found
	if err == nil {
		return Error.NewStatusError("Данный логин уже занят", http.StatusConflict)
	}

	upd := &primitive.E{"login", newlogin}

	return repo.update(filter, upd, true)
}

func (repo *repository) ChangePassword(filter *userdto.Filter, newPassword string) *Error.Status {
	if err := authorization.Authorize(authorization.Action.ChangePassword, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewStatusError("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	if err := user.VerifyPassword(newPassword); err != nil {
		return err
	}

	hashedPassword, e := bcrypt.GenerateFromPassword([]byte(newPassword), 12)

	if e != nil {
		return Error.NewStatusError("Не удалось изменить пароль: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	upd := &primitive.E{"password", hashedPassword}

	return repo.update(filter, upd, true)
}

func (repo *repository) ChangeRoles(filter *userdto.Filter, newRoles []string) *Error.Status {
	if err := authorization.Authorize(authorization.Action.ChangeRoles, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
		return Error.NewStatusError("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	upd := &primitive.E{"roles", newRoles}

	return repo.update(filter, upd, true)
}

func (repo *repository) GetRoles(filter *userdto.Filter) ([]string, *Error.Status) {
	if err := authorization.Authorize(authorization.Action.GetRole, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return []string{}, err
	}

	user, err := driver.FindUserByID(filter.TargetUID)

	if err != nil {
		return []string{}, Error.NewStatusError("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	return user.Roles, nil
}
