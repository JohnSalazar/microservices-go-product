package redis_repository

import (
	"context"
	"fmt"
	"log"
	"product/src/models"
	"strconv"
	"strings"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type ProductRepository interface {
	GetAll(ctx context.Context, name string, page int, size int) ([]*models.Product, error)
	Set(ctx context.Context, product *models.Product) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) (*models.Product, error)
	Refresh(ctx context.Context, products []*models.Product) error
}

type productRepository struct {
	database *redis.Client
}

var search *redisearch.Client
var schema *redisearch.Schema

func NewProductRepository(database *redis.Client) *productRepository {
	result := &productRepository{
		database: database,
	}

	addr := database.Options().Addr
	search = redisearch.NewClient(addr, "productsIndex")

	schema = result.schema()
	search.Drop()

	return result
}

func (r *productRepository) GetAll(ctx context.Context, name string, page int, size int) ([]*models.Product, error) {
	products := []*models.Product{}

	if len(name) == 1 && len(strings.Trim(name, " ")) == 0 {
		name = "*"
	} else {
		name = fmt.Sprintf("@name:*%s*", name)
	}

	docs, _, err := search.Search(redisearch.NewQuery(name).
		Limit((page-1)*size, size).
		SetSortBy("name", true).
		SetReturnFields("id", "name", "slug", "description", "price", "quantity", "image", "version"))

	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		product, err := r.mapProduct(&doc)
		if err != nil {
			log.Fatal("mapProduct error: ", err)
		}

		products = append(products, product)
	}

	return products, nil
}

func (r *productRepository) Set(ctx context.Context, product *models.Product) (*models.Product, error) {
	if product == nil {
		return nil, fmt.Errorf("product is nil")
	}

	doc := redisearch.NewDocument(product.ID.String(), 1.0)
	doc.Set("id", product.ID.String()).
		Set("name", product.Name).
		Set("slug", product.Slug).
		Set("description", product.Description).
		Set("price", product.Price).
		Set("quantity", product.Quantity).
		Set("image", product.Image).
		Set("version", product.Version)

	if err := search.Index(doc); err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) Update(ctx context.Context, product *models.Product) (*models.Product, error) {
	search.DeleteDocument(product.ID.String())

	newdoc := redisearch.NewDocument(product.ID.String(), 1.0)
	newdoc.Set("id", product.ID).
		Set("name", product.Name).
		Set("slug", product.Slug).
		Set("description", product.Description).
		Set("price", product.Price).
		Set("quantity", product.Quantity).
		Set("image", product.Image).
		Set("version", product.Version)

	if err := search.Index(newdoc); err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) Refresh(ctx context.Context, products []*models.Product) error {
	if len(products) == 0 {
		return nil
	}

	result := r.database.FlushDB(ctx)
	fmt.Println(result)

	var docs []redisearch.Document
	for _, product := range products {
		doc := redisearch.NewDocument(product.ID.String(), 1.0)
		doc.Set("id", product.ID.String()).
			Set("name", product.Name).
			Set("slug", product.Slug).
			Set("description", product.Description).
			Set("price", product.Price).
			Set("quantity", product.Quantity).
			Set("image", product.Image).
			Set("version", product.Version)

		docs = append(docs, doc)
	}

	if err := search.CreateIndex(schema); err != nil {
		return err
	}

	if err := search.Index(docs...); err != nil {
		return err
	}

	return nil
}

func (r *productRepository) schema() *redisearch.Schema {
	schema := redisearch.NewSchema(redisearch.DefaultOptions).
		// AddField(redisearch.NewTagFieldOptions("id", redisearch.TagFieldOptions{Separator: byte(';')})).
		AddField(redisearch.NewTagFieldOptions("id", redisearch.TagFieldOptions{})).
		AddField(redisearch.NewTextFieldOptions("name", redisearch.TextFieldOptions{})).
		AddField(redisearch.NewTextFieldOptions("slug", redisearch.TextFieldOptions{})).
		AddField(redisearch.NewTextFieldOptions("description", redisearch.TextFieldOptions{})).
		AddField(redisearch.NewNumericFieldOptions("price", redisearch.NumericFieldOptions{})).
		AddField(redisearch.NewNumericFieldOptions("quantity", redisearch.NumericFieldOptions{})).
		AddField(redisearch.NewTextFieldOptions("image", redisearch.TextFieldOptions{})).
		AddField(redisearch.NewNumericFieldOptions("version", redisearch.NumericFieldOptions{}))

	return schema
}

func (r *productRepository) mapProduct(object *redisearch.Document) (*models.Product, error) {
	product := &models.Product{
		Name: object.Properties["name"].(string),
	}

	ID, err := uuid.Parse(object.Properties["id"].(string))
	if err != nil {
		return nil, err
	}
	product.ID = ID

	slug := object.Properties["slug"]
	if slug != nil {
		product.Slug = slug.(string)
	}

	description := object.Properties["description"]
	if description != nil {
		product.Description = description.(string)
	}

	price := object.Properties["price"]
	if price != nil {
		value, _ := strconv.ParseFloat(price.(string), 32)
		product.Price = float32(value)
	}

	quantity := object.Properties["quantity"]
	if quantity != nil {
		value, _ := strconv.ParseUint(quantity.(string), 10, 32)
		product.Quantity = uint(value)
	}

	image := object.Properties["image"]
	if image != nil {
		product.Image = image.(string)
	}

	version := object.Properties["version"]
	if version != nil {
		value, _ := strconv.ParseUint(version.(string), 10, 32)
		product.Version = uint(value)
	}

	return product, nil
}
