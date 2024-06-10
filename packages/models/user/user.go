package user

import (
	"log"
	"net/http"
	"sentinel/packages/DB"
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
	Email    string
	Password string
	Role     role.Role
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

func (m *Model) Create(email string, password string) (primitive.ObjectID, error) {
	var uid primitive.ObjectID

	if err := verifyPassword(password); err != nil {
		return uid, err
	}

	if _, err := m.search.FindUserByEmail(email); err == nil {
		// Invalid email or password, currently we know only about email,
		// but there are no point to tell user about this, due to security reasons.
		return uid, ExternalError.New("Пользователь с таким e-mail'ом уже существует.", http.StatusConflict)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)

	if err != nil {
		return uid, ExternalError.New("Не удалось создать пользователя: Внутреняя ошибка сервера.", http.StatusInternalServerError)
	}

	user := user{
		Email:    email,
		Password: string(hashedPassword),
		// TODO Add possibility to control, which role will be assigned for each new user
		Role: role.UnconfirmedUser,
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
	var err *ExternalError.Error

	if deleted {
		_, err = m.search.FindSoftDeletedUserByID(filter.TargetUID)
	} else {
		_, err = m.search.FindUserByID(filter.TargetUID)
	}

	// There is 1 case in wich error will presence but user will be non-empty:
	// If failed to close db cursor. (See comment in "findUserBy" method for detatils)
	if err != nil {
		if isExternal, _ := ExternalError.Is(err); isExternal {
			return ExternalError.New("Пользователь не был найден", http.StatusNotFound)
		}

		log.Fatalln(err.Error())
	}

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	update := bson.D{{"$set", bson.D{*upd}}}

	_, updError := m.collection.UpdateByID(ctx, DB.ObjectIDFromHex(filter.TargetUID), update)

	if updError != nil {
		log.Println("[ ERROR ] Failed to update user (query error) \"" + filter.TargetUID + "\" - " + updError.Error())

		return ExternalError.New("Внутренняя ошибка сервера", http.StatusInternalServerError)
	}

	return nil
}

func (m *Model) SoftDelete(filter *Filter) *ExternalError.Error {
	// If user want to delete not himself, but another user and he isn't authorize to do that
	if filter.TargetUID != filter.RequesterUID {
		if err := auth.Rulebook.SoftDeleteUser.Authorize(filter.RequesterRole); err != nil {
			return err
		}
	}

	targetUser, err := m.search.FindUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	if targetUser.Role == role.Administrator {
		return ExternalError.New("Невозможно удалить пользователя с ролью администратора. (Обратитесь напрямую в базу данных)", http.StatusForbidden)
	}

	upd := &primitive.E{"deletedAt", util.UnixTimeNow()}

	return m.update(filter, upd, false)
}

func (m *Model) Restore(filter *Filter) *ExternalError.Error {
	if err := auth.Rulebook.RestoreSoftDeletedUser.Authorize(filter.RequesterRole); err != nil {
		return err
	}

	upd := &primitive.E{"deletedAt", primitive.Null{}}

	return m.update(filter, upd, true)
}

func (m *Model) ChangeEmail(filter *Filter, newEmail string) *ExternalError.Error {
	if err := auth.Rulebook.RestoreSoftDeletedUser.Authorize(filter.RequesterRole); err != nil {
		return err
	}

	// Need to ensure that new email is not already used by some other user,
	// for that err must be not nil and have a type of ExternalError,
	// if this both condition satisfied then user with this email wasn't found.
	// (which means that it can be used)
	_, err := m.search.FindUserByEmail(newEmail)

	// user with new email was found
	if err == nil {
		return ExternalError.New("Данный E-Mail уже занят", http.StatusConflict)
	}

	// Check is error external. (if not -> return error)
	// External Error returned only if user wasn't found
	if isExternal, _ := ExternalError.Is(err); !isExternal {
		return ExternalError.New("Не удалось подтвердить доступность запрошенного E-Mail'а: Внутренняя ошибка сервера", http.StatusInternalServerError)
	}

	upd := &primitive.E{"email", newEmail}

	return m.update(filter, upd, true)
}

func (m *Model) ChangePassword(filter *Filter, newPassword string) *ExternalError.Error {
	if err := auth.Rulebook.RestoreSoftDeletedUser.Authorize(filter.RequesterRole); err != nil {
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
	if err := auth.Rulebook.RestoreSoftDeletedUser.Authorize(filter.RequesterRole); err != nil {
		return err
	}

	upd := &primitive.E{"role", newRole}

	return m.update(filter, upd, true)
}
