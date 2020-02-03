package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
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
	"github.com/aws/aws-sdk-go/service/s3"
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

		// connect to collection
		collection := db.Database("dmobilizer").Collection("users")

		data := goContext.Get(req, "user")

		mapstructure.Decode(data, &tokenData)

		rules := govalidator.MapData{
			"firstName":       []string{"required", "between:3,10"},
			"lastName":        []string{"required", "between:3,10"},
			"middleName":      []string{"required", "between:3,10"},
			"email":           []string{"required", "min:4", "max:20", "email"},
			"type":            []string{"required"},
			"dob":             []string{"required"},
			"phoneNumber":     []string{"required"},
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
			log.Fatal(err)
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
			log.Fatal(err)
		}

		fmt.Println(indexName)

		// set password and hash string
		reqData.Password = string(hash)

		adminID, _ := primitive.ObjectIDFromHex(tokenData.ID)

		// insert into database
		result, err := collection.InsertOne(
			context.Background(),
			bson.D{
				primitive.E{Key: "adminId", Value: adminID},
				primitive.E{Key: "firstName", Value: reqData.FirstName},
				primitive.E{Key: "lastName", Value: reqData.LastName},
				primitive.E{Key: "email", Value: reqData.Email},
				primitive.E{Key: "dob", Value: reqData.DOB},
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
				// log.Fatal(err)
				res.Header().Set("Content-type", "application/json")
				res.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(res).Encode("An Error ocurred")
				return
			}

			// append single market
			agents = append(agents, &doc)
		}

		if err := cursor.Err(); err != nil {
			log.Fatal(err)
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

// UploadImage to file system
func (c Controller) UploadImage() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("File Upload Endpoint Hit")

		maxSize := int64(20000000)

		// // Parse our multipart form, 10 << 20 specifies a maximum
		// // upload of 10 MB files.
		// req.ParseMultipartForm(10 << 20)

		err := req.ParseMultipartForm(maxSize)
		if err != nil {
			log.Println(err)
			fmt.Fprintf(res, "Image too large. Max Size: %v", maxSize)
			return
		}

		// FormFile returns the first file for the given key `myFile`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file
		file, fileHeader, err := req.FormFile("myFile")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Printf("Uploaded File: %+v\n", fileHeader.Filename)
		fmt.Printf("File Size: %+v\n", fileHeader.Size)
		fmt.Printf("MIME Header: %+v\n", fileHeader.Header)

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
			fmt.Fprintf(res, "Could not upload file")
		}
		fileName, err := UploadFileToS3(s, file, fileHeader)
		if err != nil {
			fmt.Fprintf(res, "Could not upload file")
			return
		}
		fmt.Fprintf(res, "Image uploaded successfully: %v", fileName)

	}
}

// UploadFileToS3 saves a file to aws bucket and returns the url to // the file and an error if there's any
func UploadFileToS3(s *session.Session, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
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

// UploadImage to file system
func (c Controller) UploadImage() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("File Upload Endpoint Hit")

		maxSize := int64(20000000)

		// // Parse our multipart form, 10 << 20 specifies a maximum
		// // upload of 10 MB files.
		// req.ParseMultipartForm(10 << 20)

		err := req.ParseMultipartForm(maxSize)
		if err != nil {
			log.Println(err)
			fmt.Fprintf(res, "Image too large. Max Size: %v", maxSize)
			return
		}

		// FormFile returns the first file for the given key `myFile`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file
		file, fileHeader, err := req.FormFile("myFile")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Printf("Uploaded File: %+v\n", fileHeader.Filename)
		fmt.Printf("File Size: %+v\n", fileHeader.Size)
		fmt.Printf("MIME Header: %+v\n", fileHeader.Header)

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
			fmt.Fprintf(res, "Could not upload file")
		}
		fileName, err := UploadFileToS3(s, file, fileHeader)
		if err != nil {
			fmt.Fprintf(res, "Could not upload file")
			return
		}
		fmt.Fprintf(res, "Image uploaded successfully: %v", fileName)

	}
}

