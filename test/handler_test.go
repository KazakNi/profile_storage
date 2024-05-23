package test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"users/config"
	storage "users/internal/db"
	delivery "users/internal/user/infrastructure/delivery/http"
	"users/internal/user/infrastructure/repository"
	slogger "users/pkg/logger"

	"gopkg.in/go-playground/assert.v1"
)

var (
	loginAdmin string
	userdb     storage.InMemoryStorage
	authdb     storage.InMemoryStorage
	handler    delivery.UserHandler
	repo       repository.UserRepository
	admin      Admin
)

type Admin struct {
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
}

func TestCreateAdminUser(t *testing.T) {

	b, _ := json.Marshal(admin)

	req := httptest.NewRequest(http.MethodPost, "/user/", bytes.NewBuffer(b))
	req.SetBasicAuth("admin", "admin")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	res := w.Result()
	assert.Equal(t, res.StatusCode, 201)
}
