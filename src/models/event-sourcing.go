package models

import (
	"time"

	"github.com/google/uuid"
)

type EventSourcing struct {
	ID          uuid.UUID `json:"id" bson:"_id"`
	AggregateID uuid.UUID `json:"aggregateId" bson:"aggregateId"`
	MessageType string    `json:"messageType" bson:"messageType"`
	Timestamp   time.Time `json:"timestamp" bson:"timestamp"`
	Data        string    `json:"data" bson:"data"`
}
