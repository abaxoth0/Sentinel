package auth

import (
	"net/http"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/search"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

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

// Returns indexedUser if auth data is correct, ExternalError otherwise.
func (m *Model) Login(login string, password string) (*search.IndexedUser, *ExternalError.Error) {
	user, err := m.search.FindUserByLogin(login)

	// If user was found (user != indexedUser{}) and there are error, that means cursor closing failed. (see `findUserBy` method)
	// If user wasn't found and there are error, that means occured an unexpected error.
	if err != nil && (*user != search.IndexedUser{}) {
		return user, ExternalError.New("Неверный логин или пароль", http.StatusBadRequest)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// We know that on this stage incorrect only password,
		// but there are no point to tell user about this, due to security reasons.
		return user, ExternalError.New("Неверный логин или пароль", http.StatusBadRequest)
	}

	return user, nil
}
