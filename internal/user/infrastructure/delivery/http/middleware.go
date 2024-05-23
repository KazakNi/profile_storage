package delivery

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"users/config"
	"users/internal/cookies"
	"users/internal/user/infrastructure/dto"
	"users/internal/user/infrastructure/repository"
	slogger "users/pkg/logger"
)

func AuthRequiredCheck(repo repository.UserRepository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		username, password, ok := r.BasicAuth()

		var loggingCredentials = dto.AuthPermission{Password: password}

		if ok {
			dbCredentials, ok := repo.GetCredentialsByUsername(username)

			if ok && dto.CheckPassword(loggingCredentials.Password, dbCredentials.Password) {
				if *dbCredentials.Admin {
					setRoleCookieHandler(w, "admin")
					next.ServeHTTP(w, r)
					return

				} else {
					setRoleCookieHandler(w, "user")
					next.ServeHTTP(w, r)
					return

				}

			}

		}
		slogger.Logger.Info("Unauthorized access", "username", username)
		w.Header().Set("WWW-Authenticate", `Basic realm="username/password", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func IsAdminCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, err := getCookieHandler(r)
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Error(w, "Unallowed action", http.StatusForbidden)
				return
			case errors.Is(err, cookies.ErrInvalidValue):
				http.Error(w, "Unallowed action", http.StatusForbidden)
				return
			default:
				log.Println(err)
				http.Error(w, "server error", http.StatusInternalServerError)
				return

			}

		}
		if role == "admin" {
			next.ServeHTTP(w, r)
			return
		} else {
			http.Error(w, "Unallowed action", http.StatusForbidden)
			return
		}

	})
}

func setRoleCookieHandler(w http.ResponseWriter, userRole string) {

	cookie := http.Cookie{
		Name:     "Role",
		Value:    userRole,
		Path:     "/",
		MaxAge:   3600 * 12,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	secretKey, err := hex.DecodeString(config.Cfg.Token.Secret)
	if err != nil {
		log.Fatal(err)
	}

	err = cookies.WriteSigned(w, &cookie, secretKey)
	if err != nil {
		log.Println(err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
}

func getCookieHandler(r *http.Request) (string, error) {

	secretKey, err := hex.DecodeString(config.Cfg.Token.Secret)
	if err != nil {
		log.Fatal(err)
	}
	value, err := cookies.ReadSigned(r, "Role", secretKey)

	if err != nil {
		return "", err
	}
	return value, nil
}

type ResponseWriterWrapper struct {
	w          *http.ResponseWriter
	body       *bytes.Buffer
	statusCode *int
}

func (rww ResponseWriterWrapper) String() string {
	var buf bytes.Buffer

	buf.WriteString("Response:")

	buf.WriteString("Headers:")
	for k, v := range (*rww.w).Header() {
		buf.WriteString(fmt.Sprintf("%s: %v", k, v))
	}

	buf.WriteString(fmt.Sprintf(" Status Code: %d", *(rww.statusCode)))

	buf.WriteString("Body")
	buf.WriteString(rww.body.String())
	return buf.String()
}
func (rww ResponseWriterWrapper) Write(buf []byte) (int, error) {
	rww.body.Write(buf)
	return (*rww.w).Write(buf)
}

func (rww ResponseWriterWrapper) Header() http.Header {
	return (*rww.w).Header()

}

func (rww ResponseWriterWrapper) WriteHeader(statusCode int) {
	(*rww.statusCode) = statusCode
	(*rww.w).WriteHeader(statusCode)
}
func NewResponseWriterWrapper(w http.ResponseWriter) ResponseWriterWrapper {
	var buf bytes.Buffer
	var statusCode int = 200
	return ResponseWriterWrapper{
		w:          &w,
		body:       &buf,
		statusCode: &statusCode,
	}
}

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slogger.Logger.Info("income request", "endpoint", r.URL, "method", r.Method, "rBody", r.Body)
		defer func() {
			rww := NewResponseWriterWrapper(w)
			slogger.Logger.Info("Response data", "Request", r, "Response", rww.String())
		}()
		next.ServeHTTP(w, r)

	})
}
