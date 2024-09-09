package server

import (
	"log"
	"net/http"
	"os"

	"city-tags-api/internal/api_errors"
	"city-tags-api/internal/gcp"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	env = os.Getenv("ENV")
)

func getTokenAuth() *jwtauth.JWTAuth {
	var encKey string
	if env == "LOCAL" {
		encKey = os.Getenv("ENC_KEY")
	} else {
		sm := gcp.NewSecretManager()
		param := sm.GetSecret("city-tags-api-secret")
		var ok bool
		encKey, ok = param["ENC_KEY"]
		if !ok {
			log.Fatal("SSM Parameter doesn't have ENC_KEY key")
		}
	}
	return jwtauth.New("HS256", []byte(encKey), nil)
}

func Authenticator(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return NewHandler(
			func(w http.ResponseWriter, r *http.Request) error {
				token, _, err := jwtauth.FromContext(r.Context())

				if err != nil {
					return &api_errors.ClientErr{
						HttpCode: http.StatusUnauthorized,
						Message:  err.Error(),
					}
				}

				if token == nil || jwt.Validate(token, []jwt.ValidateOption{}...) != nil {
					return &api_errors.ClientErr{
						HttpCode: http.StatusUnauthorized,
						Message:  http.StatusText(http.StatusUnauthorized),
					}
				}

				next.ServeHTTP(w, r)
				return nil
			},
		)
	}
}
