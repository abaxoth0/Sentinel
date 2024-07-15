package user

import (
	"log"
	"net/http"
	"sentinel/packages/DB"
	"sentinel/packages/cache"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/auth"
	"sentinel/packages/models/role"
	"sentinel/packages/models/search"
	"sentinel/packages/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	Login     string
	Password  string
	Role      role.Role
	DeletedAt int
}

type Filter struct {
	TargetUID     string
	RequesterUID  string
	RequesterRole role.Role
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
		// TODO Add possibility to control, which role will be assigned for each new user
		Role:      role.UnconfirmedUser,
		DeletedAt: 0,
	}

	ctx, cancel := DB.DefaultTimeoutContext()

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
		if _, err := search.FindUserByID(filter.TargetUID); err != nil {
			return err
		}
	}

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	update := bson.D{{"$set", bson.D{*upd}}}

	_, updError := DB.UserCollection.UpdateByID(ctx, DB.ObjectIDFromHex(filter.TargetUID), update)

	if updError != nil {
		log.Println("[ ERROR ] Failed to update user (query error) \"" + filter.TargetUID + "\" - " + updError.Error())

		return ExternalError.New("Внутренняя ошибка сервера", http.StatusInternalServerError)
	}

	cache.Delete(cache.UserKeyPrefix + filter.TargetUID)

	return nil
}

func SoftDelete(filter *Filter) *ExternalError.Error {
	if filter.TargetUID != filter.RequesterUID {
		if err := auth.Rulebook.SoftDeleteUser.Authorize(filter.RequesterRole); err != nil {
			return err
		}
	}

	if isAdmin, err := isUserAdmin(filter.TargetUID); isAdmin {
		if err != nil {
			return err
		}

		return ExternalError.New("Невозможно удалить пользователя с ролью администратора. (Обратитесь напрямую в базу данных)", http.StatusForbidden)
	}

	upd := &primitive.E{"deletedAt", util.UnixTimeNow()}

	return update(filter, upd, false)
}

func Restore(filter *Filter) *ExternalError.Error {
	if err := auth.Rulebook.RestoreSoftDeletedUser.Authorize(filter.RequesterRole); err != nil {
		return err
	}

	x, err := search.FindSoftDeletedUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	println(x.ID)

	upd := &primitive.E{"deletedAt", 0}

	return update(filter, upd, true)
}

// Hard delete
func Drop(filter *Filter) *ExternalError.Error {
	if filter.TargetUID != filter.RequesterUID {
		if err := auth.Rulebook.DropUser.Authorize(filter.RequesterRole); err != nil {
			return err
		}
	}

	user, err := search.FindAnyUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	if user.Role == role.Administrator {
		return ExternalError.New("Невозможно удалить пользователя с ролью администратора. (Обратитесь напрямую в базу данных)", http.StatusForbidden)
	}

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	if _, e := DB.UserCollection.DeleteOne(ctx, bson.D{{"_id", filter.TargetUID}}); e != nil {
		return ExternalError.New("Не удалось удалить пользователя", http.StatusInternalServerError)
	}

	cache.Delete(cache.UserKeyPrefix + filter.TargetUID)

	return nil
}

func ChangeLogin(filter *Filter, newlogin string) *ExternalError.Error {
	if err := auth.Rulebook.ChangeUserLogin.Authorize(filter.RequesterRole); err != nil {
		return err
	}

	_, err := search.FindUserByLogin(newlogin)

	// user with new login was found
	if err == nil {
		return ExternalError.New("Данный логин уже занят", http.StatusConflict)
	}

	upd := &primitive.E{"login", newlogin}

	return update(filter, upd, true)
}

func ChangePassword(filter *Filter, newPassword string) *ExternalError.Error {
	if err := auth.Rulebook.ChangeUserPassword.Authorize(filter.RequesterRole); err != nil {
		return err
	}

	if err := verifyPassword(newPassword); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)

	if err != nil {
		return ExternalError.New("Не удалось изменить пароль: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	upd := &primitive.E{"password", hashedPassword}

	return update(filter, upd, true)
}

func ChangeRole(filter *Filter, newRole string) *ExternalError.Error {
	if err := auth.Rulebook.ChangeUserRole.Authorize(filter.RequesterRole); err != nil {
		return err
	}

	if isAdmin, err := isUserAdmin(filter.TargetUID); isAdmin {
		if err != nil {
			return err
		}

		return ExternalError.New("Невозможно изменить роль администратора. (Обратитесь напрямую в базу данных)", http.StatusForbidden)
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

func GetRole(filter *Filter) (role.Role, *ExternalError.Error) {
	var emptyRole role.Role

	if err := auth.Rulebook.GetUserRole.Authorize(filter.RequesterRole); err != nil {
		return emptyRole, err
	}

	user, err := search.FindUserByID(filter.TargetUID)

	if err != nil {
		return emptyRole, err
	}

	return user.Role, nil
}

func isUserAdmin(uid string) (bool, *ExternalError.Error) {
	targetUser, err := search.FindUserByID(uid)

	if err != nil {
		return false, err
	}

	return targetUser.Role == role.Administrator, nil
}
