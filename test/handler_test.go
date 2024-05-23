package test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"users/config"

	"github.com/brianvoe/gofakeit/v7"

	storage "users/internal/db"

	delivery "users/internal/user/infrastructure/delivery/http"
	"users/internal/user/infrastructure/dto"
	"users/internal/user/infrastructure/repository"

	slogger "users/pkg/logger"

	"gopkg.in/go-playground/assert.v1"
)

var (
	loginAdmin  string
	userdb      storage.InMemoryStorage
	authdb      storage.InMemoryStorage
	handler     delivery.UserHandler
	repo        repository.UserRepository
	admin       Admin
	user        User
	updatedUser UpdatedUser
	userid      dto.UserId
)

type Admin struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Admin    bool   `json:"admin"`
}

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Admin    bool   `json:"admin"`
}

type UpdatedUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Admin    bool   `json:"admin"`
}

func TestMain(m *testing.M) {
	setup()
	os.Exit(m.Run())

}

func setup() {
	config.LoadConfig()
	slogger.Logger = slogger.GetLogger()
	loginAdmin = base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	userdb, authdb = storage.InMemoryStorage{Storage: make(map[string][]byte)}, storage.InMemoryStorage{Storage: make(map[string][]byte)}

	repo = repository.NewBannerRepository(&userdb, &authdb)
	repo.CreateAdmin()
	handler = *delivery.NewUserHandler(repo)

	admin = Admin{
		Username: "Kayle",
		Email:    "hello@world.ru",
		Password: "fagsddf",
		Admin:    true,
	}
	user = User{
		Username: gofakeit.Name(),
		Email:    "Marley@world.ru",
		Password: "hhttrrr",
		Admin:    false,
	}
	updatedUser = UpdatedUser{
		Username: gofakeit.Name(),
		Email:    "Doe@world.ru",
		Password: "niceone",
		Admin:    false,
	}
}

func tearDown(id string) {
	handler.Store.DeleteUser(id)
}

func CreateUser(data []byte) *http.Response {
	req := httptest.NewRequest(http.MethodPost, "/user/", bytes.NewBuffer(data))
	req.SetBasicAuth("admin", "admin")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	return w.Result()
}

func TestCreateAdminUser(t *testing.T) {

	b, _ := json.Marshal(admin)
	res := CreateUser(b)

	assert.Equal(t, res.StatusCode, 201)

	json.NewDecoder(res.Body).Decode(&userid)
	tearDown(userid.Id)

}

func TestCreateNonAdminUser(t *testing.T) {

	b, _ := json.Marshal(user)

	res := CreateUser(b)
	assert.Equal(t, res.StatusCode, 201)
	json.NewDecoder(res.Body).Decode(&userid)

	tearDown(userid.Id)
}

func TestGetByNonAdminUser(t *testing.T) {
	// creating non admin user
	b, _ := json.Marshal(user)

	res := CreateUser(b)
	json.NewDecoder(res.Body).Decode(&userid)

	// testing by logging as non admin
	req := httptest.NewRequest(http.MethodGet, "/user/", nil)
	req.SetBasicAuth(user.Username, user.Password)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res = w.Result()
	assert.Equal(t, res.StatusCode, 200)
	tearDown(userid.Id)
}

func TestUpdateCreatedUser(t *testing.T) {

	b, _ := json.Marshal(user)

	req := httptest.NewRequest(http.MethodPost, "/user/", bytes.NewBuffer(b))
	req.SetBasicAuth("admin", "admin")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()

	json.NewDecoder(res.Body).Decode(&userid)

	b, _ = json.Marshal(updatedUser)
	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/user/%s", userid.Id), bytes.NewBuffer(b))
	req.SetBasicAuth("admin", "admin")

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res = w.Result()
	assert.Equal(t, res.StatusCode, 204)

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s", userid.Id), nil)
	req.SetBasicAuth("admin", "admin")

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res = w.Result()
	assert.Equal(t, res.StatusCode, 200)

	var user dto.ListUser
	json.NewDecoder(res.Body).Decode(&user)
	assert.Equal(t, user.Username, updatedUser.Username)

}

func TestDeleteCreatedUser(t *testing.T) {
	b, _ := json.Marshal(user)

	res := CreateUser(b)

	json.NewDecoder(res.Body).Decode(&userid)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/user/%s", userid.Id), nil)
	req.SetBasicAuth("admin", "admin")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res = w.Result()
	assert.Equal(t, res.StatusCode, 204)

}

func BenchmarkAdminCreation(b *testing.B) {
	b.Run("Endpoint: POST /user/", func(b *testing.B) {
		data, _ := json.Marshal(admin)

		req := httptest.NewRequest(http.MethodPost, "/user/", bytes.NewBuffer(data))
		req.SetBasicAuth("admin", "admin")

		w := httptest.NewRecorder()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			handler.ServeHTTP(w, req)
		}
	})
}

func BenchmarkGetUsers(b *testing.B) {
	b.Run("Endpoint: GET /user/", func(b *testing.B) {

		req := httptest.NewRequest(http.MethodGet, "/user/", nil)
		req.SetBasicAuth("admin", "admin")

		w := httptest.NewRecorder()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			handler.ServeHTTP(w, req)
		}
	})
}
