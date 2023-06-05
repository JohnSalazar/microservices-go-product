package dtos

import (
	"github.com/google/uuid"
)

type PaymentStore struct {
	ID   uuid.UUID `json:"id"`
	Sold bool      `json:"sold"`
}
