package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Client
var err error

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// AppStartUp is to start the database
func AppStartUp() *mongo.Client {
	fmt.Println("Starting programme")

	// db, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	db, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))

	// create acontext
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// close the context at the end of the function
	defer cancel()

	err = db.Connect(ctx)

	err := db.Ping(context.TODO(), nil)
	logFatal(err)

	fmt.Println("Connected to MongoDB!")

	return db
}

// AppClose is to close connection
func AppClose() {
	// Close the connection once no longer needed
	err := db.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Connection to MongoDB closed.")
	}
}

// // Set client options
// clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
// clientOptions := options.Client().ApplyURI("mongodb://34.235.146.205:27017")

// db, err = mongo.Connect(ctx, clientOptions)
// logFatal(err)

// // Check the connection
// err := db.Ping(context.TODO(), nil)
// logFatal(err)

// fmt.Println("Connected to MongoDB!")
