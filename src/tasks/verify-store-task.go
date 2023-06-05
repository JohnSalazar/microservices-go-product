package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	postgres_repository_interface "product/src/data/repositories/interfaces"
	"product/src/models"
	"product/src/nats/subjects"
	"time"

	command "product/src/application/commands/store"

	"github.com/google/uuid"
	common_nats "github.com/oceano-dev/microservices-go-common/nats"
	common_service "github.com/oceano-dev/microservices-go-common/services"
	trace "github.com/oceano-dev/microservices-go-common/trace/otel"
)

type VerifyStoreTask interface {
	AddStore(store *models.Store)
	Run()
}

type verifyStoreTask struct {
	storePostgresRepository postgres_repository_interface.StoreRepository
	email                   common_service.EmailService
	publisher               common_nats.Publisher
}

type taskStore struct {
	ID       uuid.UUID `json:"id"`
	BookedAt time.Time `json:"booked_at"`
}

func NewStoreTask(
	storePostgresRepository postgres_repository_interface.StoreRepository,
	email common_service.EmailService,
	publisher common_nats.Publisher,
) *verifyStoreTask {
	return &verifyStoreTask{
		storePostgresRepository: storePostgresRepository,
		email:                   email,
		publisher:               publisher,
	}
}

var stores []*taskStore
var loadedStore bool

func (task *verifyStoreTask) AddStore(store *models.Store) {
	s := &taskStore{ID: store.ID, BookedAt: store.BookedAt}
	stores = append(stores, s)
}

func (task *verifyStoreTask) Run() {
	if !loadedStore {
		task.loadStore()
	}

	ticker := time.NewTicker(2 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				for i := range stores {
					if stores[i].BookedAt.IsZero() {
						stores[i] = nil
					}
					if stores[i] != nil && stores[i].BookedAt.Before(time.Now().UTC()) {
						ctx := context.Background()
						storeCommand := &command.UnbookStoreCommand{
							ID: stores[i].ID,
						}
						data, _ := json.Marshal(storeCommand)
						err := task.publisher.Publish(string(subjects.StoreUnbookPostgres), data)
						if err != nil {
							_, span := trace.NewSpan(ctx, "tasks.VerifyStoreTask")
							defer span.End()
							msg := fmt.Sprintf("error unbooking store %s: %s", stores[i].ID, err.Error())
							trace.FailSpan(span, msg)
							log.Print(msg)
							go task.email.SendSupportMessage(msg)
							ticker.Reset(15 * time.Second)
							return
						}

						stores[i] = nil
					}
				}

				if len(stores) > 0 {
					task.clearStore()
				}

				// fmt.Printf("store success checked %s\n", time.Now().UTC())
				ticker.Reset(5 * time.Second)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (task *verifyStoreTask) loadStore() {
	ctx := context.Background()
	result, err := task.storePostgresRepository.LoadBookedStore(ctx)
	if err != nil {
		log.Printf("error loading stores: %s", err.Error())
		return
	}

	for _, store := range result {
		task.AddStore(store)
	}

	loadedStore = true
}

func (task *verifyStoreTask) clearStore() {
	newStore := []*taskStore{}

	for i := range stores {
		if stores[i] != nil && !stores[i].BookedAt.IsZero() {
			newStore = append(newStore, stores[i])
		}
	}

	stores = newStore
}
