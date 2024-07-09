package user

import (
	"log"
	"net/http"
	"sentinel/packages/DB"
	"sentinel/packages/cache"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/auth"
	"sentinel/packages/models/role"
	"sentinel/packages/models/search"
	"sentinel/packages/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

type Model struct {
	search     *search.Model
	dbClient   *mongo.Client
	collection *mongo.Collection
}

func New(dbClient *mongo.Client, searchModel *search.Model) *Model {
	return &Model{
		search:     searchModel,
		dbClient:   dbClient,
		collection: dbClient.Database(config.DB.Name).Collection(config.DB.UserCollectionName),
	}
}

func (m *Model) Create(login string, password string) (primitive.ObjectID, error) {
	var uid primitive.ObjectID

	if err := verifyPassword(password); err != nil {
		return uid, err
	}

	if _, err := m.search.FindUserByLogin(login); err == nil {
		// Invalid login or password, currently we know only about login,
		// but there are no point to tell user about this, due to security reasons.
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

	result, err := m.collection.InsertOne(ctx, user)

	if err != nil {
		return uid, ExternalError.New("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	uid = result.InsertedID.(primitive.ObjectID)

	return uid, nil
}

func (m *Model) update(filter *Filter, upd *primitive.E, deleted bool) *ExternalError.Error {
	if deleted {
		if _, err := m.search.FindSoftDeletedUserByID(filter.TargetUID); err != nil {
			return err
		}
	} else {
		if _, err := m.search.FindUserByID(filter.TargetUID); err != nil {
			return err
		}
	}

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	update := bson.D{{"$set", bson.D{*upd}}}

	_, updError := m.collection.UpdateByID(ctx, DB.ObjectIDFromHex(filter.TargetUID), update)

	if updError != nil {
		log.Println("[ ERROR ] Failed to update user (query error) \"" + filter.TargetUID + "\" - " + updError.Error())

		return ExternalError.New("Внутренняя ошибка сервера", http.StatusInternalServerError)
	}

	cache.Delete(cache.UserKeyPrefix + filter.TargetUID)

	return nil
}

func (m *Model) SoftDelete(filter *Filter) *ExternalError.Error {
	// If user want to delete not himself, but another user and he isn't authorize to do that
	if filter.TargetUID != filter.RequesterUID {
		if err := auth.Rulebook.SoftDeleteUser.Authorize(filter.RequesterRole); err != nil {
			return err
		}
	}

	if isAdmin, err := m.isUserAdmin(filter.TargetUID); isAdmin {
		if err != nil {
			return err
		}

		return ExternalError.New("Невозможно удалить пользователя с ролью администратора. (Обратитесь напрямую в базу данных)", http.StatusForbidden)
	}

	upd := &primitive.E{"deletedAt", util.UnixTimeNow()}

	return m.update(filter, upd, false)
}

func (m *Model) Restore(filter *Filter) *ExternalError.Error {
	if err := auth.Rulebook.RestoreSoftDeletedUser.Authorize(filter.RequesterRole); err != nil {
		return err
	}

	x, err := m.search.FindSoftDeletedUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	println(x.ID)

	upd := &primitive.E{"deletedAt", 0}

	return m.update(filter, upd, true)
}

// Hard delete
func (m *Model) Drop(filter *Filter) *ExternalError.Error {
	// If user want to delete not himself, but another user and he isn't authorize to do that
	if filter.TargetUID != filter.RequesterUID {
		if err := auth.Rulebook.DropUser.Authorize(filter.RequesterRole); err != nil {
			return err
		}
	}

	user, err := m.search.FindAnyUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	if user.Role == role.Administrator {
		return ExternalError.New("Невозможно удалить пользователя с ролью администратора. (Обратитесь напрямую в базу данных)", http.StatusForbidden)
	}

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	if _, e := m.collection.DeleteOne(ctx, bson.D{{"_id", filter.TargetUID}}); e != nil {
		return ExternalError.New("Не удалось удалить пользователя", http.StatusInternalServerError)
	}

	cache.Delete(cache.UserKeyPrefix + filter.TargetUID)

	return nil
}

func (m *Model) ChangeLogin(filter *Filter, newlogin string) *ExternalError.Error {
	if err := auth.Rulebook.ChangeUserLogin.Authorize(filter.RequesterRole); err != nil {
		return err
	}

	// Need to ensure that new login is not already used by some other user,
	// for that err must be not nil and have a type of ExternalError,
	// if this both condition satisfied then user with this login wasn't found.
	// (which means that it can be used)
	_, err := m.search.FindUserByLogin(newlogin)

	// user with new login was found
	if err == nil {
		return ExternalError.New("Данный логин уже занят", http.StatusConflict)
	}

	upd := &primitive.E{"login", newlogin}

	return m.update(filter, upd, true)
}

func (m *Model) ChangePassword(filter *Filter, newPassword string) *ExternalError.Error {
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

	return m.update(filter, upd, true)
}

func (m *Model) ChangeRole(filter *Filter, newRole string) *ExternalError.Error {
	if err := auth.Rulebook.ChangeUserRole.Authorize(filter.RequesterRole); err != nil {
		return err
	}

	if isAdmin, err := m.isUserAdmin(filter.TargetUID); isAdmin {
		if err != nil {
			return err
		}

		return ExternalError.New("Невозможно изменить роль администратора. (Обратитесь напрямую в базу данных)", http.StatusForbidden)
	}

	upd := &primitive.E{"role", newRole}

	return m.update(filter, upd, true)
}

func (m *Model) CheckIsLoginExists(login string) (bool, *ExternalError.Error) {
	if _, err := m.search.FindUserByLogin(login); err != nil {
		if err.Status == http.StatusNotFound {
			return false, nil
		}

		return true, err
	}

	return true, nil
}

func (m *Model) isUserAdmin(uid string) (bool, *ExternalError.Error) {
	targetUser, err := m.search.FindUserByID(uid)

	if err != nil {
		return false, err
	}

	return targetUser.Role == role.Administrator, nil
}
