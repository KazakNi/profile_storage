package repository

import (
	"encoding/json"
	storage "users/internal/db"
	entity "users/internal/user/domain"
	"users/internal/user/infrastructure/dto"

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

	u.userdb.Set(id, b)

	auth_user := dto.AuthPermission{Password: db_user.Password,
		Admin: db_user.Admin}

	b, _ = json.Marshal(auth_user)
	u.authdb.Set(user.Username, b)

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
	var db_user dto.ListUser

	users := u.userdb.GetUsers()

	for _, user := range users {
		json.Unmarshal(user, &db_user)
		res = append(res, db_user)
	}

	return res

}

func (u *UserRepo) UpdateUser(uuid string, user dto.UpdateUser) {
	var userToUpdate entity.User

	json.Unmarshal(u.userdb.Storage[uuid], &userToUpdate)

	user.HashPassword()
	user.MakeUpdatedUser(&userToUpdate)

	b, _ := json.Marshal(userToUpdate)
	u.userdb.Set(uuid, b)

	if user.Username != "" {
		a := dto.AuthPermission{Password: user.Password,
			Admin: user.Admin}

		b, _ = json.Marshal(a)
		u.authdb.Delete(userToUpdate.Username)
		u.authdb.Set(user.Username, b)
	}

}

func (u *UserRepo) DeleteUser(uuid string) {

	var user entity.User

	b, _ := u.userdb.Get(uuid)
	json.Unmarshal(b, &user)

	u.userdb.Delete(uuid)
	u.authdb.Delete(user.Username)
}

func (u *UserRepo) IfUserExist(uuid string) bool {
	_, ok := u.userdb.Get(uuid)
	return ok
}

func (u *UserRepo) GetUserById(uuid string) dto.ListUser {
	var user dto.ListUser
	b, _ := u.userdb.Get(uuid)
	json.Unmarshal(b, &user)
	return user
}

func (u *UserRepo) GetCredentialsByUsername(username string) (dto.AuthPermission, bool) {

	var authCredentials dto.AuthPermission
	b, ok := u.authdb.Get(username)

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
	admin := true
	user := dto.CreateUser{Email: "lol@test.ru",
		Username: "admin",
		Password: "admin",
		Admin:    &admin}

	id := u.GenerateUUID()
	user.HashPassword()

	db_user := user.ToStorageUser(id)
	b, _ := json.Marshal(db_user)

	u.userdb.Set(id, b)

	auth_user := dto.AuthPermission{Password: db_user.Password,
		Admin: db_user.Admin}

	b, _ = json.Marshal(auth_user)

	u.authdb.Set(user.Username, b)
}
