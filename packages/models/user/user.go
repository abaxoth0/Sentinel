package user

import (
	"context"
	"log"
	"net/http"
	"sentinel/packages/DB"
	"sentinel/packages/cache"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/auth"
	"sentinel/packages/models/role"
	"sentinel/packages/models/search"
	"sentinel/packages/util"
	"slices"

	emongo "github.com/StepanAnanin/EssentialMongoDB"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	Login     string
	Password  string
	Role      string
	DeletedAt int
}

type Filter struct {
	TargetUID     string
	RequesterUID  string
	RequesterRole string
}

func Create(login string, password string) (primitive.ObjectID, error) {
	var uid primitive.ObjectID

	if err := verifyPassword(password); err != nil {
		return uid, err
	}

	if _, err := search.FindUserByLogin(login); err == nil {
		return uid, ExternalError.New("Пользователь с таким логином уже существует.", http.StatusConflict)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)

	if err != nil {
		return uid, ExternalError.New("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	user := user{
		Login:    login,
		Password: string(hashedPassword),
		// TODO FIX that, remove hardcode
		// TODO Add possibility to control, which role will be assigned for each new user
		Role: "unconfirmed_user",
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	result, err := DB.UserCollection.InsertOne(ctx, user)

	if err != nil {
		return uid, ExternalError.New("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	uid = result.InsertedID.(primitive.ObjectID)

	return uid, nil
}

func update(filter *Filter, upd *primitive.E, deleted bool) *ExternalError.Error {
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

		return ExternalError.New("Внутренняя ошибка сервера", http.StatusInternalServerError)
	}

	cache.Delete(cache.UserKeyPrefix + filter.TargetUID)

	return nil
}

func SoftDelete(filter *Filter) *ExternalError.Error {
	targetUser, err := search.FindUserByID(filter.TargetUID)

	if err != nil {
		return ExternalError.New("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	if err := auth.Rulebook.SoftDeleteUser.Authorize(filter.RequesterRole, targetUser.Role); err != nil {
		return err
	}

	user, err := search.FindUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	user.DeletedAt = int(util.UnixTimeNow())

	if e := emongo.DocumentTransfer(user, DB.UserCollection, DB.DeletedUserCollection, func() { user.DeletedAt = 0 }); e != nil {
		err = ExternalError.New(e.Error(), http.StatusInternalServerError)
	}

	if err != nil {
		return err
	}

	return nil
}

func Restore(filter *Filter) *ExternalError.Error {
	targetUser, err := search.FindUserByID(filter.TargetUID)

	if err != nil {
		return ExternalError.New("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	if err := auth.Rulebook.RestoreSoftDeletedUser.Authorize(filter.RequesterRole, targetUser.Role); err != nil {
		return err
	}

	user, err := search.FindSoftDeletedUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	deletedAtTimestamp := user.DeletedAt
	user.DeletedAt = 0

	if e := emongo.DocumentTransfer(user, DB.DeletedUserCollection, DB.UserCollection, func() { user.DeletedAt = deletedAtTimestamp }); e != nil {
		err = ExternalError.New(e.Error(), http.StatusInternalServerError)
	}

	if err != nil {
		return err
	}

	return nil
}

// Hard delete
func Drop(filter *Filter) *ExternalError.Error {
	targetUser, err := search.FindUserByID(filter.TargetUID)

	if err != nil {
		return ExternalError.New("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	if err := auth.Rulebook.DropUser.Authorize(filter.RequesterRole, targetUser.Role); err != nil {
		return err
	}

	user, err := search.FindAnyUserByID(filter.TargetUID)
	deleted := user.DeletedAt == 0

	if err != nil {
		return err
	}

	userRole, err := role.GetAuthRole(user.Role)

	if err != nil {
		return err
	}

	if slices.Contains(userRole.Permissions, role.AdminPermissionTag) {
		return ExternalError.New("Невозможно удалить пользователя с ролью администратора. (Обратитесь напрямую в базу данных)", http.StatusForbidden)
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	collection := util.Ternary(deleted, DB.DeletedUserCollection, DB.UserCollection)

	if _, e := collection.DeleteOne(ctx, bson.D{{"_id", filter.TargetUID}}); e != nil {
		return ExternalError.New("Не удалось удалить пользователя", http.StatusInternalServerError)
	}

	cacheKeyPrefix := util.Ternary(deleted, cache.DeletedUserKeyPrefix, cache.UserKeyPrefix)
	cache.Delete(cacheKeyPrefix + filter.TargetUID)

	return nil
}

func DropAllDeleted(requesterRole string) *ExternalError.Error {
	if err := auth.Rulebook.DropAllDeletedUsers.Authorize(requesterRole, role.NoneRole); err != nil {
		return err
	}

	_, err := DB.DeletedUserCollection.DeleteMany(context.TODO(), bson.D{})

	if err != nil {
		return ExternalError.New("Operation failed (Internal Server Error)", http.StatusInternalServerError)
	}

	return nil
}

func ChangeLogin(filter *Filter, newlogin string) *ExternalError.Error {
	targetUser, err := search.FindUserByID(filter.TargetUID)

	if err != nil {
		return ExternalError.New("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	if err := auth.Rulebook.ChangeUserLogin.Authorize(filter.RequesterRole, targetUser.Role); err != nil {
		return err
	}

	_, err = search.FindUserByLogin(newlogin)

	// user with new login was found
	if err == nil {
		return ExternalError.New("Данный логин уже занят", http.StatusConflict)
	}

	upd := &primitive.E{"login", newlogin}

	return update(filter, upd, true)
}

func ChangePassword(filter *Filter, newPassword string) *ExternalError.Error {
	targetUser, err := search.FindUserByID(filter.TargetUID)

	if err != nil {
		return ExternalError.New("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	if err := auth.Rulebook.ChangeUserPassword.Authorize(filter.RequesterRole, targetUser.Role); err != nil {
		return err
	}

	if err := verifyPassword(newPassword); err != nil {
		return err
	}

	hashedPassword, e := bcrypt.GenerateFromPassword([]byte(newPassword), 12)

	if e != nil {
		return ExternalError.New("Не удалось изменить пароль: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	upd := &primitive.E{"password", hashedPassword}

	return update(filter, upd, true)
}

func ChangeRole(filter *Filter, newRole string) *ExternalError.Error {
	targetUser, err := search.FindUserByID(filter.TargetUID)

	if err != nil {
		return ExternalError.New("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	if err := auth.Rulebook.ChangeUserRole.Authorize(filter.RequesterRole, targetUser.Role); err != nil {
		return err
	}

	upd := &primitive.E{"role", newRole}

	return update(filter, upd, true)
}

func CheckIsLoginExists(login string) (bool, *ExternalError.Error) {
	if _, err := search.FindUserByLogin(login); err != nil {
		if err.Status == http.StatusNotFound {
			return false, nil
		}

		return true, err
	}

	return true, nil
}

func GetRole(filter *Filter) (string, *ExternalError.Error) {
	var emptyRole string

	user, err := search.FindUserByID(filter.TargetUID)

	if err != nil {
		return emptyRole, ExternalError.New("Запрошенный пользователь не был найден", http.StatusNotFound)
	}

	if err := auth.Rulebook.GetUserRole.Authorize(filter.RequesterRole, user.Role); err != nil {
		return emptyRole, err
	}

	return user.Role, nil
}
