package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/instiq/dmobilizer/models"
	"github.com/instiq/dmobilizer/utils"

	// jwt "github.com/dgrijalva/jwt-go"
	"github.com/thedevsaddam/govalidator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// CreateAdmin Account
func (c Controller) CreateAdmin(db *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var reqData models.AdminSignUpValidator
		var admin models.Admin
		util := utils.Util{}
		var errorMessage string

		// connect to collection
		collection := db.Database("dmobilizer").Collection("admins")

		rules := govalidator.MapData{
			"firstName":       []string{"required", "between:3,10"},
			"lastName":        []string{"required", "between:3,10"},
			"email":           []string{"required", "min:4", "max:20", "email"},
			"password":        []string{"required", "min:6", "max:20"},
			"confirmPassword": []string{"required", "min:6", "max:20"},
		}

		opts := govalidator.Options{
			Request: req,
			Data:    &reqData,
			Rules:   rules,
		}

		v := govalidator.New(opts)
		e := v.ValidateJSON()
		// err := map[string]interface{}{"message": e}
		// fmt.Println(e) // your incoming JSON data in Go data struct

		if len(e) != 0 {
			// get the first error message returned
			for _, dd := range e {
				errorMessage = dd[0]
			}
			util.ResponseJSON(res, http.StatusBadRequest, errorMessage)
			return
		}

		// check if the password and confirm password are a match
		if reqData.Password != reqData.ConfirmPassword {
			errorMessage = "Password and confirm password wasn't a match"
			util.ResponseJSON(res, http.StatusBadRequest, errorMessage)
			return
		}

		// find all the data and return a cursor for all markets with the exchanges ID
		err := collection.FindOne(context.TODO(), bson.D{primitive.E{Key: "email", Value: reqData.Email}}).Decode(&admin)

		log.Println(err)

		// check for error
		if err == nil {
			util.ResponseJSON(res, http.StatusBadRequest, "Account already exist.")
			return
		}

		// hash our password
		hash, err := bcrypt.GenerateFromPassword([]byte(reqData.Password), 10)

		if err != nil {
			log.Println(err)
			util.ResponseJSON(res, http.StatusInternalServerError, "Something went wrong!")
			return
		}

		reqData.Password = string(hash)

		// insert into database
		result, err := collection.InsertOne(
			context.Background(),
			bson.D{
				primitive.E{Key: "firstName", Value: reqData.FirstName},
				primitive.E{Key: "lastName", Value: reqData.LastName},
				primitive.E{Key: "email", Value: reqData.Email},
				primitive.E{Key: "password", Value: reqData.Password},
				primitive.E{Key: "verified", Value: true},
				primitive.E{Key: "createdAt", Value: time.Now()},
			},
		)

		if err != nil {
			log.Println(err)
			util.ResponseJSON(res, http.StatusInternalServerError, "An error occured.")
		}

		if result != nil {
			util.ResponseJSON(res, http.StatusCreated, "Account successfully created.")
		}
	}
}
