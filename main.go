package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/instiq/dmobilizer/controllers"
	"github.com/instiq/dmobilizer/database"
	"github.com/instiq/dmobilizer/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

var db *mongo.Client

func main() {

	// connect database
	db = database.AppStartUp()

	// initiate a router
	router := mux.NewRouter()

	controller := controllers.Controller{}
	util := utils.Util{}

	// auth
	router.HandleFunc("/auth/admin", controller.LoginAdmin(db)).Methods("POST")
	router.HandleFunc("/auth/user", controller.LoginUser(db)).Methods("POST")

	// Admin
	router.HandleFunc("/admin", controller.CreateAdmin(db)).Methods("POST")

	// Users
	router.HandleFunc("/user", util.TokenVerifyMiddleWare(controller.CreateAgents(db), "admin")).Methods("POST")
	router.HandleFunc("/user/deposit", util.TokenVerifyMiddleWare(controller.Deposit(db), "agent")).Methods("POST")
	router.HandleFunc("/user", controller.GetAllAgents(db)).Methods("GET")

	// customer
	router.HandleFunc("/customer", util.TokenVerifyMiddleWare(controller.CreateCustomers(db), "agent")).Methods("POST")
	router.HandleFunc("/customer/{accountNumber}", util.TokenVerifyMiddleWare(controller.GetCustomerByAccountNumber(db), "agent")).Methods("GET")

	// router.HandleFunc("/img", controller.UploadImage()).Methods("POST")

	router.Use(mux.CORSMethodMiddleware(router))

	log.Println("Listen on the port 8000....")
	log.Fatal(http.ListenAndServe(":8000", router))

	// Close the connection once no longer needed
	database.AppClose()
}
