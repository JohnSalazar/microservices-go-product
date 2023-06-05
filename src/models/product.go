package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `bson:"_id" json:"id"`
	Name        string    `bson:"name" json:"name"`
	Slug        string    `bson:"slug" json:"slug"`
	Description string    `bson:"description" json:"description"`
	Price       float32   `bson:"price" json:"price"`
	Quantity    uint      `json:"quantity,omitempty"`
	Image       string    `bson:"image" json:"image,omitempty"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at,omitempty"`
	Version     uint      `bson:"version" json:"version"`
	Deleted     bool      `bson:"deleted" json:"deleted,omitempty"`
}
