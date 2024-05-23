package delivery

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"users/config"
	"users/internal/user/infrastructure/dto"
	"users/internal/user/infrastructure/repository"
	slogger "users/pkg/logger"
)

var (
	UserRe       = regexp.MustCompile(`^/user/\w*$`)
	UserReWithID = regexp.MustCompile(`^/user/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
)

type UserHandler struct {
	Store repository.UserRepository
}

func (u *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && UserRe.MatchString(r.URL.Path):
		LogRequest(AuthRequiredCheck(u.Store, IsAdminCheck(http.HandlerFunc(u.CreateUser)))).ServeHTTP(w, r)
		return

	case r.Method == http.MethodGet && UserRe.MatchString(r.URL.Path):
		LogRequest(AuthRequiredCheck(u.Store, http.HandlerFunc(u.ListUser))).ServeHTTP(w, r)
		return

	case r.Method == http.MethodGet && UserReWithID.MatchString(r.URL.Path):
		LogRequest(AuthRequiredCheck(u.Store, http.HandlerFunc(u.GetUser))).ServeHTTP(w, r)
		return

	case r.Method == http.MethodPatch && UserReWithID.MatchString(r.URL.Path):
		LogRequest(AuthRequiredCheck(u.Store, IsAdminCheck(http.HandlerFunc(u.UpdateUser)))).ServeHTTP(w, r)
		return

	case r.Method == http.MethodDelete && UserReWithID.MatchString(r.URL.Path):
		LogRequest(AuthRequiredCheck(u.Store, IsAdminCheck(http.HandlerFunc(u.DeleteUser)))).ServeHTTP(w, r)
		return

	default:
		NotFoundHandler(w, r)
		return
	}
}

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	user := &dto.CreateUser{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	if err := user.Validate(); err != nil {
		BadRequestHandler(w, r)
		slogger.Logger.Info("error while user creation validation: %s", err)
		return
	}

	if _, ok := u.Store.GetCredentialsByUsername(user.Username); ok {
		slogger.Logger.Info("username miss while creating", "username:", user.Username)
		AlreadyExistsHandler(w, r)
		return
	}

	id, err := u.Store.CreateUser(*user)

	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	StatusCreatedHandler(w, r, id)
}

func (u *UserHandler) ListUser(w http.ResponseWriter, r *http.Request) {
	var limit_value, offset_value int

	params := r.URL.Query()

	limit, offset := params.Get("limit"), params.Get("offset")

	if limit == "" || offset == "" {

		users := u.Store.GetUserList(limit_value, offset_value)
		StatusListUserHandler(w, r, users)
		return

	} else {

		limit_value, err := strconv.Atoi(limit)
		if err != nil {
			BadRequestHandler(w, r)
			slogger.Logger.Info("error params validation of ListUser", "err", err)
			return
		}

		offset_value, err = strconv.Atoi(offset)
		if err != nil {
			BadRequestHandler(w, r)
			slogger.Logger.Info("error params validation of ListUser", "err", err)
			return
		}

		users := u.Store.GetUserList(limit_value, offset_value)
		StatusListUserHandler(w, r, users)
	}

}

func (u *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {

	user := &dto.UpdateUser{}

	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		slogger.Logger.Info("error while UpdateUser decoding: %s", err)
		InternalServerErrorHandler(w, r)
		return
	}

	if err := user.Validate(); err != nil {
		slogger.Logger.Info("error while UpdateUser validation: %s", err)
		BadRequestHandler(w, r)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/user/")

	if ok := u.Store.IfUserExist(id); !ok {
		NotFoundHandler(w, r)
		return
	}

	_, ok := u.Store.GetCredentialsByUsername(id)

	if u.Store.GetUserById(id).Username != user.Username && ok {
		slogger.Logger.Info("username already exists", "username:", user.Username)
		AlreadyExistsHandler(w, r)
		return
	}

	u.Store.UpdateUser(id, *user)

	w.WriteHeader(http.StatusNoContent)

}

func (u *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/user/")

	if ok := u.Store.IfUserExist(id); !ok {
		NotFoundHandler(w, r)
		return
	}

	u.Store.DeleteUser(id)

	w.WriteHeader(http.StatusNoContent)

}

func (u *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/user/")

	if ok := u.Store.IfUserExist(id); !ok {
		NotFoundHandler(w, r)
		return
	}

	user := u.Store.GetUserById(id)

	StatusOkContent(w, r, user)
}

func NewUserHandler(s repository.UserRepository) *UserHandler {
	return &UserHandler{
		Store: s,
	}
}

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	b, _ := json.Marshal(dto.ErrorResponse{Error: "500 Internal Server Error"})
	w.Write([]byte(b))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	b, _ := json.Marshal(`{"Error": "404 Not Found"}`)
	w.Write([]byte(b))
}

func AlreadyExistsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	b, _ := json.Marshal(`{"Error": "409 Conflict", "message": "username already exists"}`)
	w.Write([]byte(b))
}
func BadRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	b, _ := json.Marshal(dto.ErrorResponse{Error: "400 Bad request"})
	w.Write([]byte(b))
}

func StatusCreatedHandler(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	b, _ := json.Marshal(dto.UserId{Id: id})
	w.Write([]byte(b))
}

func StatusListUserHandler(w http.ResponseWriter, r *http.Request, users []dto.ListUser) {
	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(users)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func StatusOkContent(w http.ResponseWriter, r *http.Request, user dto.ListUser) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	b, _ := json.Marshal(user)
	w.Write(b)
}

func ReDoc(w http.ResponseWriter, r *http.Request) {
	static_path := config.Cfg.Swagger.HtmlPath
	tmpl, err := template.ParseFiles(static_path)
	if err != nil {
		log.Fatal(err)
	}
	tmpl.Execute(w, nil)
}
