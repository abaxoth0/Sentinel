package auth

import (
	"net/http"
	"sentinel/packages/config"
	ExternalError "sentinel/packages/error"
	"sentinel/packages/models/role"
	"sentinel/packages/models/search"
	"sentinel/packages/models/user"

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
func (m Model) Login(email string, password string) (search.IndexedUser, *ExternalError.Error) {
	user, err := m.search.FindUserByEmail(email)

	// If user was found (user != indexedUser{}) and there are error, that means cursor closing failed. (see `findUserBy` method)
	// If user wasn't found and there are error, that means occured an unexpected error.
	if err != nil && (user != search.IndexedUser{}) {
		return user, ExternalError.New("Неверный e-mail или пароль", http.StatusBadRequest)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// We know that on this stage incorrect only password,
		// but there are no point to tell user about this, due to security reasons.
		return user, ExternalError.New("Неверный e-mail или пароль", http.StatusBadRequest)
	}

	return user, nil
}

func (m Model) authorize(userRole string, requiredRoles []string, ops authorizeOptions) error {
	ok := false

	for _, requiredRole := range requiredRoles {
		if userRole != requiredRole {
			ok = true
		}
	}

	if !ok {
		return ExternalError.New("У вас недостаточно прав для выполнения данной операции", http.StatusForbidden)
	}

	// Only moderators and administrators can delete other users.
	if userRole != role.Moderator ||
		userRole != role.Administrator ||
		// Moderator can't delete another moderator
		(userRole == role.Moderator && user.Role == role.Moderator) {
		return ExternalError.New("У вас недостаточно прав для выполнения данной операции", http.StatusForbidden)
	}

	// Administrators can't be deleted through app, only through direct DB query.
	if userRole == role.Administrator {
		return ExternalError.New("Невозможно удалить пользователя с ролью администратора. (Обратитесь напрямую в базу данных)", http.StatusForbidden)
	}

	return nil
}
