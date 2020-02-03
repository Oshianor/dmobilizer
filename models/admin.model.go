package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Admin Mongo Database model structure
type Admin struct {
	ObjectID    primitive.ObjectID `json:"_id" bson:"_id"`
	Email       string             `json:"email" bson:"email"`
	Password    string             `json:"password" bson:"password"`
	FirstName   string             `json:"firstName" bson:"firstName"`
	LastName    string             `json:"lastName" bson:"lastName"`
	PhoneNumber string             `json:"phoneNumber" bson:"phoneNumber"`
	Status      bool               `json:"status" bson:"status"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
}

// AdminSignUpValidator structure validator
type AdminSignUpValidator struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
}

// AdminSignInValidator structure validator
type AdminSignInValidator struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
