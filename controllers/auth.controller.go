package controllers

import (
	"context"
	"log"
	"net/http"

	"github.com/instiq/dmobilizer/models"
	"github.com/instiq/dmobilizer/utils"

	"github.com/thedevsaddam/govalidator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// LoginAdmin to login as an admin
func (c Controller) LoginAdmin(db *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var loginAdmin models.AdminSignInValidator
		util := utils.Util{}
		var errorMessage string

		rules := govalidator.MapData{
			"email":    []string{"required", "min:4", "max:20", "email"},
			"password": []string{"required", "min:6", "max:20"},
		}

		opts := govalidator.Options{
			Request: req,
			Data:    &loginAdmin,
			Rules:   rules,
		}

		v := govalidator.New(opts)
		e := v.ValidateJSON()

		if len(e) != 0 {
			// get the first error message returned
			for _, dd := range e {
				errorMessage = dd[0]
			}
			util.ResponseJSON(res, http.StatusBadRequest, errorMessage)
			return
		}

		// connect to collection
		collection := db.Database("dmobilizer").Collection("admins")

		// options
		// options := options.FindOne()

		var admin models.Admin

		// find all the data and return a cursor for all markets with the exchanges ID
		err := collection.FindOne(context.TODO(), bson.D{primitive.E{Key: "email", Value: loginAdmin.Email}}).Decode(&admin)

		// check for error
		if err != nil {
			log.Println(err)
			util.ResponseJSON(res, http.StatusBadRequest, "Invalid email or password.")
			return
		}

		// if admin.Email == "" {
		// 	util.ResponseJSON(res, http.StatusBadRequest, "Invalid email or password.")
		// 	return
		// }

		hashedPassword := admin.Password
		password := loginAdmin.Password

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

		if err != nil {
			log.Println(err)
			util.ResponseJSON(res, http.StatusBadRequest, "Invalid email or password")
			return
		}

		token, err := util.GenerateAdminToken(admin)

		if err != nil {
			util.ResponseJSON(res, http.StatusBadRequest, "Something Went wrong!")
			return
		}

		res.Header().Set("AuthToken", token)
		util.ResponseJSON(res, http.StatusOK, "Successfully Logged In")
		return
	}
}

// LoginUser to login as an admin
func (c Controller) LoginUser(db *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var loginUser models.UserSignInValidator
		util := utils.Util{}
		var errorMessage string

		rules := govalidator.MapData{
			"email":    []string{"required", "min:4", "max:20", "email"},
			"password": []string{"required", "min:6", "max:20"},
		}

		opts := govalidator.Options{
			Request: req,
			Data:    &loginUser,
			Rules:   rules,
		}

		v := govalidator.New(opts)
		e := v.ValidateJSON()

		if len(e) != 0 {
			// get the first error message returned
			for _, dd := range e {
				errorMessage = dd[0]
			}
			util.ResponseJSON(res, http.StatusBadRequest, errorMessage)
			return
		}

		// connect to collection
		collection := db.Database("dmobilizer").Collection("users")

		// options
		// options := options.FindOne()

		var user models.Users

		// find all the data and return a cursor for all markets with the exchanges ID
		err := collection.FindOne(context.TODO(), bson.D{primitive.E{Key: "email", Value: loginUser.Email}}).Decode(&user)

		// check for error
		if err != nil {
			log.Println(err)
			util.ResponseJSON(res, http.StatusBadRequest, "Invalid email or password.")
			return
		}

		// if user.Email == "" {
		// 	util.ResponseJSON(res, http.StatusBadRequest, "Invalid email or password.")
		// 	return
		// }

		// check if the account is active
		if !user.Status {
			util.ResponseJSON(res, http.StatusBadRequest, "This account has been deactivated")
			return
		}

		hashedPassword := user.Password
		password := loginUser.Password

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

		if err != nil {
			util.ResponseJSON(res, http.StatusBadRequest, "Invalid email or password")
			return
		}
		// util.ResponseJSON(res, http.StatusBadRequest, admin)
		// return

		token, err := util.GenerateUserToken(user)

		if err != nil {
			log.Println(err)
			util.ResponseJSON(res, http.StatusInternalServerError, "Something went wrong!")
			return
		}

		res.Header().Set("AuthToken", token)
		util.ResponseJSON(res, http.StatusOK, "Successfully Logged In")
		return
	}
}
