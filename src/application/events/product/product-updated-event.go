package postgres_event

import (
	"time"

	"github.com/google/uuid"
)

type ProductUpdatedEvent struct {
	AggregateID uuid.UUID `json:"aggregateId"`
	MessageType string    `json:"messageType"`
	Timestamp   time.Time `json:"timestamp"`
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Price       float32   `json:"price"`
	Quantity    uint      `json:"quantity"`
	Image       string    `json:"image,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Version     uint      `json:"version"`
}
