package user

import (
	"log"
	"net/http"
	"sentinel/packages/DB"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/role"
	"sentinel/packages/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// TODO add timeouts for DB queries (https://www.mongodb.com/docs/drivers/go/current/fundamentals/context/)

type user struct {
	Email    string
	Password string
	Role     string
}

type indexedUser struct {
	ID       string `bson:"_id"`
	Email    string
	Password string
	Role     string
	// If in DB this property will be nil, then here it will be 0
	DeletedAt int `bson:"deletedAt,omitempty"`
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

// TODO Actually better to create an additional model `auth` and move this method in it. (According to SRP)
// 		But since it's only 1 method i won't do that, cuz it's not realy confusing.

// Returns indexedUser if auth data is correct, ExternalError otherwise.
func (m Model) Login(email string, password string) (indexedUser, *ExternalError.ExternalError) {
	user, err := m.FindUserByEmail(email)

	// If user was found (user != indexedUser{}) and there are error, that means cursor closing failed. (see `findUserBy` method)
	// If user wasn't found and there are error, that means occured an unexpected error.
	if err != nil && (user != indexedUser{}) {
		return user, ExternalError.New("Неверный e-mail или пароль", http.StatusBadRequest)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// We know that on this stage incorrect only password,
		// but there are no point to tell user about this, due to security reasons.
		return user, ExternalError.New("Неверный e-mail или пароль", http.StatusBadRequest)
	}

	return user, nil
}

func (m Model) Create(email string, password string) (primitive.ObjectID, error) {
	var uid primitive.ObjectID

	passwordSize := len(password)

	// bcrypt can handle password with maximum size of 72 bytes
	if passwordSize < 8 || passwordSize > 64 {
		return uid, ExternalError.New("Недопустимый размер пароля. Пароль должен находится в диапозоне от 8 до 64 символов.", http.StatusBadRequest)
	}

	_, err := m.FindUserByEmail(email)

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

func (m Model) SoftDelete(uid string) error {
	user, err := m.FindUserByID(uid)

	if err != nil || (user == indexedUser{}) {
		return ExternalError.New("Пользователь не был найден", http.StatusNotFound)
	}

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	update := bson.D{{
		"$set", bson.D{{
			"deletedAt", util.UnixTimeNow(),
		}},
	}}

	_, err = m.collection.UpdateByID(ctx, DB.ObjectIDFromHex(uid), update)

	if err != nil {
		log.Println("[ ERROR ] Failed to update user (query error) \"" + uid + "\" - " + err.Error())

		return ExternalError.New("Внутренняя ошибка сервера", http.StatusInternalServerError)
	}

	return nil
}

func (m Model) FindUserByID(uid string) (indexedUser, error) {
	return m.findUserBy("_id", DB.ObjectIDFromHex(uid))
}

func (m Model) FindUserByEmail(email string) (indexedUser, error) {
	return m.findUserBy("email", email)
}

func (m Model) findUserBy(key string, value any) (indexedUser, error) {
	var user indexedUser

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	userFilter := bson.D{{key, value}, {"deletedAt", primitive.Null{}}}

	cur, err := m.collection.Find(ctx, userFilter)

	if err != nil {
		log.Fatalln(err)
	}

	if hasResult := cur.Next(ctx); !hasResult {
		return user, ExternalError.New("user not found", http.StatusNotFound)
	}

	err = cur.Decode(&user)

	if err != nil {
		log.Fatalln(err)
	}

	// This actually not a critical problem, cuz on finishing request processing goroutine will be terminated
	// and garbage collector should kill cursor, but idk how it will work in practice.
	if err := cur.Close(ctx); err != nil {
		log.Printf("[ ERROR ] Failed to close cursor. ID: %s, E-Mail:%s\n", user.ID, user.Email)

		// user will be non-empty, but error will still presence
		return user, err
	}

	return user, nil
}
