package controllers

import (
	"context"
	"dmobilizer/models"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Controller empty struct
type Controller struct{}

var agent models.Agents

// CreateAgents api to creat new agents
func (c Controller) CreateAgents(db *mongo.Client) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		reqBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode("Kindly enter data with the event title and description only in order to update")
		}

		// unmarshal data to the struct
		json.Unmarshal(reqBody, &agent)

		v := validator.New()
		a := models.AgentsValidator{agent}
		err := v.Struct(a)

		for _, e := range err.(validator.ValidationErrors) {
			// fmt.Println(e)/
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode(e)
		}

		// connect to collection
		collection := db.Database("dmobilizer").Collection("agents")

		// insert into database
		result, err := collection.InsertOne(
			context.Background(),
			bson.D{
				primitive.E{Key: "firstName", Value: agent.FirstName},
				primitive.E{Key: "lastName", Value: agent.LastName},
				primitive.E{Key: "dateOfBirth", Value: agent.DateOfBirth},
				primitive.E{Key: "status", Value: false},
				primitive.E{Key: "createdAt", Value: time.Now()},
			},
		)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(res).Encode("An Error ocurred")
		}

		if result != nil {
			res.WriteHeader(http.StatusCreated)
			json.NewEncoder(res).Encode("Agent account succesfully created")
		}

	}
}
