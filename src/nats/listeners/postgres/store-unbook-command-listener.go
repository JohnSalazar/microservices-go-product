package postgres_listeners

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	postgres_command "product/src/application/commands/store/postgres"

	command "product/src/application/commands/store"

	common_nats "github.com/JohnSalazar/microservices-go-common/nats"
	common_service "github.com/JohnSalazar/microservices-go-common/services"
	"github.com/nats-io/nats.go"

	trace "github.com/JohnSalazar/microservices-go-common/trace/otel"
)

type StoreUnbookCommandListener struct {
	postgresCommandHandler *postgres_command.StoreCommandHandler
	email                  common_service.EmailService
	errorHelper            *common_nats.CommandErrorHelper
}

func NewStoreUnbookCommandListener(
	postgresCommandHandler *postgres_command.StoreCommandHandler,
	email common_service.EmailService,
	errorHelper *common_nats.CommandErrorHelper,
) *StoreUnbookCommandListener {
	return &StoreUnbookCommandListener{
		postgresCommandHandler: postgresCommandHandler,
		email:                  email,
		errorHelper:            errorHelper,
	}
}

func (c *StoreUnbookCommandListener) ProcessStoreUnbookCommand() nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx := context.Background()
		_, span := trace.NewSpan(ctx, fmt.Sprintf("publish.%s\n", msg.Subject))
		defer span.End()

		storeCommand := &command.UnbookStoreCommand{}
		err := json.Unmarshal(msg.Data, storeCommand)
		if c.errorHelper.CheckUnmarshal(msg, err) == nil {
			err = c.postgresCommandHandler.UnbookStoreCommandHandler(ctx, storeCommand)
			c.errorHelper.CheckCommandError(span, msg, err)
		}

		err = msg.Ack()
		if err != nil {
			log.Printf("stan msg.Ack error: %v\n", err)
		}
	}
}
