package commands

import (
	"product/src/models"
	"time"

	"github.com/google/uuid"
)

type UnbookStoreCommand struct {
	AggregateID uuid.UUID       `json:"aggregateId"`
	MessageType string          `json:"messageType"`
	Timestamp   time.Time       `json:"timestamp"`
	ID          uuid.UUID       `json:"id"`
	Sold        bool            `json:"sold"`
	BookedAt    time.Time       `json:"booked_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Version     uint            `json:"version"`
	Stores      []*models.Store `json:"stores"`
}
