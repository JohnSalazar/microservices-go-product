package events

import (
	"product/src/models"
	"time"

	"github.com/google/uuid"
)

type StoreCreatedEvent struct {
	AggregateID uuid.UUID       `json:"aggregateId"`
	MessageType string          `json:"messageType"`
	Timestamp   time.Time       `json:"timestamp"`
	ID          uuid.UUID       `json:"id"`
	ProductID   uuid.UUID       `json:"productId"`
	Quantity    uint            `json:"quantity"`
	CreatedAt   time.Time       `json:"createdAt"`
	Stores      []*models.Store `json:"stores"`
}
