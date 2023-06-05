package dtos

import (
	"product/src/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UpdateStoreOrder struct {
	ID     primitive.ObjectID `json:"id"`
	Stores []*models.Store    `json:"stores"`
}
