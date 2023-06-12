package postgres_command

import (
	"context"
	"errors"
	"fmt"
	commands "product/src/application/commands/store"
	events "product/src/application/events/store"
	postgres_event_handler "product/src/application/events/store/postgres"
	"strings"

	repository_interface "product/src/data/repositories/interfaces"

	"product/src/dtos"
	"product/src/models"
	"product/src/validators"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"

	common_models "github.com/JohnSalazar/microservices-go-common/models"
	common_nats "github.com/JohnSalazar/microservices-go-common/nats"
	common_validator "github.com/JohnSalazar/microservices-go-common/validators"
)

type StoreCommandHandler struct {
	storePostgresRepository      repository_interface.StoreRepository
	eventSourcingMongoRepository repository_interface.EventSourcingRepository
	postgresEventHandler         *postgres_event_handler.StoreEventHandler
	publisher                    common_nats.Publisher
}

func NewStoreCommandHandler(
	storePostgresRepository repository_interface.StoreRepository,
	eventSourcingMongoRepository repository_interface.EventSourcingRepository,
	postgresEventHandler *postgres_event_handler.StoreEventHandler,
	publisher common_nats.Publisher,
) *StoreCommandHandler {
	common_validator.NewValidator("en")
	return &StoreCommandHandler{
		storePostgresRepository:      storePostgresRepository,
		eventSourcingMongoRepository: eventSourcingMongoRepository,
		postgresEventHandler:         postgresEventHandler,
		publisher:                    publisher,
	}
}

func (store *StoreCommandHandler) CreateStoreCommandHandler(ctx context.Context, command *commands.CreateStoreCommand) error {
	storeDto := &dtos.AddStore{
		ProductID: command.ProductID,
		Quantity:  command.Quantity,
	}

	result := validators.ValidateAddStore(storeDto)
	if result != nil {
		return errors.New(strings.Join(result.([]string), ""))
	}

	stores := []*models.Store{}
	eventsSourcing := []*models.EventSourcing{}

	for i := 0; i < int(storeDto.Quantity); i++ {
		storeModel := &models.Store{
			ID:        uuid.New(),
			ProductID: storeDto.ProductID,
			BookedAt:  time.Time{},
			Sold:      false,
			CreatedAt: time.Now().UTC(),
			Version:   0,
			Deleted:   false,
		}
		stores = append(stores, storeModel)

		data, _ := json.Marshal(storeModel)
		eventSourcing := &models.EventSourcing{
			ID:          uuid.New(),
			AggregateID: storeDto.ProductID,
			MessageType: "store.create",
			Timestamp:   time.Now().UTC(),
			Data:        string(data),
		}
		eventsSourcing = append(eventsSourcing, eventSourcing)
	}

	err := store.storePostgresRepository.Create(ctx, stores)
	if err != nil {
		return err
	}

	go store.eventSourcingMongoRepository.CreateMany(ctx, eventsSourcing)

	storeEvent := &events.StoreCreatedEvent{
		AggregateID: command.ProductID,
		MessageType: eventsSourcing[0].MessageType,
		Timestamp:   eventsSourcing[0].Timestamp,
		ProductID:   command.ProductID,
		Quantity:    command.Quantity,
		CreatedAt:   stores[0].CreatedAt,
		Stores:      stores,
	}

	go store.postgresEventHandler.StoreCreatedEventHandler(ctx, storeEvent)

	return nil
}

func (store *StoreCommandHandler) BookStoreCommandHandler(ctx context.Context, command *commands.BookStoreCommand) error {
	bookStoreDto := &dtos.BookStore{
		Products: command.Products,
	}

	result := validators.ValidateBookStore(bookStoreDto)
	if result != nil {
		return errors.New(strings.Join(result.([]string), ""))
	}

	updateStatusOrder := &dtos.UpdateStatusOrder{
		ID:       command.OrderID,
		Status:   uint(common_models.OrderCanceled),
		StatusAt: time.Now().UTC(),
	}
	dataUpdateStatusOrder, _ := json.Marshal(updateStatusOrder)

	eventsSourcing := []*models.EventSourcing{}
	listStores := []*models.Store{}
	for _, product := range command.Products {
		stores, err := store.storePostgresRepository.Book(ctx, product.ID, product.Quantity)
		if err != nil {
			return err
		}

		if len(stores) != int(product.Quantity) {
			go store.publisher.Publish(string(common_nats.OrderStatus), dataUpdateStatusOrder)
			return errors.New("not enough stores")
		}

		for _, store := range stores {
			store.BookedAt = time.Now().UTC().Add(1 * time.Minute)
			store.Version++
			store.UpdatedAt = time.Now().UTC()
		}

		stores, err = store.storePostgresRepository.Update(ctx, stores)
		if err != nil {
			go store.publisher.Publish(string(common_nats.OrderStatus), dataUpdateStatusOrder)
			return err
		}

		listStores = append(listStores, stores...)

		for _, store := range stores {
			data, _ := json.Marshal(store)
			eventSourcing := &models.EventSourcing{
				ID:          uuid.New(),
				AggregateID: store.ProductID,
				MessageType: "store.book",
				Timestamp:   time.Now().UTC(),
				Data:        string(data),
			}
			eventsSourcing = append(eventsSourcing, eventSourcing)
		}
	}

	go store.eventSourcingMongoRepository.CreateMany(ctx, eventsSourcing)

	storeEvent := &events.StoreBookedEvent{
		AggregateID: uuid.New(),
		MessageType: eventsSourcing[0].MessageType,
		Timestamp:   eventsSourcing[0].Timestamp,
		OrderID:     command.OrderID,
		Stores:      listStores,
	}

	go store.postgresEventHandler.StoreBookedEventHandler(ctx, storeEvent)

	return nil
}

func (store *StoreCommandHandler) UnbookStoreCommandHandler(ctx context.Context, command *commands.UnbookStoreCommand) error {
	if command.ID == uuid.Nil {
		return nil
	}

	unbookStoreDto := &dtos.UnbookStore{
		ID: command.ID,
	}

	result := validators.ValidateUnbookStore(unbookStoreDto)
	if result != nil {
		return errors.New(strings.Join(result.([]string), ""))
	}

	_store, err := store.storePostgresRepository.FindByID(ctx, unbookStoreDto.ID)
	if err != nil {
		return err
	}

	_store.BookedAt = time.Time{}
	_store.Version++
	_store.UpdatedAt = time.Now().UTC()

	stores := []*models.Store{_store}

	stores, err = store.storePostgresRepository.Update(ctx, stores)
	if err != nil {
		return err
	}

	eventsSourcing := []*models.EventSourcing{}
	for _, store := range stores {
		data, _ := json.Marshal(store)
		eventSourcing := &models.EventSourcing{
			ID:          uuid.New(),
			AggregateID: store.ProductID,
			MessageType: "store.unbook",
			Timestamp:   time.Now().UTC(),
			Data:        string(data),
		}
		eventsSourcing = append(eventsSourcing, eventSourcing)
	}

	go store.eventSourcingMongoRepository.CreateMany(ctx, eventsSourcing)

	storeEvent := &events.StoreUnbookedEvent{
		AggregateID: command.ID,
		MessageType: eventsSourcing[0].MessageType,
		Timestamp:   eventsSourcing[0].Timestamp,
		Stores:      stores,
	}

	go store.postgresEventHandler.StoreUnbookedEventHandler(ctx, storeEvent)

	return nil
}

func (store *StoreCommandHandler) PaymentStoreCommandHandler(ctx context.Context, command *commands.PaymentStoreCommand) ([]*models.Store, error) {

	eventsSourcing := []*models.EventSourcing{}
	listStores := []*models.Store{}

	for _, myStore := range command.Stores {
		paymentStoreDto := &dtos.PaymentStore{
			ID:   myStore.ID,
			Sold: true,
		}

		result := validators.ValidatePaymentStore(paymentStoreDto)
		if result != nil {
			return nil, errors.New(strings.Join(result.([]string), ""))
		}

		_store, err := store.storePostgresRepository.FindByID(ctx, paymentStoreDto.ID)
		if err != nil {
			return nil, err
		}

		if _store == nil {
			return nil, fmt.Errorf("store id: %v not found", paymentStoreDto.ID)
			// return errors.New(fmt.Sprintf("store id: %s not found", paymentStoreDto.ID))
		}

		_store.BookedAt = time.Time{}
		_store.Sold = paymentStoreDto.Sold
		_store.Version++
		_store.UpdatedAt = time.Now().UTC()

		listStores = append(listStores, _store)
	}

	stores, err := store.storePostgresRepository.Update(ctx, listStores)
	if err != nil {
		return nil, err
	}

	for _, store := range stores {
		data, _ := json.Marshal(store)
		eventSourcing := &models.EventSourcing{
			ID:          uuid.New(),
			AggregateID: store.ProductID,
			MessageType: "store.payment",
			Timestamp:   time.Now().UTC(),
			Data:        string(data),
		}
		eventsSourcing = append(eventsSourcing, eventSourcing)
	}

	go store.eventSourcingMongoRepository.CreateMany(ctx, eventsSourcing)

	storeEvent := &events.StorePaidEvent{
		AggregateID: uuid.New(),
		MessageType: eventsSourcing[0].MessageType,
		Timestamp:   eventsSourcing[0].Timestamp,
		Stores:      stores,
	}

	go store.postgresEventHandler.StorePaidEventHandler(ctx, storeEvent)

	return stores, nil
}
