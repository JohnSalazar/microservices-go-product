package mongo_command

import (
	"context"
	"errors"
	"strings"

	commands "product/src/application/commands/product"
	events "product/src/application/events/product"

	mongo_event_handler "product/src/application/events/product/mongo"
	"product/src/data/repositories/interfaces"
	"product/src/dtos"
	"product/src/models"
	"product/src/validators"
	"time"

	common_validator "github.com/oceano-dev/microservices-go-common/validators"
)

type ProductCommandHandler struct {
	productMongoRepository interfaces.ProductRepository
	mongoEventHandler      *mongo_event_handler.ProductEventHandler
}

func NewProductCommandHandler(
	productMongoRepository interfaces.ProductRepository,
	mongoEventHandler *mongo_event_handler.ProductEventHandler,
) *ProductCommandHandler {
	common_validator.NewValidator("en")
	return &ProductCommandHandler{
		productMongoRepository: productMongoRepository,
		mongoEventHandler:      mongoEventHandler,
	}
}

func (product *ProductCommandHandler) CreateProductCommandHandler(ctx context.Context, command *commands.CreateProductCommand) error {
	productDto := &dtos.AddProduct{
		Name:        command.Name,
		Slug:        command.Slug,
		Description: command.Description,
		Price:       command.Price,
		Quantity:    command.Quantity,
		Image:       command.Image,
	}

	result := validators.ValidateAddProduct(productDto)
	if result != nil {
		return errors.New(strings.Join(result.([]string), ""))
	}

	productModel := &models.Product{
		ID:          command.ID,
		Name:        productDto.Name,
		Slug:        productDto.Slug,
		Description: productDto.Description,
		Price:       productDto.Price,
		Quantity:    productDto.Quantity,
		Image:       productDto.Image,
		CreatedAt:   command.CreatedAt,
		UpdatedAt:   command.UpdatedAt,
		Version:     command.Version,
		Deleted:     command.Deleted,
	}

	productExists, _ := product.productMongoRepository.FindByName(ctx, productModel.Name)
	if productExists != nil {
		return errors.New("product already exists")
	}

	productMongoExists, _ := product.productMongoRepository.FindByName(ctx, productModel.Name)
	if productMongoExists != nil && productMongoExists.ID != productModel.ID {
		return errors.New("product with this name already exists with another id")
	}

	_, err := product.productMongoRepository.Create(ctx, productModel)
	if err != nil {
		return err
	}

	productEvent := &events.ProductCreatedEvent{
		AggregateID: productModel.ID,
		MessageType: "product.create",
		Timestamp:   time.Now().UTC(),
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

	go product.mongoEventHandler.ProductCreatedEventHandler(productEvent)

	return nil
}

func (product *ProductCommandHandler) UpdateProductCommandHandler(ctx context.Context, command *commands.UpdateProductCommand) error {
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
		return errors.New(strings.Join(result.([]string), ""))
	}

	productModel := &models.Product{
		ID:          productDto.ID,
		Name:        productDto.Name,
		Slug:        productDto.Slug,
		Description: productDto.Description,
		Price:       productDto.Price,
		Image:       productDto.Image,
		Version:     productDto.Version,
	}

	productMongoExists, _ := product.productMongoRepository.FindByName(ctx, productModel.Name)
	if productMongoExists != nil && productMongoExists.ID != productModel.ID {
		return errors.New("product with this name already exists with another id")
	}

	_, err := product.productMongoRepository.Update(ctx, productModel)
	if err != nil {
		return err
	}

	productEvent := &events.ProductUpdatedEvent{
		AggregateID: productModel.ID,
		MessageType: "product.update",
		Timestamp:   time.Now().UTC(),
		ID:          productModel.ID,
		Name:        productModel.Name,
		Slug:        productModel.Slug,
		Description: productModel.Description,
		Price:       productModel.Price,
		Quantity:    productModel.Quantity,
		Image:       productModel.Image,
		Version:     productModel.Version,
	}

	go product.mongoEventHandler.ProductUpdatedEventHandler(productEvent)

	return nil
}
