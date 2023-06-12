package mongo_command

import (
	"context"

	commands "product/src/application/commands/store"
	events "product/src/application/events/store"

	mongo_event_handler "product/src/application/events/store/mongo"
	"product/src/data/repositories/interfaces"
	"time"

	common_validator "github.com/JohnSalazar/microservices-go-common/validators"
)

type StoreCommandHandler struct {
	storeMongoRepository interfaces.StoreRepository
	mongoEventHandler    *mongo_event_handler.StoreEventHandler
}

func NewStoreCommandHandler(
	storeMongoRepository interfaces.StoreRepository,
	mongoEventHandler *mongo_event_handler.StoreEventHandler,
) *StoreCommandHandler {
	common_validator.NewValidator("en")
	return &StoreCommandHandler{
		storeMongoRepository: storeMongoRepository,
		mongoEventHandler:    mongoEventHandler,
	}
}

func (store *StoreCommandHandler) CreateStoreCommandHandler(ctx context.Context, command *commands.CreateStoreCommand) error {

	// stores := []*models.Store{}

	// for _, store := range commands.Stores {
	// 	storeModel := &models.Store{
	// 		ID:        store.ID,
	// 		ProductID: store.ProductID,
	// 		CreatedAt: store.CreatedAt,
	// 	}
	// 	stores = append(stores, storeModel)
	// }

	err := store.storeMongoRepository.Create(ctx, command.Stores)
	if err != nil {
		return err
	}

	storeEvent := &events.StoreCreatedEvent{
		AggregateID: command.ProductID,
		MessageType: command.MessageType,
		Timestamp:   time.Now().UTC(),
		Stores:      command.Stores,
	}

	go store.mongoEventHandler.StoreCreatedEventHandler(storeEvent)

	return nil
}

func (store *StoreCommandHandler) BookStoreCommandHandler(ctx context.Context, command *commands.BookStoreCommand) error {

	// stores := []*models.Store{}

	// for _, store := range command.Stores {
	// 	storeModel := &models.Store{
	// 		ID:        store.ID,
	// 		BookedAt:  store.BookedAt,
	// 		Sold:      store.Sold,
	// 		UpdatedAt: store.UpdatedAt,
	// 		Version:   store.Version,
	// 	}
	// 	stores = append(stores, storeModel)
	// }

	stores, err := store.storeMongoRepository.Update(ctx, command.Stores)
	if err != nil {
		return err
	}

	storeEvent := &events.StoreBookedEvent{
		AggregateID: command.AggregateID,
		MessageType: command.MessageType,
		Timestamp:   time.Now().UTC(),
		Stores:      stores,
	}

	go store.mongoEventHandler.StoreBookedEventHandler(storeEvent)

	return nil
}

func (store *StoreCommandHandler) UnbookStoreCommandHandler(ctx context.Context, command *commands.UnbookStoreCommand) error {

	// stores := []*models.Store{}

	// for _, store := range commands {
	// 	storeModel := &models.Store{
	// 		ID:        store.ID,
	// 		Sold:      store.Sold,
	// 		BookedAt:  store.BookedAt,
	// 		UpdatedAt: store.UpdatedAt,
	// 		Version:   store.Version,
	// 	}
	// 	stores = append(stores, storeModel)
	// }

	stores, err := store.storeMongoRepository.Update(ctx, command.Stores)
	if err != nil {
		return err
	}

	storeEvent := &events.StoreUnbookedEvent{
		AggregateID: command.ID,
		MessageType: command.MessageType,
		Timestamp:   time.Now().UTC(),
		Stores:      stores,
	}

	go store.mongoEventHandler.StoreUnbookedEventHandler(storeEvent)

	return nil
}

func (store *StoreCommandHandler) PaymentStoreCommandHandler(ctx context.Context, command *commands.PaymentStoreCommand) error {

	// stores := []*models.Store{}

	// for _, store := range commands.Stores {
	// 	storeModel := &models.Store{
	// 		ID:        store.ID,
	// 		Sold:      store.Sold,
	// 		UpdatedAt: store.UpdatedAt,
	// 		Version:   store.Version,
	// 	}
	// 	stores = append(stores, storeModel)
	// }

	stores, err := store.storeMongoRepository.Update(ctx, command.Stores)
	if err != nil {
		return err
	}

	storeEvent := &events.StorePaidEvent{
		AggregateID: command.AggregateID,
		MessageType: command.MessageType,
		Timestamp:   time.Now().UTC(),
		Stores:      stores,
	}

	go store.mongoEventHandler.StorePaidEventHandler(storeEvent)

	return nil
}
