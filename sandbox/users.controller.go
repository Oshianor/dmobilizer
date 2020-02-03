package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/instiq/dmobilizer/models"

	"github.com/thedevsaddam/govalidator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Controller empty struct
type Controller struct{}

// CreateAgents api to creat new agents
func (c Controller) CreateAgents(db *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var user models.UserSignUpValidator
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
			Data:    &user,
			Rules:   rules,
		}

		v := govalidator.New(opts)
		e := v.ValidateJSON()
		fmt.Println(e) // your incoming JSON data in Go data struct

		if len(e) != 0 {
			// err := map[string]interface{}{"validationError": e}
			res.Header().Set("Content-type", "application/json")
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode(e)
			return
		}

		// connect to collection
		collection := db.Database("dmobilizer").Collection("users")

		// insert into database
		result, err := collection.InsertOne(
			context.Background(),
			bson.D{
				primitive.E{Key: "firstName", Value: user.FirstName},
				primitive.E{Key: "lastName", Value: user.LastName},
				primitive.E{Key: "email", Value: user.Email},
				primitive.E{Key: "dob", Value: user.DOB},
				primitive.E{Key: "status", Value: false},
				primitive.E{Key: "createdAt", Value: time.Now()},
			},
		)

		if err != nil {
			res.Header().Set("Content-type", "application/json")
			res.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(res).Encode("An Error ocurred")
		}

		if result != nil {
			res.Header().Set("Content-type", "application/json")
			res.WriteHeader(http.StatusCreated)
			json.NewEncoder(res).Encode("Agent account succesfully created")
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

		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		req.ParseMultipartForm(10 << 20)

		// FormFile returns the first file for the given key `myFile`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file
		file, handler, err := req.FormFile("myFile")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Printf("Uploaded File: %+v\n", handler.Filename)
		fmt.Printf("File Size: %+v\n", handler.Size)
		fmt.Printf("MIME Header: %+v\n", handler.Header)

		// Create a temporary file within our temp-images directory that follows
		// a particular naming pattern
		tempFile, err := ioutil.TempFile("/tmp", "upload-*.png")
		if err != nil {
			fmt.Println(err)
		}
		defer tempFile.Close()

		// read all of the contents of our uploaded file into a
		// byte array
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(tempFile)

		// write this byte array to our temporary file
		tempFile.Write(fileBytes)
		// return that we have successfully uploaded our file!
		fmt.Fprintf(res, "Successfully Uploaded File\n")

	}
}

// reqBody, err := ioutil.ReadAll(req.Body)
// if err != nil {
// 	res.WriteHeader(http.StatusBadRequest)
// 	json.NewEncoder(res).Encode("Kindly enter data with the event title and description only in order to update")
// }

// // unmarshal data to the struct
// json.Unmarshal(reqBody, &agent)
