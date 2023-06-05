package dtos

import (
	"github.com/google/uuid"
)

type AddStore struct {
	ProductID uuid.UUID `json:"productid"`
	Quantity  uint      `json:"quantity"`
}
