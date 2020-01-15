package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Agents Mongo Database model structure for user Document
type Agents struct {
	ObjectID    primitive.ObjectID `json:"_id" bson:"_id"`
	FirstName   string             `json:"firstName" bson:"firstName"`
	LastName    string             `json:"lastName" bson:"lastName"`
	DateOfBirth string             `json:"dateOfBirth" bson:"dateOfBirth"`
	Status      bool               `json:"status" bson:"status"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
}

// AgentsValidator for create agents
type AgentsValidator struct {
	FirstName   string `json:"firstName" validate:"required"`
	LastName    string `json:"lastName" validate:"required"`
	DateOfBirth string `json:"dateOfBirth" validate:"required"`
	Status      bool   `json:"status" validate:"required"`
}
