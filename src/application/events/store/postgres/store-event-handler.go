package postgres_event

import (
	"context"
	"encoding/json"
	"fmt"

	common_nats "github.com/JohnSalazar/microservices-go-common/nats"

	command "product/src/application/commands/store"
	events "product/src/application/events/store"
	"product/src/dtos"
	"product/src/nats/subjects"
	"product/src/tasks"
)

type StoreEventHandler struct {
	storeTask tasks.VerifyStoreTask
	publisher common_nats.Publisher
}

func NewStoreEventHandler(
	storeTask tasks.VerifyStoreTask,
	publisher common_nats.Publisher,
) *StoreEventHandler {
	return &StoreEventHandler{
		storeTask: storeTask,
		publisher: publisher,
	}
}

func (store *StoreEventHandler) StoreCreatedEventHandler(ctx context.Context, event *events.StoreCreatedEvent) error {

	// createStoreCommands := []*command.CreateStoreCommand{}
	// for _, store := range event.Stores {
	// 	createStoreMongoCommand := &command.CreateStoreCommand{
	// 		AggregateID: event.AggregateID,
	// 		MessageType: event.MessageType,
	// 		Timestamp:   event.Timestamp,
	// 		ID:          store.ID,
	// 		ProductID:   store.ProductID,
	// 		//Quantity:    0,
	// 		CreatedAt: store.CreatedAt,
	// 	}
	// 	createStoreCommands = append(createStoreCommands, createStoreMongoCommand)
	// }

	data, _ := json.Marshal(event)
	err := store.publisher.Publish(string(subjects.StoreCreateMongo), data)
	if err != nil {
		return err
	}

	return nil
}

func (store *StoreEventHandler) StoreBookedEventHandler(ctx context.Context, event *events.StoreBookedEvent) error {
	bookStoreCommand := &command.BookStoreCommand{
		AggregateID: event.AggregateID,
		MessageType: event.MessageType,
		Timestamp:   event.Timestamp,
		OrderID:     event.OrderID,
		Stores:      event.Stores,
	}

	updateStoreOrderCommand := &dtos.UpdateStoreOrder{
		ID:     event.OrderID,
		Stores: event.Stores,
	}
	dataOrder, _ := json.Marshal(updateStoreOrderCommand)
	err := store.publisher.Publish(string(common_nats.StoreBooked), dataOrder)
	if err != nil {
		fmt.Println(err)
	}

	dataMongo, _ := json.Marshal(bookStoreCommand)
	err = store.publisher.Publish(string(subjects.StoreBookMongo), dataMongo)
	if err != nil {
		return err
	}

	// bookStoreCommands := []*command.BookStoreCommand{}
	for _, _store := range event.Stores {
		// bookStoreMongoCommand := &command.BookStoreCommand{
		// 	AggregateID: event.AggregateID,
		// 	MessageType: event.MessageType,
		// 	Timestamp:   event.Timestamp,
		// 	ID:          store.ID,
		// 	BookedAt:    store.BookedAt,
		// 	Sold:        store.Sold,
		// 	UpdatedAt:   store.UpdatedAt,
		// 	Version:     store.Version,
		// }

		// bookStoreCommands = append(bookStoreCommands, bookStoreMongoCommand)

		store.storeTask.AddStore(_store)
	}

	return nil
}

func (store *StoreEventHandler) StoreUnbookedEventHandler(ctx context.Context, event *events.StoreUnbookedEvent) error {
	// unbookStoreCommands := []*command.UnbookStoreCommand{}
	// for _, _store := range event.Stores {
	// 	unbookStoreMongoCommand := &command.UnbookStoreCommand{
	// 		AggregateID: event.AggregateID,
	// 		MessageType: event.MessageType,
	// 		Timestamp:   event.Timestamp,
	// 		ID:          _store.ID,
	// 		Sold:        _store.Sold,
	// 		BookedAt:    _store.BookedAt,
	// 		UpdatedAt:   _store.UpdatedAt,
	// 		Version:     _store.Version,
	// 	}

	// 	unbookStoreCommands = append(unbookStoreCommands, unbookStoreMongoCommand)
	// 	store.storeTask.AddStore(_store)
	// }

	data, _ := json.Marshal(event)
	err := store.publisher.Publish(string(subjects.StoreUnbookMongo), data)
	if err != nil {
		return err
	}

	return nil
}

func (store *StoreEventHandler) StorePaidEventHandler(ctx context.Context, event *events.StorePaidEvent) error {

	data, _ := json.Marshal(event)
	err := store.publisher.Publish(string(subjects.StorePaymentMongo), data)
	if err != nil {
		return err
	}

	return nil
}
