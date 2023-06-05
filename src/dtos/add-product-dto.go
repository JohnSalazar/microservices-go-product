package dtos

import "github.com/google/uuid"

type AddProduct struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Price       float32   `json:"price"`
	Quantity    uint      `json:"quantity"`
	Image       string    `json:"image,omitempty"`
}
