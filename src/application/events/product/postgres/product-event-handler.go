package postgres_event

import (
	"context"
	"encoding/json"

	common_nats "github.com/oceano-dev/microservices-go-common/nats"

	commandProduct "product/src/application/commands/product"
	commandStore "product/src/application/commands/store"
	"product/src/nats/subjects"

	events "product/src/application/events/product"
)

type ProductEventHandler struct {
	publisher common_nats.Publisher
}

func NewProductEventHandler(
	publisher common_nats.Publisher,
) *ProductEventHandler {
	return &ProductEventHandler{
		publisher: publisher,
	}
}

func (product *ProductEventHandler) ProductCreatedEventHandler(ctx context.Context, event *events.ProductCreatedEvent) error {
	createStorePostgresCommand := &commandStore.CreateStoreCommand{
		ProductID: event.ID,
		Quantity:  event.Quantity,
	}

	dataCommand, _ := json.Marshal(createStorePostgresCommand)
	err := product.publisher.Publish(string(subjects.StoreCreatePostgres), dataCommand)
	if err != nil {
		return err
	}

	dataEvent, _ := json.Marshal(event)
	err = product.publisher.Publish(string(subjects.ProductCreateMongo), dataEvent)
	if err != nil {
		return err
	}

	return nil
}

func (product *ProductEventHandler) ProductUpdatedEventHandler(ctx context.Context, event *events.ProductUpdatedEvent) error {
	updateProductMongoCommand := &commandProduct.UpdateProductCommand{
		ID:          event.ID,
		Name:        event.Name,
		Slug:        event.Slug,
		Description: event.Description,
		Price:       event.Price,
		Image:       event.Image,
		Version:     event.Version,
	}

	dataCommand, _ := json.Marshal(updateProductMongoCommand)
	err := product.publisher.Publish(string(subjects.ProductUpdateMongo), dataCommand)
	if err != nil {
		return err
	}

	return nil
}
