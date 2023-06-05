package mongo_listeners

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	command "product/src/application/commands/store"
	mongo_command_handler "product/src/application/commands/store/mongo"

	"github.com/nats-io/nats.go"
	common_nats "github.com/oceano-dev/microservices-go-common/nats"
	common_service "github.com/oceano-dev/microservices-go-common/services"
	trace "github.com/oceano-dev/microservices-go-common/trace/otel"
)

type StoreCreateCommandListener struct {
	mongoStoreCommandHandler *mongo_command_handler.StoreCommandHandler
	email                    common_service.EmailService
	errorHelper              *common_nats.CommandErrorHelper
}

func NewStoreCreateCommandListener(
	mongoStoreCommandHandler *mongo_command_handler.StoreCommandHandler,
	email common_service.EmailService,
	errorHelper *common_nats.CommandErrorHelper,
) *StoreCreateCommandListener {
	return &StoreCreateCommandListener{
		mongoStoreCommandHandler: mongoStoreCommandHandler,
		email:                    email,
		errorHelper:              errorHelper,
	}
}

func (c *StoreCreateCommandListener) ProcessStoreCreateCommand() nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx := context.Background()
		_, span := trace.NewSpan(ctx, fmt.Sprintf("publish.%s\n", msg.Subject))
		defer span.End()

		storeCommand := &command.CreateStoreCommand{}
		err := json.Unmarshal(msg.Data, &storeCommand)
		if c.errorHelper.CheckUnmarshal(msg, err) == nil {
			err = c.mongoStoreCommandHandler.CreateStoreCommandHandler(ctx, storeCommand)
			c.errorHelper.CheckCommandError(span, msg, err)
		}

		err = msg.Ack()
		if err != nil {
			log.Printf("stan msg.Ack error: %v\n", err)
		}
	}
}
