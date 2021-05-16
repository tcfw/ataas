package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"pm.tcfw.com.au/source/trader/api/pb/passport"
)

var (
	passportSvc passport.PassportSeviceClient

	authWhitelistPrefixes []string = []string{
		`^\/v1\/auth\/(register|social|login)$`,
	}
)

//authHandler validates the http request for auth tokens and validates them
//against the passport service
func authHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken, checkCSRF := getAuthToken(r.Context(), r)

		if r.URL.Path == "/v1/gw/csrf" {
			generateCSRFTokenFromRequest(w, r)
			return
		}

		if r.Method != http.MethodGet && checkCSRF && r.Header.Get("X-GOBBLE") == "" {
			http.Error(w, "Missing turkeys", http.StatusTeapot)
			return
		}

		if shouldValidate(r) {
			if authToken == "" {
				http.Error(w, "Forbidden. No API Key provided", http.StatusForbidden)
				return
			}

			if passportSvc == nil {
				passportEndpoint, envExists := os.LookupEnv("PASSPORT_HOST")
				if !envExists {
					passportEndpoint = viper.GetString("grpc.addr")
				}
				conn, err := grpc.Dial(passportEndpoint, grpc.WithInsecure())
				if err != nil {
					http.Error(w, "Failed to validate token", http.StatusInternalServerError)
					panic(err)
				}

				passportSvc = passport.NewPassportSeviceClient(conn)
			}

			response, err := passportSvc.VerifyToken(r.Context(), &passport.VerifyTokenRequest{Token: authToken})
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to validate token: %s", err), http.StatusInternalServerError)
				return
			}
			if !response.Valid {
				http.Error(w, "Forbidden. Invalid API Key provided", http.StatusForbidden)
				return
			}

			if checkCSRF && r.Method == http.MethodPost {
				valid, err := validateCSRF(r)
				if !valid && err == nil {
					http.Error(w, "invalid CSRF token", http.StatusBadRequest)
				}
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to validate token: %s", err), http.StatusBadRequest)
					return
				}

			}
		}

		next.ServeHTTP(w, r)
	})
}

func shouldValidate(r *http.Request) bool {
	url := r.URL.String()

	for _, prefix := range authWhitelistPrefixes {
		if match, _ := regexp.MatchString(prefix, url); match {
			return false
		}
	}

	return true
}

func getAuthToken(ctx context.Context, r *http.Request) (string, bool) {
	if r.Header.Get("authorization") != "" {
		return r.Header.Get("authorization"), false
	}

	cookie, err := r.Cookie("session")
	if err == http.ErrNoCookie {
		r.Header.Set("authorization", "")
		return "", false
	}
	if err != nil {
		return "", true
	}

	r.Header.Set("authorization", cookie.Value)

	return cookie.Value, cookie.Value != ""
}
