package main

import (
	"dmobilizer/controllers"
	"dmobilizer/database"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

var db *mongo.Client

func main() {

	// connect database
	db = database.AppStartUp()

	// initiate a router
	router := mux.NewRouter()

	controller := controllers.Controller{}

	// Creat Agents Account
	router.HandleFunc("/agent", controller.CreateAgents(db)).Methods("POST")

	log.Println("Listen on the port 8000....")
	log.Fatal(http.ListenAndServe(":8000", router))

	// Close the connection once no longer needed
	database.AppClose()
}
