package commands

import (
	"time"

	"github.com/google/uuid"
)

type UpdateProductCommand struct {
	AggregateID uuid.UUID `json:"aggregateId"`
	MessageType string    `json:"messageType"`
	Timestamp   time.Time `json:"timestamp"`
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Price       float32   `json:"price"`
	Image       string    `json:"image,omitempty"`
	Version     uint      `json:"version"`
}
