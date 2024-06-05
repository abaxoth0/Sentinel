package user

import (
	"log"
	"net/http"
	"sentinel/packages/DB"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"
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

func (m *Model) SoftDelete(targetID string, requesterID string, requesterRole role.Role) error {
	user, err := m.search.FindUserByID(targetID)

	// There are 1 case in wich error will presence but user will be non-empty:
	// If failed to close db cursor. (See comment in "findUserBy" method for detatils)
	if err != nil {
		if isExternal, _ := ExternalError.Is(err); isExternal {
			return ExternalError.New("Пользователь не был найден", http.StatusNotFound)
		}

		log.Fatalln(err.Error())
	}

	// If user want to delete another user, not himself and hi doesn't have access to do that
	if targetID != requesterID {
		// Only moderators and administrators can delete other users.
		if requesterRole != role.Moderator ||
			requesterRole != role.Administrator ||
			// Moderator can't delete another moderator
			(requesterRole == role.Moderator && user.Role == role.Moderator) {
			return ExternalError.New("У вас недостаточно прав для выполнения данной операции", http.StatusForbidden)
		}

		// Administrators can't be deleted through app, only through direct DB query.
		if user.Role == role.Administrator {
			return ExternalError.New("Невозможно удалить пользователя с ролью администратора. (Обратитесь напрямую в базу данных)", http.StatusForbidden)
		}
	}

	ctx, cancel := DB.DefaultTimeoutContext()

	defer cancel()

	update := bson.D{{
		"$set", bson.D{{
			"deletedAt", util.UnixTimeNow(),
		}},
	}}

	_, err = m.collection.UpdateByID(ctx, DB.ObjectIDFromHex(targetID), update)

	if err != nil {
		log.Println("[ ERROR ] Failed to update user (query error) \"" + targetID + "\" - " + err.Error())

		return ExternalError.New("Внутренняя ошибка сервера", http.StatusInternalServerError)
	}

	return nil
}

func (m *Model) Restore(targetID string, requesterID string, requesterRole string) error {
	return nil
}

// func (m *Model) FindUserByID(uid string) (IndexedUser, error) {
// 	return m.findUserBy("_id", DB.ObjectIDFromHex(uid))
// }

// func (m *Model) FindUserByEmail(email string) (IndexedUser, error) {
// 	return m.findUserBy("email", email)
// }

// func (m *Model) findUserBy(key string, value any) (IndexedUser, error) {
// 	var user IndexedUser

// 	ctx, cancel := DB.DefaultTimeoutContext()

// 	defer cancel()

// 	userFilter := bson.D{{key, value}, {"deletedAt", primitive.Null{}}}

// 	cur, err := m.collection.Find(ctx, userFilter)

// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	if hasResult := cur.Next(ctx); !hasResult {
// 		return user, ExternalError.New("user not found", http.StatusNotFound)
// 	}

// 	err = cur.Decode(&user)

// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	// This actually not a critical problem, cuz on finishing request processing goroutine will be terminated
// 	// and garbage collector should kill cursor, but idk how it will work in practice.
// 	if err := cur.Close(ctx); err != nil {
// 		log.Printf("[ ERROR ] Failed to close cursor. ID: %s, E-Mail:%s\n", user.ID, user.Email)

// 		// user will be non-empty, but error will still presence
// 		return user, err
// 	}

// 	return user, nil
// }
