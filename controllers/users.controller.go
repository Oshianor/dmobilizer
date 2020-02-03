package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/instiq/dmobilizer/models"
	"github.com/instiq/dmobilizer/utils"
	"github.com/mitchellh/mapstructure"

	goContext "github.com/gorilla/context"
	"github.com/thedevsaddam/govalidator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"golang.org/x/crypto/bcrypt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Controller empty struct
type Controller struct{}

type tokenData struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Type  string `json:"type"`
}

// CreateAgents api to creat new agents
func (c Controller) CreateAgents(db *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var reqData models.UserSignUpValidator
		var user models.Users
		util := utils.Util{}
		var errorMessage string
		var tokenData tokenData

		// fmt.Println(req.FormValue("email"))
		reqData.Email = req.FormValue("email")
		reqData.FirstName = req.FormValue("firstName")
		reqData.LastName = req.FormValue("lastName")
		reqData.PhoneNumber = req.FormValue("phoneNumber")
		reqData.Type = req.FormValue("type")
		reqData.DOB = req.FormValue("dob")
		reqData.Password = req.FormValue("password")
		reqData.ConfirmPassword = req.FormValue("confirmPassword")
		reqData.MiddleName = req.FormValue("middleName")

		// connect to collection
		collection := db.Database("dmobilizer").Collection("users")

		data := goContext.Get(req, "user")

		mapstructure.Decode(data, &tokenData)

		rules := govalidator.MapData{
			"firstName":       []string{"required", "between:3,10"},
			"lastName":        []string{"required", "between:3,10"},
			"middleName":      []string{"required", "between:3,10"},
			"email":           []string{"required", "min:4", "max:20", "email"},
			"type":            []string{"required", "in:teller,agent"},
			"dob":             []string{"required"},
			"phoneNumber":     []string{"required", "len:11"},
			"password":        []string{"required", "min:6", "max:20"},
			"confirmPassword": []string{"required", "min:6", "max:20"},
		}

		opts := govalidator.Options{
			// Request: req,
			Data:            &reqData,
			Rules:           rules,
			RequiredDefault: true,
		}

		v := govalidator.New(opts)
		// e := v.ValidateJSON()
		e := v.ValidateStruct()

		log.Println(e)
		log.Println(reqData)
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

		// reqData.Type check that reqData type must be either teller or agent
		if reqData.Type != "teller" && reqData.Type != "agent" {
			errorMessage = "Account type must be either Teller or Agent"
			util.ResponseJSON(res, http.StatusBadRequest, errorMessage)
			return
		}

		// find all the data and return a cursor for all markets with the exchanges ID
		err := collection.FindOne(context.TODO(), bson.D{primitive.E{Key: "email", Value: reqData.Email}}).Decode(&user)

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

		// create database index
		indexName, err := collection.Indexes().CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys:    bsonx.Doc{{"email", bsonx.Int32(1)}},
				Options: options.Index().SetUnique(true),
			},
		)

		if err != nil {
			log.Println(err)
			util.ResponseJSON(res, http.StatusInternalServerError, "Something went wrong!")
			return
		}

		fmt.Println(indexName)

		// image upload for profile image
		maxSize := int64(20000000)

		imgERR := req.ParseMultipartForm(maxSize)
		if imgERR != nil {
			util.ResponseJSON(res, http.StatusBadRequest, "Image too large. Max Size should be less or equal to 10mb")
			return
		}

		// FormFile returns the first file for the given key `myFile`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file
		profileImageFile, profileImageFileHeader, err := req.FormFile("profileImage")
		if err != nil {
			util.ResponseJSON(res, http.StatusBadRequest, "Error Retrieving the File")
			return
		}
		defer profileImageFile.Close()

		IDFile, IDFileHeader, err := req.FormFile("idImage")
		if err != nil {
			util.ResponseJSON(res, http.StatusBadRequest, "Error Retrieving the File")
			return
		}
		defer IDFile.Close()

		// create an AWS session which can be
		// reused if we're uploading many files
		s, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(
				"AWS-ID",                     // id
				"AWS-SECRET", // secret
				""), // token can be left blank for now
		})
		if err != nil {
			// fmt.Fprintf(res, "Could not upload file")
			util.ResponseJSON(res, http.StatusBadRequest, "Could not upload file")
			return
		}

		profileImageFileName, err := util.UploadFileToS3(s, profileImageFile, profileImageFileHeader)
		if err != nil {
			util.ResponseJSON(res, http.StatusBadRequest, "Could not upload file")
			return
		}

		IDFileName, err := util.UploadFileToS3(s, IDFile, IDFileHeader)
		if err != nil {
			util.ResponseJSON(res, http.StatusBadRequest, "Could not upload file")
			return
		}

		// set password and hash string
		reqData.Password = string(hash)

		adminID, _ := primitive.ObjectIDFromHex(tokenData.ID)

		// get the total number of user records
		// agentID, err := collection.EstimatedDocumentCount(context.Background())
		// if err != nil {
		// 	util.ResponseJSON(res, http.StatusBadRequest, "Something went wrong")
		// 	return
		// }

		// insert into database
		result, err := collection.InsertOne(
			context.Background(),
			bson.D{
				primitive.E{Key: "adminId", Value: adminID},
				// primitive.E{Key: "agentId", Value: agentID + 1},
				primitive.E{Key: "firstName", Value: reqData.FirstName},
				primitive.E{Key: "lastName", Value: reqData.LastName},
				primitive.E{Key: "email", Value: reqData.Email},
				primitive.E{Key: "dob", Value: reqData.DOB},
				primitive.E{Key: "status", Value: true},
				primitive.E{Key: "idImage", Value: IDFileName},
				primitive.E{Key: "profileImage", Value: profileImageFileName},
				primitive.E{Key: "type", Value: reqData.Type},
				primitive.E{Key: "password", Value: reqData.Password},
				primitive.E{Key: "phoneNumber", Value: reqData.PhoneNumber},
				primitive.E{Key: "createdAt", Value: time.Now()},
			},
		)

		if err != nil {
			util.ResponseJSON(res, http.StatusInternalServerError, "An error occured.")
		}

		if result != nil {
			util.ResponseJSON(res, http.StatusCreated, "Account successfully created.")
		}

	}
}

// GetAllAgents api to creat new agents
func (c Controller) GetAllAgents(db *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		util := utils.Util{}

		// connect to collection
		collection := db.Database("dmobilizer").Collection("agents")

		// options
		options := options.Find()

		// Sort by `_id` field descending
		// options.SetSort(bson.D{{"_id", -1}})

		// Limit by 10 documents only
		options.SetLimit(2)

		// find all the data and return a cursor for all markets with the exchanges ID
		cursor, err := collection.Find(context.Background(), bson.D{{}}, options)

		if err != nil {
			res.Header().Set("Content-type", "application/json")
			res.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(res).Encode("An Error ocurred")
			return
		}

		var agents []*models.Users

		// Close the cursor once finished
		defer cursor.Close(context.TODO())

		// Finding multiple documents returns a cursor
		// Iterating through the cursor allows us to decode documents one at a time
		for cursor.Next(context.TODO()) {

			// create a value into which the single document can be decoded
			var doc models.Users
			err := cursor.Decode(&doc)
			if err != nil {
				res.Header().Set("Content-type", "application/json")
				res.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(res).Encode("An Error ocurred")
				return
			}

			// append single market
			agents = append(agents, &doc)
		}

		if err := cursor.Err(); err != nil {
			log.Println(err)
			util.ResponseJSON(res, http.StatusInternalServerError, "Something went wrong!")
			return
		}

		log.Println(agents)

		// if agents != nil {
		res.Header().Set("Content-type", "application/json")
		res.WriteHeader(http.StatusAccepted)
		json.NewEncoder(res).Encode(agents)
		return
		// }
	}
}

// Deposit money for customer
func (c Controller) Deposit(db *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var reqData models.Deposit
		// var customer models.Customer
		util := utils.Util{}
		var errorMessage string
		// var tokenData tokenData

		// // connect to collection
		// collection := db.Database("dmobilizer").Collection("customers")

		// num, _ := strconv.Atoi(vars["accountNumber"])

		rules := govalidator.MapData{
			"customerId": []string{"required"},
			"amount":     []string{"required", "numeric_between:1,50000"},
		}

		opts := govalidator.Options{
			Request:         req,
			Data:            &reqData,
			Rules:           rules,
			RequiredDefault: true,
		}

		v := govalidator.New(opts)
		e := v.ValidateJSON()

		log.Println(reqData)

		if len(e) != 0 {
			// get the first error message returned
			for _, dd := range e {
				errorMessage = dd[0]
			}
			util.ResponseJSON(res, http.StatusBadRequest, errorMessage)
			return
		}

	}
}
