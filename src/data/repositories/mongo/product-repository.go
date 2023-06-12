package mongo_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"product/src/models"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JohnSalazar/microservices-go-common/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type productRepository struct {
	database *mongo.Database
}

func NewProductRepository(
	database *mongo.Database,
) *productRepository {
	return &productRepository{
		database: database,
	}
}

func (r *productRepository) collectionName() string {
	return "products"
}

func (r *productRepository) collection() *mongo.Collection {
	return r.database.Collection(r.collectionName())
}

func (r *productRepository) aggregate(ctx context.Context, pipeline interface{}) (*models.Product, error) {
	cursor, err := r.collection().Aggregate(ctx, pipeline)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	products := []*models.Product{}
	product := &models.Product{}

	for cursor.Next(ctx) {
		object := map[string]interface{}{}

		err = cursor.Decode(object)
		if err != nil {
			return nil, err
		}

		product, err = r.mapProduct(object)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	product = nil

	if len(products) > 0 {
		product = products[0]
	}

	return product, nil
}

func (r *productRepository) find(ctx context.Context, filter interface{}, page int, size int) ([]*models.Product, error) {
	findOptions := options.FindOptions{}
	findOptions.SetSort(bson.M{"name": 1})

	page64 := int64(page)
	size64 := int64(size)
	findOptions.SetSkip((page64 - 1) * size64)
	findOptions.SetLimit(size64)

	newFilter := map[string]interface{}{
		"deleted": false,
	}
	mergeFilter := helpers.MergeFilters(newFilter, filter)

	cursor, err := r.collection().Find(ctx, mergeFilter, &findOptions)
	if err != nil {
		defer cursor.Close(ctx)
		return nil, err
	}

	products := []*models.Product{}

	for cursor.Next(ctx) {
		object := map[string]interface{}{}

		err = cursor.Decode(object)
		if err != nil {
			return nil, err
		}

		product, err := r.mapProduct(object)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	return products, nil
}

func (r *productRepository) findOne(ctx context.Context, filter interface{}) (*models.Product, error) {
	findOneOptions := options.FindOneOptions{}
	findOneOptions.SetSort(bson.M{"version": -1})

	newFilter := map[string]interface{}{
		"deleted": false,
	}
	mergeFilter := helpers.MergeFilters(newFilter, filter)

	object := map[string]interface{}{}
	err := r.collection().FindOne(ctx, mergeFilter, &findOneOptions).Decode(object)
	if err != nil {
		return nil, err
	}

	product, err := r.mapProduct(object)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) findOneAndUpdate(ctx context.Context, filter interface{}, fields interface{}) *mongo.SingleResult {
	findOneAndUpdateOptions := options.FindOneAndUpdateOptions{}
	findOneAndUpdateOptions.SetReturnDocument(options.After)

	result := r.collection().FindOneAndUpdate(ctx, filter, bson.M{"$set": fields}, &findOneAndUpdateOptions)

	return result
}

func (r *productRepository) GetAll(ctx context.Context, name string, page int, size int) ([]*models.Product, error) {
	// filter := bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: name, Options: "i"}}}

	filter := map[string]interface{}{}
	if len(strings.TrimSpace(name)) > 0 {
		filter = bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: name, Options: "i"}}}
	}

	return r.find(ctx, filter, page, size)
}

func (r *productRepository) FindByID(ctx context.Context, ID uuid.UUID) (*models.Product, error) {
	filter := bson.M{"_id": ID.String()}

	return r.findOne(ctx, filter)
}

func (r *productRepository) FindBySlug(ctx context.Context, slug string) (*models.Product, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{"slug": slug, "deleted": false},
		},
		{
			"$lookup": bson.M{
				"from":         "stores",
				"localField":   "_id",
				"foreignField": "product_id",
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"deleted": false,
							"sold":    false,
							"booked_at": bson.M{
								"$lte": time.Now().UTC(),
							},
						},
					},
				},
				"as": "quantity",
			},
		},
		{
			"$addFields": bson.M{
				"quantity": bson.M{"$size": "$quantity"},
			},
		},
	}

	return r.aggregate(ctx, pipeline)
}

func (r *productRepository) FindByName(ctx context.Context, name string) (*models.Product, error) {
	filter := bson.M{"name": name}

	return r.findOne(ctx, filter)
}

func (r *productRepository) Create(ctx context.Context, product *models.Product) (*models.Product, error) {
	product.CreatedAt = time.Now().UTC()

	fields := bson.M{
		"_id":         product.ID.String(),
		"name":        product.Name,
		"slug":        product.Slug,
		"description": product.Description,
		"price":       product.Price,
		"image":       product.Image,
		"created_at":  product.CreatedAt,
		"updated_at":  product.UpdatedAt,
		"version":     product.Version,
		"deleted":     product.Deleted,
	}

	_, err := r.collection().InsertOne(ctx, fields)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) Update(ctx context.Context, product *models.Product) (*models.Product, error) {
	// product.Version++
	product.UpdatedAt = time.Now().UTC()

	fields := bson.M{
		"name":        product.Name,
		"slug":        product.Slug,
		"description": product.Description,
		"price":       product.Price,
		"image":       product.Image,
		"updated_at":  product.UpdatedAt,
		"version":     product.Version,
	}

	filter := r.filterUpdate(product)

	result := r.findOneAndUpdate(ctx, filter, fields)
	if result.Err() != nil {
		return nil, result.Err()
	}

	object := map[string]interface{}{}
	err := result.Decode(object)
	if err != nil {
		return nil, err
	}

	modelProduct, err := r.mapProduct(object)
	if err != nil {
		return nil, err
	}

	return modelProduct, err
}

func (r *productRepository) Delete(ctx context.Context, ID uuid.UUID) error {
	filter := bson.M{"_id": ID.String()}

	fields := bson.M{"deleted": true}

	result := r.findOneAndUpdate(ctx, filter, fields)
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}

func (r *productRepository) filterUpdate(product *models.Product) interface{} {
	filter := bson.M{
		"_id": product.ID.String(),
		//"version": product.Version - 1,
		"deleted": false,
	}

	return filter
}

func (r *productRepository) mapProduct(object map[string]interface{}) (*models.Product, error) {
	jsonStr, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	var product models.Product
	if err := json.Unmarshal(jsonStr, &product); err != nil {
		return nil, err
	}

	ID, err := uuid.Parse(object["_id"].(string))
	if err != nil {
		return nil, err
	}

	product.ID = ID

	return &product, nil
}
