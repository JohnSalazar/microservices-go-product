package mongo_event

import (
	"context"
	"fmt"

	common_nats "github.com/oceano-dev/microservices-go-common/nats"

	interfaces "product/src/data/repositories/redis"
	"product/src/models"

	events "product/src/application/events/product"
)

type ProductEventHandler struct {
	productRedisRepository interfaces.ProductRepository
	publisher              common_nats.Publisher
}

func NewProductEventHandler(
	productRedisRepository interfaces.ProductRepository,
	publisher common_nats.Publisher,
) *ProductEventHandler {
	return &ProductEventHandler{
		productRedisRepository: productRedisRepository,
		publisher:              publisher,
	}
}

func (product *ProductEventHandler) ProductCreatedEventHandler(event *events.ProductCreatedEvent) error {

	_product := &models.Product{
		ID:          event.ID,
		Name:        event.Name,
		Slug:        event.Slug,
		Description: event.Description,
		Price:       event.Price,
		Quantity:    event.Quantity,
		Image:       event.Image,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
		Version:     event.Version,
		Deleted:     event.Deleted,
	}

	ctx := context.Background()
	_, err := product.productRedisRepository.Set(ctx, _product)
	if err != nil {
		return err
	}

	fmt.Println("product set redis successfully!")

	return nil
}

func (product *ProductEventHandler) ProductUpdatedEventHandler(event *events.ProductUpdatedEvent) error {

	_product := &models.Product{
		ID:          event.ID,
		Name:        event.Name,
		Slug:        event.Slug,
		Description: event.Description,
		Price:       event.Price,
		Quantity:    event.Quantity,
		Image:       event.Image,
		UpdatedAt:   event.UpdatedAt,
		Version:     event.Version,
	}

	ctx := context.Background()
	_, err := product.productRedisRepository.Update(ctx, _product)
	if err != nil {
		return err
	}

	fmt.Println("product updated redis successfully!")

	return nil
}
