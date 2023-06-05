package commands

import (
	"product/src/models"
	"time"

	"github.com/google/uuid"
)

type CreateStoreCommand struct {
	AggregateID uuid.UUID       `json:"aggregateId"`
	MessageType string          `json:"messageType"`
	Timestamp   time.Time       `json:"timestamp"`
	ID          uuid.UUID       `json:"id"`
	ProductID   uuid.UUID       `json:"productId"`
	Quantity    uint            `json:"quantity"`
	Stores      []*models.Store `json:"stores"`
	CreatedAt   time.Time       `json:"created_at"`
}
