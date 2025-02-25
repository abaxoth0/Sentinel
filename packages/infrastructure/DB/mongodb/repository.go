package mongodb

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sentinel/packages/core/user"
	UserDTO "sentinel/packages/core/user/DTO"
	Error "sentinel/packages/errors"
	"sentinel/packages/infrastructure/auth/authorization"
	"sentinel/packages/infrastructure/cache"
	"sentinel/packages/util"
	"slices"

	emongo "github.com/StepanAnanin/EssentialMongoDB"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// TODO a lot of warnings on queries, need to check them

type repository struct {
	//
}

func (repo *repository) Create(login string, password string) (error) {
	if err := user.VerifyPassword(password); err != nil {
		return err
	}

	if _, err := driver.FindUserByLogin(login); err == nil {
		return Error.NewStatusError(
            "Пользователь с таким логином уже существует.",
            http.StatusConflict,
        )
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)

	if err != nil {
		return Error.StatusInternalError
    }

	data := user.Model{
		Login: login,
		Password: string(hashedPassword),
		Roles: []string{authorization.Host.OriginRoleName},
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	_, err = driver.UserCollection.InsertOne(ctx, data)

	if err != nil {
		return Error.StatusInternalError
	}

	return nil
}

func (repo *repository) update(filter *UserDTO.Filter, upd *primitive.E, deleted bool) *Error.Status {
	// TODO replace this shit via driver.FindAnyUserByID
	if deleted {
		if _, err := driver.FindSoftDeletedUserByID(filter.TargetUID); err != nil {
			return err
		}
	} else {
		if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
			return err
		}
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	update := bson.D{{"$set", bson.D{*upd}}}

    uid, err := primitive.ObjectIDFromHex(filter.TargetUID)

    if err != nil {
        fmt.Printf("failed to format uid to objectID: %v\n", err.Error())

        return Error.StatusInternalError
    }

	_, updError := driver.UserCollection.UpdateByID(ctx, uid, update)

	if updError != nil {
		log.Println("[ ERROR ] Failed to update user (query error) \"" + filter.TargetUID + "\" - " + updError.Error())

		return Error.StatusInternalError
    }

	cache.Delete(cache.UserKeyPrefix + filter.TargetUID)

	return nil
}

func (repo *repository) SoftDelete(filter *UserDTO.Filter) *Error.Status {
	// TODO add possibility to config what kind of users can delete themselves
    // all users can delete themselves, except admins (TEMP)
    if filter.TargetUID != filter.RequesterUID {
        if err := authorization.Authorize(authorization.Action.SoftDelete, authorization.Resource.User, filter.RequesterRoles); err != nil {
            return err
        }
    }

	target, err := driver.FindUserByID(filter.TargetUID);

    if err != nil {
		return Error.StatusUserNotFound
    }

    if slices.Contains(target.Roles, "admin") {
        return Error.NewStatusError("Нельзя удалить пользователя с ролью администратора", http.StatusBadRequest)
    }

	user, err := driver.FindUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	user.DeletedAt = int(util.UnixTimeNow())

    uid, e := primitive.ObjectIDFromHex(user.ID)

    if e != nil {
        fmt.Printf("failed to format uid to objectID: %v\n", err.Error())

        return Error.StatusInternalError
    }

    transferData := &emongo.TransferData{
        DocumentID: uid,
        Document: UserDTO.IndexedToUnindexed(user),
    }

	if e := emongo.DocumentTransfer(transferData, driver.UserCollection, driver.DeletedUserCollection, func() { user.DeletedAt = 0 }); e != nil {
		err = Error.NewStatusError(e.Error(), http.StatusInternalServerError)
	}

	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) Restore(filter *UserDTO.Filter) *Error.Status {
	if err := authorization.Authorize(authorization.Action.Restore, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
		return Error.StatusUserNotFound
    }

	user, err := driver.FindSoftDeletedUserByID(filter.TargetUID)

	if err != nil {
		return err
	}

	deletedAtTimestamp := user.DeletedAt
	user.DeletedAt = 0

    uid, e := primitive.ObjectIDFromHex(user.ID)

    if e != nil {
        fmt.Printf("failed to format uid to objectID: %v\n", err.Error())

        return Error.StatusInternalError
    }

    transferData := &emongo.TransferData{
        DocumentID: uid,
        Document: UserDTO.IndexedToUnindexed(user),
    }

	if err := emongo.DocumentTransfer(transferData, driver.DeletedUserCollection, driver.UserCollection, func() { user.DeletedAt = deletedAtTimestamp }); err != nil {
		return Error.NewStatusError(err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// Hard delete
func (repo *repository) Drop(filter *UserDTO.Filter) *Error.Status {
	if err := authorization.Authorize(authorization.Action.Drop, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
		return Error.StatusUserNotFound
    }

	user, err := driver.FindAnyUserByID(filter.TargetUID)
	deleted := user.DeletedAt == 0

	if err != nil {
		return err
	}

	ctx, cancel := emongo.DefaultTimeoutContext()

	defer cancel()

	collection := util.Ternary(deleted, driver.DeletedUserCollection, driver.UserCollection)

	if _, e := collection.DeleteOne(ctx, bson.D{{"_id", filter.TargetUID}}); e != nil {
        fmt.Printf("failed to delete user: %v\n", e.Error())

        return Error.NewStatusError(
            "Не удалось удалить пользователя",
            http.StatusInternalServerError,
        )
	}

	cacheKeyPrefix := util.Ternary(deleted, cache.DeletedUserKeyPrefix, cache.UserKeyPrefix)
	cache.Delete(cacheKeyPrefix + filter.TargetUID)

	return nil
}

func (repo *repository) DropAllSoftDeleted(requesterRoles []string) *Error.Status {
	if err := authorization.Authorize(authorization.Action.DropAllSoftDeleted, authorization.Resource.User, requesterRoles); err != nil {
		return err
	}

	_, err := driver.DeletedUserCollection.DeleteMany(context.TODO(), bson.D{})

	if err != nil {
		return Error.StatusInternalError
    }

	return nil
}

func (repo *repository) ChangeLogin(filter *UserDTO.Filter, newlogin string) *Error.Status {
	if err := authorization.Authorize(authorization.Action.ChangeLogin, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
		return Error.StatusUserNotFound
    }

	_, err := driver.FindUserByLogin(newlogin)

	// user with new login was found
	if err == nil {
		return Error.NewStatusError("Данный логин уже занят", http.StatusConflict)
	}

	upd := &primitive.E{"login", newlogin}

	return repo.update(filter, upd, true)
}

func (repo *repository) ChangePassword(filter *UserDTO.Filter, newPassword string) *Error.Status {
	if err := authorization.Authorize(authorization.Action.ChangePassword, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
		return Error.StatusUserNotFound
    }

	if err := user.VerifyPassword(newPassword); err != nil {
		return err
	}

	hashedPassword, e := bcrypt.GenerateFromPassword([]byte(newPassword), 12)

	if e != nil {
        fmt.Printf("failed to hash password: %v\n", e.Error())

		return Error.StatusInternalError
    }

	upd := &primitive.E{"password", hashedPassword}

	return repo.update(filter, upd, true)
}

func (repo *repository) ChangeRoles(filter *UserDTO.Filter, newRoles []string) *Error.Status {
	if err := authorization.Authorize(authorization.Action.ChangeRoles, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return err
	}

	if _, err := driver.FindUserByID(filter.TargetUID); err != nil {
		return Error.StatusUserNotFound
    }

	upd := &primitive.E{"roles", newRoles}

	return repo.update(filter, upd, true)
}

func (repo *repository) GetRoles(filter *UserDTO.Filter) ([]string, *Error.Status) {
	if err := authorization.Authorize(authorization.Action.GetRole, authorization.Resource.User, filter.RequesterRoles); err != nil {
		return []string{}, err
	}

	user, err := driver.FindUserByID(filter.TargetUID)

	if err != nil {
		return []string{}, Error.StatusUserNotFound
    }

	return user.Roles, nil
}

