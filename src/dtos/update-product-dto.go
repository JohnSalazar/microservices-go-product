package dtos

import (
	"github.com/google/uuid"
)

type UpdateProduct struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Price       float32   `json:"price"`
	Image       string    `json:"image,omitempty"`
	Version     uint      `json:"version"`
}
