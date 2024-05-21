package repository

import (
	"encoding/json"
	"fmt"
	entity "users/internal/user/domain"
	"users/internal/user/infrastructure/dto"
	storage "users/pkg/db"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(user dto.CreateUser) (uuid string, err error)
	GetUserList(limit, offset int) []dto.ListUser
	UpdateUser(uuid string, user dto.UpdateUser)
	DeleteUser(uuid string)
	IfUserExist(uuid string) bool
	GetCredentialsByUsername(username string) (dto.AuthPermission, bool)
	GetUserById(uuid string) dto.ListUser
	CreateAdmin()
}

type UserRepo struct {
	userdb *storage.InMemoryStorage // profile Storage by id
	authdb *storage.InMemoryStorage // credentials Storage by username
}

func NewBannerRepository(userdb *storage.InMemoryStorage, authdb *storage.InMemoryStorage) UserRepository {
	return &UserRepo{userdb: userdb, authdb: authdb}
}

func (u *UserRepo) CreateUser(user dto.CreateUser) (uuid string, err error) {
	id := u.GenerateUUID()
	err = user.HashPassword()

	if err != nil {
		return "", err
	}

	db_user := user.ToStorageUser(id)
	b, _ := json.Marshal(db_user)

	u.userdb.Lock()
	u.userdb.Storage[id] = b
	u.userdb.Unlock()

	auth_user := dto.AuthPermission{Password: db_user.Password,
		Admin: db_user.Admin}

	b, _ = json.Marshal(auth_user)
	u.authdb.Lock()
	u.authdb.Storage[user.Username] = b
	u.authdb.Unlock()

	return id, nil

}

func (u *UserRepo) GetUserList(limit, offset int) []dto.ListUser {

	if offset != 0 && limit != 0 {

		res := make([]dto.ListUser, 0, limit)
		cnt := 0

		u.userdb.RLock()
		for _, v := range u.userdb.Storage {
			cnt++
			if cnt <= offset {
				continue
			}

			var user dto.ListUser
			json.Unmarshal(v, &user)
			res = append(res, user)

			if cnt == offset+limit {
				break
			}

		}
		u.userdb.RUnlock()
		return res

	}

	res := []dto.ListUser{}

	u.userdb.RLock()

	for _, v := range u.userdb.Storage {
		var user dto.ListUser
		json.Unmarshal(v, &user)
		res = append(res, user)
	}

	u.userdb.RUnlock()

	return res

}

func (u *UserRepo) UpdateUser(uuid string, user dto.UpdateUser) {
	var userToUpdate entity.User

	json.Unmarshal(u.userdb.Storage[uuid], &userToUpdate)

	fmt.Println("userToUpdate", userToUpdate)

	user.HashPassword()
	user.MakeUpdatedUser(&userToUpdate)

	fmt.Println("userToUpdate", userToUpdate)

	b, _ := json.Marshal(userToUpdate)
	u.userdb.Lock()
	u.userdb.Storage[uuid] = b
	u.userdb.Unlock()

	if user.Username != "" {
		a := dto.AuthPermission{Password: user.Password,
			Admin: user.Admin}

		b, _ = json.Marshal(a)
		u.authdb.Lock()
		delete(u.authdb.Storage, userToUpdate.Username)
		u.authdb.Storage[user.Username] = b
		u.authdb.Unlock()
	}

}

func (u *UserRepo) DeleteUser(uuid string) {

	var user entity.User

	u.userdb.Lock()

	b := u.userdb.Storage[uuid]
	json.Unmarshal(b, &user)

	delete(u.userdb.Storage, uuid)

	u.userdb.Unlock()

	u.authdb.Lock()
	delete(u.authdb.Storage, user.Username)
	u.authdb.Unlock()
}

func (u *UserRepo) IfUserExist(uuid string) bool {
	u.userdb.RLock()
	_, ok := u.userdb.Storage[uuid]
	u.userdb.RUnlock()
	return ok
}

func (u *UserRepo) GetUserById(uuid string) dto.ListUser {
	var user dto.ListUser

	u.userdb.RLock()

	b := u.userdb.Storage[uuid]
	json.Unmarshal(b, &user)

	u.userdb.RUnlock()

	return user
}

func (u *UserRepo) GetCredentialsByUsername(username string) (dto.AuthPermission, bool) {

	var authCredentials dto.AuthPermission
	u.authdb.RLock()
	defer u.authdb.RUnlock()
	b, ok := u.authdb.Storage[username]

	if !ok {
		return authCredentials, ok
	}
	json.Unmarshal(b, &authCredentials)

	return authCredentials, true
}

func (u *UserRepo) GenerateUUID() string {
	return uuid.New().String()
}

func (u *UserRepo) CreateAdmin() {
	user := dto.CreateUser{Email: "lol@test.ru",
		Username: "admin",
		Password: "admin",
		Admin:    true}

	id := u.GenerateUUID()
	user.HashPassword()

	db_user := user.ToStorageUser(id)
	b, _ := json.Marshal(db_user)

	u.userdb.Lock()
	u.userdb.Storage[id] = b
	u.userdb.Unlock()

	auth_user := dto.AuthPermission{Password: db_user.Password,
		Admin: db_user.Admin}

	b, _ = json.Marshal(auth_user)

	u.authdb.Lock()
	u.authdb.Storage[user.Username] = b
	u.authdb.Unlock()

}
