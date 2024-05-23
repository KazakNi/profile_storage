package dto

import (
	"users/config"
	entity "users/internal/user/domain"

	"dario.cat/mergo"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var falseValue = false

type CreateUser struct {
	Username string `json:"username" validate:"required,max=150"`
	Email    string `json:"email" validate:"required,email,max=150"`
	Password string `json:"password" validate:"required,alphanumunicode,max=100"`
	Admin    *bool  `json:"admin" validate:"required,boolean"`
}

func (c *CreateUser) ToStorageUser(id string) entity.User {
	return entity.User{Id: id,
		Username: c.Username,
		Email:    c.Email,
		Password: c.Password,
		Admin:    c.Admin}
}

func (c *CreateUser) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)

	if err != nil {
		return err
	}
	return nil
}

func (c *CreateUser) HashPassword() error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(c.Password+config.Cfg.Token.Salt), 4)
	if err != nil {
		return err
	}
	c.Password = string(bytes)
	return nil
}

type UpdateUser struct {
	Username string `json:"username,omitempty" validate:"omitempty,max=150"`
	Email    string `json:"email,omitempty" validate:"omitempty,email,max=150"`
	Password string `json:"password,omitempty" validate:"omitempty,alphanumunicode,max=100"`
	Admin    *bool  `json:"admin,omitempty" validate:"omitempty,boolean"`
}

func (u *UpdateUser) Validate() error {
	validate := validator.New()
	err := validate.Struct(u)

	if err != nil {
		return err
	}
	return nil
}

func (u *UpdateUser) MakeUpdatedUser(userToUpdate *entity.User) {
	if u.Admin == nil {
		u.Admin = &falseValue
	}
	updatedEntity := entity.User{
		Id:       userToUpdate.Id,
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
		Admin:    u.Admin,
	}
	mergo.Merge(userToUpdate, updatedEntity, mergo.WithOverride, mergo.WithoutDereference)

}

func (u *UpdateUser) HashPassword() error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(u.Password+config.Cfg.Token.Salt), 4)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

type UserId struct {
	Id string `json:"id"`
}

type AuthPermission struct {
	Password string `json:"password"`
	Admin    *bool  `json:"admin"`
}

func (a *AuthPermission) HashPassword() error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(a.Password+config.Cfg.Token.Salt), 4)
	if err != nil {
		return err
	}
	a.Password = string(bytes)
	return nil
}

type ListUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Admin    bool   `json:"admin"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func CheckPassword(providedPassword string, db_password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(db_password), []byte(providedPassword+config.Cfg.Token.Salt))

	return err == nil
}
