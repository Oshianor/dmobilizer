package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Customer Mongo Database model structure for user Document
type Customer struct {
	ObjectID      primitive.ObjectID `json:"_id" bson:"_id"`
	AgentID       string             `json:"agentId" bson:"agentId"`
	Email         string             `json:"email" bson:"email"`
	AccountNumber int                `json:"accountNumber" bson:"accountNumber"`
	FirstName     string             `json:"firstName" bson:"firstName"`
	LastName      string             `json:"lastName" bson:"lastName"`
	MiddleName    string             `json:"middleName" bson:"middleName"`
	DOB           string             `json:"dob" bson:"dob"`
	BVN           string             `json:"bvn" bson:"bvn"`
	PhoneNumber   string             `json:"phoneNumber" bson:"phoneNumber"`
	Address       string             `json:"address" bson:"address"`
	Gender        string             `json:"gender" bson:"gender"`
	ProfileImage  string             `json:"profileImage" bson:"profileImage"`
	Status        bool               `json:"status" bson:"status"`
	Password      string             `json:"password" bson:"password"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
}

// CustomerSignUpValidatior for creating customer account
type CustomerSignUpValidatior struct {
	Email         string `json:"email"`
	AccountNumber int    `json:"accountNumber"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	MiddleName    string `json:"middleName"`
	DOB           string `json:"dob"`
	BVN           string `json:"bvn"`
	PhoneNumber   string `json:"phoneNumber"`
	Address       string `json:"address"`
	Gender        string `json:"gender"`
}
