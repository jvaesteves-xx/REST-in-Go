package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/auth0-community/auth0"
	jose "gopkg.in/square/go-jose.v2"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		secret := []byte(os.Getenv("API_SECRET"))
		secretProvider := auth0.NewKeyProvider(secret)
		audience := []string{os.Getenv("API_AUDIENCE")}

		configuration := auth0.NewConfiguration(secretProvider, audience, os.Getenv("API_ISSUER"), jose.HS256)
		validator := auth0.NewValidator(configuration)

		if token, err := validator.ValidateRequest(request); err != nil {
			RespondWithError(response, http.StatusUnauthorized, err.Error())
			fmt.Println(err)
			fmt.Println("Token is not valid:", token)
		} else {
			next.ServeHTTP(response, request)
		}
	})
}
