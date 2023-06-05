package interfaces

import (
	"context"
	"product/src/models"
)

type EventSourcingRepository interface {
	Create(ctx context.Context, eventStore *models.EventSourcing) error
	CreateMany(ctx context.Context, eventStore []*models.EventSourcing) error
}
