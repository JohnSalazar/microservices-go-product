package models

import (
	"time"

	"github.com/google/uuid"
)

type Store struct {
	ID        uuid.UUID `bson:"_id" json:"id"`
	ProductID uuid.UUID `bson:"product_id" json:"productid"`
	BookedAt  time.Time `bson:"booked_at" json:"booked_at"`
	Sold      bool      `bson:"sold" json:"sold"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	Version   uint      `bson:"version" json:"version"`
	Deleted   bool      `bson:"deleted" json:"deleted"`
}
