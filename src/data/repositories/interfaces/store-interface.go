package interfaces

import (
	"context"
	"product/src/models"

	"github.com/google/uuid"
)

type StoreRepository interface {
	LoadBookedStore(ctx context.Context) ([]*models.Store, error)
	FindByID(ctx context.Context, ID uuid.UUID) (*models.Store, error)
	Book(ctx context.Context, productID uuid.UUID, quantity uint) ([]*models.Store, error)
	Create(ctx context.Context, stores []*models.Store) error
	Update(ctx context.Context, stores []*models.Store) ([]*models.Store, error)
	Delete(ctx context.Context, ID uuid.UUID) error
}
