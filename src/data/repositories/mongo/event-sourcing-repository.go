package mongo_repository

import (
	"context"
	"product/src/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type eventSourcingRepository struct {
	database *mongo.Database
}

func NewEventSourcingRepository(
	database *mongo.Database,
) *eventSourcingRepository {
	return &eventSourcingRepository{
		database: database,
	}
}

func (r *eventSourcingRepository) collectionName() string {
	return "events"
}

func (r *eventSourcingRepository) collection() *mongo.Collection {
	return r.database.Collection(r.collectionName())
}

func (r *eventSourcingRepository) Create(ctx context.Context, eventStore *models.EventSourcing) error {

	fields := bson.M{
		"_id":         eventStore.ID.String(),
		"aggregateId": eventStore.AggregateID.String(),
		"messageType": eventStore.MessageType,
		"timestamp":   eventStore.Timestamp,
		"data":        eventStore.Data,
	}

	_, err := r.collection().InsertOne(ctx, fields)
	if err != nil {
		return err
	}

	return nil
}

func (r *eventSourcingRepository) CreateMany(ctx context.Context, eventsStore []*models.EventSourcing) error {

	fields := make([]interface{}, len(eventsStore))
	for i, event := range eventsStore {
		fields[i] = bson.M{
			"_id":         event.ID.String(),
			"aggregateId": event.AggregateID.String(),
			"messageType": event.MessageType,
			"timestamp":   event.Timestamp,
			"data":        event.Data,
		}
	}

	_, err := r.collection().InsertMany(ctx, fields)
	if err != nil {
		return err
	}

	return nil
}
