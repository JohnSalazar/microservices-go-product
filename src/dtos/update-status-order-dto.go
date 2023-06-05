package dtos

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateStatusOrder struct {
	ID       primitive.ObjectID `json:"id"`
	Status   uint               `json:"status"`
	StatusAt time.Time          `json:"status_at"`
}
