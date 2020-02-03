package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"

	goContext "github.com/gorilla/context"
	"github.com/instiq/dmobilizer/models"
	"github.com/instiq/dmobilizer/utils"
	"github.com/thedevsaddam/govalidator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// CreateCustomers account
func (c Controller) CreateCustomers(db *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var reqData models.CustomerSignUpValidatior
		var customer models.Customer
		util := utils.Util{}
		var errorMessage string
		var tokenData tokenData

		reqData.Email = req.FormValue("email")
		reqData.FirstName = req.FormValue("firstName")
		reqData.LastName = req.FormValue("lastName")
		reqData.PhoneNumber = req.FormValue("phoneNumber")
		reqData.BVN = req.FormValue("bvn")
		reqData.DOB = req.FormValue("dob")
		reqData.Address = req.FormValue("address")
		reqData.Gender = req.FormValue("gender")
		reqData.MiddleName = req.FormValue("middleName")

		// connect to collection
		collection := db.Database("dmobilizer").Collection("customers")

		// Declare Context type object for managing multiple API requests
		ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

		data := goContext.Get(req, "user")

		mapstructure.Decode(data, &tokenData)

		rules := govalidator.MapData{
			"email":       []string{"min:4", "max:20", "email"},
			"firstName":   []string{"required", "between:3,10"},
			"lastName":    []string{"required", "between:3,10"},
			"middleName":  []string{"required", "between:3,10"},
			"dob":         []string{"required"},
			"bvn":         []string{"min:6", "max:15"},
			"phoneNumber": []string{"required", "len:11"},
			"address":     []string{"required", "between:3,100"},
			"gender":      []string{"required", "in:male,female,others"},
			// "profileImage": []string{"required", "mime:jpg,png", "ext:jpg,png,jpeg"},
		}

		opts := govalidator.Options{
			Data:  &reqData,
			Rules: rules,
			// RequiredDefault: true,
		}

		v := govalidator.New(opts)
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

		err := collection.FindOne(context.TODO(), bson.D{primitive.E{Key: "phoneNumber", Value: reqData.PhoneNumber}}).Decode(&customer)

		log.Println("find doc", err)

		// check for error
		if err == nil {
			util.ResponseJSON(res, http.StatusBadRequest, "Account already exist.")
			return
		}

		// create multiple index for customer collection
		_, err = collection.Indexes().CreateMany(
			ctx,
			[]mongo.IndexModel{
				{
					Keys:    bsonx.Doc{{Key: "phoneNumber", Value: bsonx.Int32(1)}},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys:    bsonx.Doc{{Key: "accountNumber", Value: bsonx.Int32(1)}},
					Options: options.Index().SetUnique(true),
				},
			},
			options.CreateIndexes().SetMaxTime(10*time.Second),
		)
		if err != nil {
			log.Println(err)
			util.ResponseJSON(res, http.StatusInternalServerError, "Something went wrong!")
			return
		}

		// image upload for profile image
		maxSize := int64(20000000)

		imageErr := req.ParseMultipartForm(maxSize)
		if imageErr != nil {
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

		// create an AWS session which can be
		// reused if we're uploading many files
		s, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(
				"AKIAVIVVIRG7WZMLRODH",                     // id
				"4Jn+0LWXZ8AhI+RFHE6c2EimBSDn0EqVKBbAy2qd", // secret
				""), // token can be left blank for now
		})
		if err != nil {
			util.ResponseJSON(res, http.StatusBadRequest, "Could not upload file")
			return
		}

		profileImageFileName, err := util.UploadFileToS3(s, profileImageFile, profileImageFileHeader)
		if err != nil {
			util.ResponseJSON(res, http.StatusBadRequest, "Could not upload file")
			return
		}

		// wrap convert string to mongo objectid
		adminID, _ := primitive.ObjectIDFromHex(tokenData.ID)

		// insert into database
		result, err := collection.InsertOne(
			context.Background(),
			bson.D{
				primitive.E{Key: "adminId", Value: adminID},
				primitive.E{Key: "accountNumber", Value: util.RandomGen(1000000000, 99999999999)},
				primitive.E{Key: "firstName", Value: reqData.FirstName},
				primitive.E{Key: "lastName", Value: reqData.LastName},
				primitive.E{Key: "email", Value: reqData.Email},
				primitive.E{Key: "dob", Value: reqData.DOB},
				primitive.E{Key: "status", Value: true},
				primitive.E{Key: "profileImage", Value: profileImageFileName},
				primitive.E{Key: "phoneNumber", Value: reqData.PhoneNumber},
				primitive.E{Key: "phoneNumber", Value: reqData.Gender},
				primitive.E{Key: "phoneNumber", Value: reqData.Address},
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

// GetCustomerByAccountNumber to pull their details
func (c Controller) GetCustomerByAccountNumber(db *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var reqData models.CustomerSignUpValidatior
		var customer models.Customer
		util := utils.Util{}
		var errorMessage string
		vars := mux.Vars(req)

		// connect to collection
		collection := db.Database("dmobilizer").Collection("customers")

		num, _ := strconv.Atoi(vars["accountNumber"])

		reqData.AccountNumber = num

		log.Println("reqData.AccountNumber", reqData.AccountNumber)

		rules := govalidator.MapData{
			"accountNumber": []string{"required", "len:11"},
		}

		opts := govalidator.Options{
			Data:            &reqData,
			Rules:           rules,
			RequiredDefault: true,
		}

		v := govalidator.New(opts)
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

		err := collection.FindOne(context.TODO(), bson.D{primitive.E{Key: "accountNumber", Value: reqData.AccountNumber}}).Decode(&customer)

		log.Println("find doc", err, customer)

		// check for error
		if err != nil {
			util.ResponseJSON(res, http.StatusBadRequest, "No Account was found")
			return
		}

		customer.Password = ""

		util.ResponseJSON(res, http.StatusOK, customer)
		return
	}
}
