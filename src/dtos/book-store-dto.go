package dtos

import (
	"product/src/models"
)

type BookStore struct {
	Products []*models.Product `json:"products"`
}
