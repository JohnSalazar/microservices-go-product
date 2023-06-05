package mongo_listeners

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	mongo_command "product/src/application/commands/store/mongo"

	command "product/src/application/commands/store"

	"github.com/nats-io/nats.go"
	common_nats "github.com/oceano-dev/microservices-go-common/nats"
	common_service "github.com/oceano-dev/microservices-go-common/services"
	trace "github.com/oceano-dev/microservices-go-common/trace/otel"
)

type StoreUnbookCommandListener struct {
	mongoCommandHandler *mongo_command.StoreCommandHandler
	email               common_service.EmailService
	errorHelper         *common_nats.CommandErrorHelper
}

func NewStoreUnbookCommandListener(
	mongoCommandHandler *mongo_command.StoreCommandHandler,
	email common_service.EmailService,
	errorHelper *common_nats.CommandErrorHelper,
) *StoreUnbookCommandListener {
	return &StoreUnbookCommandListener{
		mongoCommandHandler: mongoCommandHandler,
		email:               email,
		errorHelper:         errorHelper,
	}
}

func (c *StoreUnbookCommandListener) ProcessStoreUnbookCommand() nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx := context.Background()
		_, span := trace.NewSpan(ctx, fmt.Sprintf("publish.%s\n", msg.Subject))
		defer span.End()

		storeCommand := &command.UnbookStoreCommand{}
		err := json.Unmarshal(msg.Data, &storeCommand)
		if c.errorHelper.CheckUnmarshal(msg, err) == nil {
			err = c.mongoCommandHandler.UnbookStoreCommandHandler(ctx, storeCommand)
			c.errorHelper.CheckCommandError(span, msg, err)
		}

		err = msg.Ack()
		if err != nil {
			log.Printf("stan msg.Ack error: %v\n", err)
		}
	}
}
