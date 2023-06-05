package commands

import (
	"time"

	"github.com/google/uuid"
)

type CreateProductCommand struct {
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
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Version     uint      `json:"version"`
	Deleted     bool      `json:"deleted,omitempty"`
}
