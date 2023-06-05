package mongo_event

import (
	events "product/src/application/events/store"
)

type StoreEventHandler struct{}

func NewStoreEventHandler() *StoreEventHandler {
	return &StoreEventHandler{}
}

func (store *StoreEventHandler) StoreCreatedEventHandler(event *events.StoreCreatedEvent) error {

	//fmt.Println(event)

	return nil
}

func (store *StoreEventHandler) StoreBookedEventHandler(event *events.StoreBookedEvent) error {

	//fmt.Println(event)

	return nil
}

func (store *StoreEventHandler) StoreUnbookedEventHandler(event *events.StoreUnbookedEvent) error {

	//fmt.Println(event)

	return nil
}

func (store *StoreEventHandler) StorePaidEventHandler(event *events.StorePaidEvent) error {

	//fmt.Println(event)

	return nil
}
