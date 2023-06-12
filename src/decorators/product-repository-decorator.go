package decorators

import (
	"context"
	"encoding/json"
	"fmt"
	product_repository "product/src/data/repositories/interfaces"
	redis_repository "product/src/data/repositories/redis"
	"product/src/models"
	"product/src/nats/subjects"

	command_product "product/src/application/commands/product"

	common_nats "github.com/JohnSalazar/microservices-go-common/nats"
	trace "github.com/JohnSalazar/microservices-go-common/trace/otel"
	"github.com/google/uuid"
)

type ProductRepositoryDecorator interface {
	GetAll(ctx context.Context, name string, page int, size int) ([]*models.Product, error)
	FindByID(ctx context.Context, ID uuid.UUID) (*models.Product, error)
	FindBySlug(ctx context.Context, slug string) (*models.Product, error)
}

type productRepositoryDecorator struct {
	mongoRepository    product_repository.ProductRepository
	postgresRepository product_repository.ProductRepository
	redisRepository    redis_repository.ProductRepository
	publisher          common_nats.Publisher
}

func NewProductRepositoryDecorator(
	mongoRepository product_repository.ProductRepository,
	postgresRepository product_repository.ProductRepository,
	redisRepository redis_repository.ProductRepository,
	publisher common_nats.Publisher,
) *productRepositoryDecorator {
	return &productRepositoryDecorator{
		mongoRepository:    mongoRepository,
		postgresRepository: postgresRepository,
		redisRepository:    redisRepository,
		publisher:          publisher,
	}
}

func (decorator *productRepositoryDecorator) GetAll(ctx context.Context, name string, page int, size int) ([]*models.Product, error) {
	_, span := trace.NewSpan(ctx, "ProductRepositoryAdapter.GetAll")
	defer span.End()

	db := "redis"
	products, err := decorator.redisRepository.GetAll(ctx, name, page, size)
	if err != nil {
		fmt.Println("err redis: ", err)
	}
	if len(products) == 0 {
		db = "mongo"
		products, err = decorator.mongoRepository.GetAll(ctx, name, page, size)
		if err != nil {
			fmt.Println("err mongo: ", err)
		}
		if len(products) == 0 {
			db = "postgres"
			products, err = decorator.postgresRepository.GetAll(ctx, name, page, size)
			if err != nil {
				fmt.Println("err postgres: ", err)
			}
		}
	}

	// db := "redis"
	// products, err := decorator.redisRepository.GetAll(ctx, name, page, size)

	fmt.Println("page: ", page)
	fmt.Println(db)
	return products, err
}

func (decorator *productRepositoryDecorator) FindByID(ctx context.Context, ID uuid.UUID) (*models.Product, error) {
	_, span := trace.NewSpan(ctx, "ProductRepositoryAdapter.FindByID")
	defer span.End()

	db := "mongo"
	product, err := decorator.mongoRepository.FindByID(ctx, ID)
	if err != nil {
		fmt.Println("err mongo: ", err)
	}
	if product == nil {
		db = "postgres"
		product, err = decorator.postgresRepository.FindByID(ctx, ID)
		if err != nil {
			fmt.Println("err postgres: ", err)
		}
	}

	if db == "postgres" && product != nil {
		decorator.updateRepositories(ctx, product)
		// err := decorator.updateRepositories(ctx, product)
		// if err != nil {
		// 	fmt.Println("updateRepositories error: ", err)
		// }
	}

	fmt.Println(db)
	return product, err
}

func (decorator *productRepositoryDecorator) FindBySlug(ctx context.Context, slug string) (*models.Product, error) {
	_, span := trace.NewSpan(ctx, "ProductRepositoryAdapter.FindBySlug")
	defer span.End()

	db := "mongo"
	product, err := decorator.mongoRepository.FindBySlug(ctx, slug)
	if err != nil {
		fmt.Println("err mongo: ", err)
	}
	if product == nil {
		db = "postgres"
		product, err = decorator.postgresRepository.FindBySlug(ctx, slug)
		if err != nil {
			fmt.Println("err postgres: ", err)
		}
	}

	if db == "postgres" && product != nil {
		decorator.updateRepositories(ctx, product)
	}

	fmt.Println(db)
	return product, err
}

func (decorator *productRepositoryDecorator) updateRepositories(ctx context.Context, product *models.Product) error {
	_, span := trace.NewSpan(ctx, "ProductController.updateRepositories")
	defer span.End()

	createProductCommand := &command_product.CreateProductCommand{
		ID:          product.ID,
		Name:        product.Name,
		Slug:        product.Slug,
		Description: product.Description,
		Price:       product.Price,
		Quantity:    product.Quantity,
		Image:       product.Image,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
		Version:     product.Version,
		Deleted:     product.Deleted,
	}

	data, err := json.Marshal(createProductCommand)
	if err != nil {
		trace.FailSpan(span, "error json parse")
		fmt.Println(fmt.Errorf("error json parse: %v", err))
		return err
	}

	err = decorator.publisher.Publish(string(subjects.ProductCreateMongo), data)
	if err != nil {
		trace.FailSpan(span, "error product mongo publish")
		fmt.Println(fmt.Errorf("error product mongo publish: %v", err))
		return err
	}

	decorator.redisRepository.Set(ctx, product)

	// _, err = decorator.redisRepository.Set(ctx, product)
	// if err != nil {
	// 	trace.FailSpan(span, "error redis set")
	// 	fmt.Println(fmt.Errorf("error redis set: %v", err))
	// 	return err
	// }

	return nil
}
