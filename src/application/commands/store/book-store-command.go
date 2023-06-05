package commands

import (
	"time"

	"product/src/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookStoreCommand struct {
	AggregateID uuid.UUID `json:"aggregateId"`
	MessageType string    `json:"messageType"`
	Timestamp   time.Time `json:"timestamp"`
	//ID          uuid.UUID         `json:"id"`
	OrderID primitive.ObjectID `json:"orderId"`
	//ProductID   uuid.UUID         `json:"productId"`
	Products []*models.Product `json:"products"`
	Stores   []*models.Store   `json:"stores"`
	//Quantity    uint              `json:"quantity"`
	//BookedAt    time.Time         `json:"booked_at"`
	//Sold        bool              `json:"sold"`
	//UpdatedAt   time.Time         `json:"updated_at"`
	//Version     uint              `json:"version"`
}
