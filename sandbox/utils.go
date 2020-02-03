package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/instiq/dmobilizer/models"

	jwt "github.com/dgrijalva/jwt-go"
	// "go.mongodb.org/mongo-driver/bson/primitive"
)

// Util struct for exported functions
type Util struct{}

type Error struct {
	Message string `json:"message"`
}

type JWT struct {
	Token string `json:"token"`
}

// TokenVerifyMiddleWare verify
func (u Util) TokenVerifyMiddleWare(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errorObject Error
		authHeader := r.Header.Get("Authorization")
		bearerToken := strings.Split(authHeader, " ")

		if len(bearerToken) == 2 {
			authToken := bearerToken[1]

			token, error := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}

				return []byte("secret"), nil
			})

			if error != nil {
				errorObject.Message = error.Error()
				// respondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}

			if token.Valid {
				next.ServeHTTP(w, r)
			} else {
				errorObject.Message = error.Error()
				// respondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}
		} else {
			errorObject.Message = "Invalid token."
			// respondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}
	})
}

// GenerateAdminToken for login
func (u Util) GenerateAdminToken(admin models.Admin) (string, error) {
	var err error
	secret := "@#$%rcvsys^&*hdjnd"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    admin.ObjectID,
		"email": admin.Email,
		"type":  "admin",
	})

	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		log.Fatal(err)
	}

	return tokenString, nil
}


// ResponseJSON for successful response to the client
func (u Util) ResponseJSON(res http.ResponseWriter, status int, data interface{}) {
	res.Header().Set("Content-type", "application/json")
	res.WriteHeader(status)
	json.NewEncoder(res).Encode(data)
}





// type User struct {
// 	ObjectID primitive.ObjectID `json:"_id"`
// 	Email    string             `json:"email"`
// 	Type     string             `json:"type"`
// }


// // RespondWithError for error response
// func (u Util) RespondWithError(res http.ResponseWriter, status int, error Error) {
// 	res.Header().Set("Content-type", "application/json")
// 	res.WriteHeader(status)
// 	json.NewEncoder(res).Encode(error)
// }
