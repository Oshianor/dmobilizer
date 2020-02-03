package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Users Mongo Database model structure for user Document
type Users struct {
	ObjectID       primitive.ObjectID `json:"_id" bson:"_id"`
	AdminID        primitive.ObjectID `json:"adminId" bson:"adminId"`
	AgentID        int64              `json:"agentId" bson:"agentId"`
	Email          string             `json:"email" bson:"email"`
	FirstName      string             `json:"firstName" bson:"firstName"`
	Password       string             `json:"password" bson:"password"`
	LastName       string             `json:"lastName" bson:"lastName"`
	MiddleName     string             `json:"middleName" bson:"middleName"`
	DOB            string             `json:"dob" bson:"dob"`
	Type           string             `json:"type" bson:"type"`
	PhoneNumber    string             `json:"phoneNumber" bson:"phoneNumber"`
	ProfileImage   string             `json:"profileImage" bson:"profileImage"`
	Identification string             `json:"identification" bson:"identification"`
	Address        string             `json:"address" bson:"address"`
	Status         bool               `json:"status" bson:"status"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
}

// UserSignUpValidator for creating agent or teller
type UserSignUpValidator struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	FirstName       string `json:"firstName"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	LastName        string `json:"lastName"`
	MiddleName      string `json:"middleName"`
	DOB             string `json:"dob"`
	Type            string `json:"type"`
	PhoneNumber     string `json:"phoneNumber"`
	ProfileImage    string `json:"profileImage"`
	Identification  string `json:"identification"`
	Address         string `json:"address"`
}

// UserSignInValidator structure validator
type UserSignInValidator struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
