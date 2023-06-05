package interfaces

import (
	"context"
	"product/src/models"

	"github.com/google/uuid"
)

type ProductRepository interface {
	GetAll(ctx context.Context, name string, page int, size int) ([]*models.Product, error)
	FindByID(ctx context.Context, ID uuid.UUID) (*models.Product, error)
	FindBySlug(ctx context.Context, slug string) (*models.Product, error)
	FindByName(ctx context.Context, name string) (*models.Product, error)
	Create(ctx context.Context, product *models.Product) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) (*models.Product, error)
	Delete(ctx context.Context, ID uuid.UUID) error
}
