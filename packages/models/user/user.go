package user

import (
	"log"
	"net/http"
	"sentinel/packages/DB"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/role"

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

// Returns indexedUser if auth data is correct, ExternalError otherwise.
func (m Model) Login(email string, password string) (indexedUser, *ExternalError.ExternalError) {
	user, err := m.FindUserByEmail(email)

	if err != nil {
		isUserFound := user != indexedUser{}

		if ok, _ := ExternalError.Is(err); ok && isUserFound {
			// Invalid email or password, currently we know only about email,
			// but there are no point to tell user about this, due to security reasons.
			return user, ExternalError.New("Неверный e-mail или пароль", http.StatusBadRequest)
		}

		// If user was found and there are error, that means cursor closing failed. (see `FindUserByEmail` method)
		// This actually not a critical problem, cuz on finishing request processing goroutine will be terminated
		// and garbage collector should kill connection, but idk how it will work in practice.
		// P.S. Log for this case perfoms in `FindUserByEmail` method.
		// If user wasn't found and there are error, that means occured an unexpected error.
		if !isUserFound {
			log.Fatalln(err)
		}
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

func (m Model) FindUserByEmail(email string) (indexedUser, error) {
	var user indexedUser

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	userFilter := bson.D{{"email", email}, {"deletedAt", nil}}

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

	// user will be non-empty, but error will still presence
	if err := cur.Close(ctx); err != nil {
		log.Printf("[ ERROR ] Failed to close cursor. ID: %s, E-Mail:%s\n", user.ID, user.Email)

		return user, err
	}

	return user, nil
}
