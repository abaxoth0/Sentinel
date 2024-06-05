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

	passwordSize := len(password)

	// bcrypt can handle password with maximum size of 72 bytes
	if passwordSize < 8 || passwordSize > 64 {
		return uid, ExternalError.New("Недопустимый размер пароля. Пароль должен находится в диапозоне от 8 до 64 символов.", http.StatusBadRequest)
	}

	_, err := m.search.FindUserByEmail(email)

	if err == nil {
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

func (m *Model) SoftDelete(targetUID string, requesterUID string, requesterRole role.Role) error {
	_, err := m.search.FindUserByID(targetUID)

	// There are 1 case in wich error will presence but user will be non-empty:
	// If failed to close db cursor. (See comment in "findUserBy" method for detatils)
	if err != nil {
		if isExternal, _ := ExternalError.Is(err); isExternal {
			return ExternalError.New("Пользователь не был найден", http.StatusNotFound)
		}

		log.Fatalln(err.Error())
	}

	// If user want to delete not himself, but another user and he isn't authorize to do that
	if targetUID != requesterUID {
		if err := auth.Rulebook.SoftDeleteUser.Authorize(requesterRole); err != nil {
			return err
		}
	}

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	update := bson.D{{
		"$set", bson.D{{
			"deletedAt", util.UnixTimeNow(),
		}},
	}}

	_, err = m.collection.UpdateByID(ctx, DB.ObjectIDFromHex(targetUID), update)

	if err != nil {
		log.Println("[ ERROR ] Failed to update user (query error) \"" + targetUID + "\" - " + err.Error())

		return ExternalError.New("Внутренняя ошибка сервера", http.StatusInternalServerError)
	}

	return nil
}

func (m *Model) Restore(targetID string, requesterID string, requesterRole role.Role) error {
	return nil
}
