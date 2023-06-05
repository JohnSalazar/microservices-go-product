package events

import (
	"product/src/models"
	"time"

	"github.com/google/uuid"
)

type StoreUnbookedEvent struct {
	AggregateID uuid.UUID       `json:"aggregateId"`
	MessageType string          `json:"messageType"`
	Timestamp   time.Time       `json:"timestamp"`
	Stores      []*models.Store `json:"stores"`
}
