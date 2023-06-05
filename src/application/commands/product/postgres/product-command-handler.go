package postgres_command

import (
	"context"
	"errors"
	commands "product/src/application/commands/product"
	events "product/src/application/events/product"
	postgres_event_handler "product/src/application/events/product/postgres"
	"strings"

	repository_interface "product/src/data/repositories/interfaces"

	"product/src/dtos"
	"product/src/models"
	"product/src/validators"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	common_validator "github.com/oceano-dev/microservices-go-common/validators"
)

type ProductCommandHandler struct {
	productPostgresRepository    repository_interface.ProductRepository
	eventSourcingMongoRepository repository_interface.EventSourcingRepository
	postgresEventHandler         *postgres_event_handler.ProductEventHandler
}

func NewProductCommandHandler(
	productPostgresRepository repository_interface.ProductRepository,
	eventSourcingMongoRepository repository_interface.EventSourcingRepository,
	postgresEventHandler *postgres_event_handler.ProductEventHandler,
) *ProductCommandHandler {
	common_validator.NewValidator("en")
	return &ProductCommandHandler{
		productPostgresRepository:    productPostgresRepository,
		eventSourcingMongoRepository: eventSourcingMongoRepository,
		postgresEventHandler:         postgresEventHandler,
	}
}

func (product *ProductCommandHandler) CreateProductCommandHandler(ctx context.Context, command *commands.CreateProductCommand) (*models.Product, error) {
	productDto := &dtos.AddProduct{
		ID:          command.ID,
		Name:        command.Name,
		Slug:        command.Slug,
		Description: command.Description,
		Price:       command.Price,
		Quantity:    command.Quantity,
		Image:       command.Image,
	}

	result := validators.ValidateAddProduct(productDto)
	if result != nil {
		return nil, errors.New(strings.Join(result.([]string), ""))
	}

	productModel := &models.Product{
		// ID:          uuid.New(),
		ID:          productDto.ID,
		Name:        productDto.Name,
		Slug:        productDto.Slug,
		Description: productDto.Description,
		Price:       productDto.Price,
		Quantity:    productDto.Quantity,
		Image:       productDto.Image,
		CreatedAt:   time.Now().UTC(),
	}

	productPostgresExists, err := product.productPostgresRepository.FindByName(ctx, productModel.Name)
	if err != nil {
		return nil, err
	}
	if productPostgresExists != nil {
		return nil, errors.New("product already exists")
	}

	productModel, err = product.productPostgresRepository.Create(ctx, productModel)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(productModel)

	eventSourcing := &models.EventSourcing{
		ID:          uuid.New(),
		AggregateID: productModel.ID,
		MessageType: "product.create",
		Timestamp:   time.Now().UTC(),
		Data:        string(data),
	}
	go product.eventSourcingMongoRepository.Create(ctx, eventSourcing)

	productEvent := &events.ProductCreatedEvent{
		AggregateID: productModel.ID,
		MessageType: eventSourcing.MessageType,
		Timestamp:   eventSourcing.Timestamp,
		ID:          productModel.ID,
		Name:        productModel.Name,
		Slug:        productModel.Slug,
		Description: productModel.Description,
		Price:       productModel.Price,
		Quantity:    productModel.Quantity,
		Image:       productModel.Image,
		CreatedAt:   productModel.CreatedAt,
		UpdatedAt:   productModel.UpdatedAt,
		Version:     productModel.Version,
		Deleted:     productModel.Deleted,
	}

	go product.postgresEventHandler.ProductCreatedEventHandler(ctx, productEvent)

	return productModel, nil
}

func (product *ProductCommandHandler) UpdateProductCommandHandler(ctx context.Context, command *commands.UpdateProductCommand) (*models.Product, error) {
	productDto := &dtos.UpdateProduct{
		ID:          command.ID,
		Name:        command.Name,
		Slug:        command.Slug,
		Description: command.Description,
		Price:       command.Price,
		Image:       command.Image,
		Version:     command.Version,
	}

	result := validators.ValidateUpdateProduct(productDto)
	if result != nil {
		return nil, errors.New(strings.Join(result.([]string), ""))
	}

	productModel := &models.Product{
		ID:          productDto.ID,
		Name:        productDto.Name,
		Slug:        productDto.Slug,
		Description: productDto.Description,
		Price:       productDto.Price,
		Image:       productDto.Image,
		Version:     productDto.Version,
		UpdatedAt:   time.Now().UTC(),
	}

	productPostgresExists, _ := product.productPostgresRepository.FindByName(ctx, productModel.Name)
	if productPostgresExists != nil && productPostgresExists.ID != productModel.ID {
		return nil, errors.New("product with this name already exists with another id")
	}

	productModel, err := product.productPostgresRepository.Update(ctx, productModel)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(productModel)

	storeEvent := &models.EventSourcing{
		ID:          uuid.New(),
		AggregateID: productModel.ID,
		MessageType: "product.update",
		Timestamp:   time.Now().UTC(),
		Data:        string(data),
	}
	go product.eventSourcingMongoRepository.Create(ctx, storeEvent)

	productEvent := &events.ProductUpdatedEvent{
		AggregateID: productModel.ID,
		MessageType: storeEvent.MessageType,
		Timestamp:   storeEvent.Timestamp,
		ID:          productModel.ID,
		Name:        productModel.Name,
		Slug:        productModel.Slug,
		Description: productModel.Description,
		Price:       productModel.Price,
		Quantity:    productModel.Quantity,
		Image:       productModel.Image,
		UpdatedAt:   productModel.UpdatedAt,
		Version:     productModel.Version,
	}

	go product.postgresEventHandler.ProductUpdatedEventHandler(ctx, productEvent)

	return productModel, nil
}
