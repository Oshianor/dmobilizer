package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Transactions Mongo Database model structure
type Transactions struct {
	ObjectID   primitive.ObjectID `json:"_id" bson:"_id"`
	AgentID    primitive.ObjectID `json:"agentId" bson:"agentId"`
	CustomerID primitive.ObjectID `json:"customerId" bson:"customerId"`
	Type       string             `json:"type" bson:"type"`     //withdraw, deposit, remit
	Status     string             `json:"status" bson:"status"` //approved, pending
	Amount     float64            `json:"amount" bson:"amount"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
}

type Deposit struct {
	CustomerID string  `json:"customerId"`
	Amount     float64 `json:"amount"`
}
