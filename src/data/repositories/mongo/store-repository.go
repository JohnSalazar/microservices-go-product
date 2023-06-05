package mongo_repository

import (
	"context"
	"encoding/json"
	"errors"
	"product/src/models"
	"time"

	"github.com/google/uuid"

	"github.com/oceano-dev/microservices-go-common/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type storeRepository struct {
	database *mongo.Database
}

func NewStoreRepository(
	database *mongo.Database,
) *storeRepository {
	return &storeRepository{
		database: database,
	}
}

func (r *storeRepository) collectionName() string {
	return "stores"
}

func (r *storeRepository) collection() *mongo.Collection {
	return r.database.Collection(r.collectionName())
}

func (r *storeRepository) find(ctx context.Context, filter interface{}, quantity uint) ([]*models.Store, error) {
	findOptions := options.FindOptions{}
	findOptions.SetSort(bson.M{"create_at": 1})
	findOptions.SetLimit(int64(quantity))

	newFilter := map[string]interface{}{
		"deleted": false,
	}

	mergeFilter := helpers.MergeFilters(newFilter, filter)

	cursor, err := r.collection().Find(ctx, mergeFilter, &findOptions)
	if err != nil {
		defer cursor.Close(ctx)
		return nil, err
	}

	stores := []*models.Store{}

	for cursor.Next(ctx) {
		object := map[string]interface{}{}

		err = cursor.Decode(object)
		if err != nil {
			return nil, err
		}

		store, err := r.mapStore(object)
		if err != nil {
			return nil, err
		}

		stores = append(stores, store)
	}

	return stores, nil
}

func (r *storeRepository) findOne(ctx context.Context, filter interface{}) (*models.Store, error) {
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

	store, err := r.mapStore(object)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (r *storeRepository) findOneAndUpdate(ctx context.Context, filter interface{}, fields interface{}) *mongo.SingleResult {
	findOneAndUpdateOptions := options.FindOneAndUpdateOptions{}
	findOneAndUpdateOptions.SetReturnDocument(options.After)

	result := r.collection().FindOneAndUpdate(ctx, filter, bson.M{"$set": fields}, &findOneAndUpdateOptions)

	return result
}

func (r *storeRepository) LoadBookedStore(ctx context.Context) ([]*models.Store, error) {
	filter := bson.M{
		"booked_at": bson.M{"$gte": time.Now().UTC()},
		"sold":      false,
		"deleted":   false,
	}

	return r.find(ctx, filter, 0)
}

func (r *storeRepository) FindByID(ctx context.Context, ID uuid.UUID) (*models.Store, error) {
	filter := bson.M{"_id": ID.String()}

	return r.findOne(ctx, filter)
}

func (r *storeRepository) Book(ctx context.Context, productID uuid.UUID, quantity uint) ([]*models.Store, error) {
	// filter := map[string]interface{}{
	// 	"product_id": productID.String(),
	// 	"sold":       false,
	// 	"booked_at":  bson.M{"$lte": time.Now().UTC()},
	// }

	// return r.find(ctx, filter, quantity)

	return nil, errors.New("not implemented")
}

func (r *storeRepository) Create(ctx context.Context, stores []*models.Store) error {
	var docs []interface{}
	for _, store := range stores {
		doc := bson.M{
			"_id":        store.ID.String(),
			"product_id": store.ProductID.String(),
			"created_at": store.CreatedAt,
			"booked_at":  time.Time{},
			"sold":       false,
			"version":    0,
			"deleted":    false,
		}

		docs = append(docs, doc)
	}

	_, err := r.collection().InsertMany(context.TODO(), docs)
	if err != nil {
		return err
	}

	return nil
}

func (r *storeRepository) Update(ctx context.Context, stores []*models.Store) ([]*models.Store, error) {
	models := []mongo.WriteModel{}
	for _, store := range stores {
		model := mongo.NewUpdateOneModel()
		model.SetFilter(bson.M{
			"_id": store.ID.String(),
			//"version": store.Version - 1,
			"deleted": false,
		})
		model.SetUpdate(bson.M{
			"$set": bson.M{
				"booked_at":  store.BookedAt,
				"sold":       store.Sold,
				"updated_at": store.UpdatedAt,
				"version":    store.Version,
			},
		})

		models = append(models, model)
	}

	opts := options.BulkWrite().SetOrdered(true)
	_, err := r.collection().BulkWrite(context.TODO(), models, opts)
	if err != nil {
		return nil, err
	}

	return stores, nil
}

func (r *storeRepository) Delete(ctx context.Context, ID uuid.UUID) error {
	filter := bson.M{"_id": ID.String()}

	fields := bson.M{"deleted": true}

	result := r.findOneAndUpdate(ctx, filter, fields)
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}

func (r *storeRepository) mapStore(object map[string]interface{}) (*models.Store, error) {
	jsonStr, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	var store models.Store
	if err := json.Unmarshal(jsonStr, &store); err != nil {
		return nil, err
	}

	ID, err := uuid.Parse(object["_id"].(string))
	if err != nil {
		return nil, err
	}
	store.ID = ID

	productID, err := uuid.Parse(object["product_id"].(string))
	if err != nil {
		return nil, err
	}
	store.ProductID = productID

	return &store, nil
}
