package events

import (
	"product/src/models"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StoreBookedEvent struct {
	AggregateID uuid.UUID          `json:"aggregateId"`
	MessageType string             `json:"messageType"`
	Timestamp   time.Time          `json:"timestamp"`
	OrderID     primitive.ObjectID `json:"orderId"`
	Stores      []*models.Store    `json:"stores"`
}
