package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/instiq/dmobilizer/models"
	"github.com/mitchellh/mapstructure"

	// "github.com/mitchellh/mapstructure"
	"github.com/gorilla/context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Util struct for exported functions
type Util struct{}

var secret string = "@#$%rcvsys^&*hdjnd"

// TokenData jwt data values
type tokenData struct {
	ID             string         `json:"id"`
	Email          string         `json:"email"`
	Type           string         `json:"type"`
	StandardClaims standardClaims `json:"StandardClaims"`
}

type standardClaims struct {
	Exp int64 `json:"exp"`
}

// TokenVerifyMiddleWare verify
func (u Util) TokenVerifyMiddleWare(next http.HandlerFunc, accountType string) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// var errorObject Error
		authToken := req.Header.Get("AuthToken")
		if authToken == "" {
			u.ResponseJSON(res, http.StatusForbidden, "No token was provided.")
			return
		}

		// log.Println(authHeader)
		token, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error")
			}

			return []byte(secret), nil
		})

		var values tokenData

		// decode the token claims to struct
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			mapstructure.Decode(claims, &values)
			// json.NewEncoder(res).Encode(values)
		} else {
			u.ResponseJSON(res, http.StatusForbidden, "Invalid token data")
			return
		}

		if values.Type != accountType {
			u.ResponseJSON(res, http.StatusForbidden, "You don't have access to this route")
			return
		}

		// convert int64 to unix time
		// var tm time.Time = time.Unix(values.StandardClaims.Exp, 0)

		// // log.Println(time.Now().Unix())
		// // log.Println(tm)

		// if time.Now().Unix() > tm {

		// }

		if err != nil {
			u.ResponseJSON(res, http.StatusUnauthorized, err.Error())
			return
		}

		// log.Println(values)

		if token.Valid {
			// ctx := context.WithValue(r.Context(), "user", values)
			// next.ServeHTTP(res, req.WithContext(ctx))
			context.Set(req, "user", values)
			next.ServeHTTP(res, req)
		} else {
			u.ResponseJSON(res, http.StatusForbidden, err.Error())
			return
		}
	})
}

// GenerateAdminToken for login
func (u Util) GenerateAdminToken(admin models.Admin) (string, error) {
	var err error

	expirationTime := time.Now().Add(2 * time.Minute)
	// expirationTime := time.Now().Add(36000 * time.Minute)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    admin.ObjectID,
		"email": admin.Email,
		"type":  "admin",
		"StandardClaims": jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	})

	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		log.Println(err)
		return "", err
	}

	return tokenString, nil
}

// GenerateUserToken for login
func (u Util) GenerateUserToken(user models.Users) (string, error) {
	var err error

	// expirationTime := time.Now().Add(2 * time.Minute)
	expirationTime := time.Now().Add(36000 * time.Minute)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      user.ObjectID,
		"email":   user.Email,
		"type":    user.Type,
		"adminId": user.AdminID,
		"StandardClaims": jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	})

	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		log.Println(err)
		return "", err
	}

	return tokenString, nil
}

// ResponseJSON for successful response to the client
func (u Util) ResponseJSON(res http.ResponseWriter, status int, data interface{}) {
	res.Header().Set("Content-type", "application/json")
	res.WriteHeader(status)
	json.NewEncoder(res).Encode(data)
}

// UploadFileToS3 saves a file to aws bucket and returns the url to // the file and an error if there's any
func (u Util) UploadFileToS3(s *session.Session, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	fmt.Println("started")
	// get the file size and read
	// the file content into a buffer
	size := fileHeader.Size
	buffer := make([]byte, size)
	file.Read(buffer)

	// create a unique file name for the file
	tempFileName := "pictures/" + primitive.NewObjectID().Hex() + filepath.Ext(fileHeader.Filename)

	log.Println("tempFileName", tempFileName)
	// config settings: this is where you choose the bucket,
	// filename, content-type and storage class of the file
	// you're uploading
	data, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:        aws.String("churchee-app-storage"),
		Key:           aws.String(tempFileName),
		ACL:           aws.String("public-read"), // could be private if you want it to be access by only authorized users
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(int64(size)),
		ContentType:   aws.String(http.DetectContentType(buffer)),
		// ContentDisposition:   aws.String("attachment"),
		// ServerSideEncryption: aws.String("AES256"),
		// StorageClass:         aws.String("INTELLIGENT_TIERING"),
	})

	log.Println("data", data)
	log.Println("err", err)

	if err != nil {
		return "", err
	}

	return tempFileName, err
}

// RandomGen to generate random numbers
func (u Util) RandomGen(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
