package postgres_listeners

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	postgres_command "product/src/application/commands/store/postgres"

	command "product/src/application/commands/store"

	"github.com/nats-io/nats.go"
	common_nats "github.com/oceano-dev/microservices-go-common/nats"
	common_service "github.com/oceano-dev/microservices-go-common/services"

	trace "github.com/oceano-dev/microservices-go-common/trace/otel"
)

type StoreCreateCommandListener struct {
	postgresCommandHandler *postgres_command.StoreCommandHandler
	email                  common_service.EmailService
	errorHelper            *common_nats.CommandErrorHelper
}

func NewStoreCreateCommandListener(
	postgresCommandHandler *postgres_command.StoreCommandHandler,
	email common_service.EmailService,
	errorHelper *common_nats.CommandErrorHelper,
) *StoreCreateCommandListener {
	return &StoreCreateCommandListener{
		postgresCommandHandler: postgresCommandHandler,
		email:                  email,
		errorHelper:            errorHelper,
	}
}

func (c *StoreCreateCommandListener) ProcessStoreCreateCommand() nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx := context.Background()
		_, span := trace.NewSpan(ctx, fmt.Sprintf("publish.%s\n", msg.Subject))
		defer span.End()

		storeCommand := &command.CreateStoreCommand{}
		err := json.Unmarshal(msg.Data, storeCommand)
		if c.errorHelper.CheckUnmarshal(msg, err) == nil {
			err = c.postgresCommandHandler.CreateStoreCommandHandler(ctx, storeCommand)
			c.errorHelper.CheckCommandError(span, msg, err)
		}

		err = msg.Ack()
		if err != nil {
			log.Printf("stan msg.Ack error: %v\n", err)
		}
	}
}
