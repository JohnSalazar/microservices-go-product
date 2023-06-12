package mongo_listeners

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	mongo_command "product/src/application/commands/store/mongo"

	command "product/src/application/commands/store"

	common_nats "github.com/JohnSalazar/microservices-go-common/nats"
	common_service "github.com/JohnSalazar/microservices-go-common/services"
	trace "github.com/JohnSalazar/microservices-go-common/trace/otel"
	"github.com/nats-io/nats.go"
)

type StorePaymentCommandListener struct {
	mongoCommandHandler *mongo_command.StoreCommandHandler
	email               common_service.EmailService
	errorHelper         *common_nats.CommandErrorHelper
}

func NewStorePaymentCommandListener(
	mongoCommandHandler *mongo_command.StoreCommandHandler,
	email common_service.EmailService,
	errorHelper *common_nats.CommandErrorHelper,
) *StorePaymentCommandListener {
	return &StorePaymentCommandListener{
		mongoCommandHandler: mongoCommandHandler,
		email:               email,
		errorHelper:         errorHelper,
	}
}

func (c *StorePaymentCommandListener) ProcessStorePaymentCommand() nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx := context.Background()
		_, span := trace.NewSpan(ctx, fmt.Sprintf("publish.%s\n", msg.Subject))
		defer span.End()

		storeCommand := &command.PaymentStoreCommand{}
		err := json.Unmarshal(msg.Data, &storeCommand)
		if c.errorHelper.CheckUnmarshal(msg, err) == nil {
			err = c.mongoCommandHandler.PaymentStoreCommandHandler(ctx, storeCommand)
			c.errorHelper.CheckCommandError(span, msg, err)
		}

		err = msg.Ack()
		if err != nil {
			log.Printf("stan msg.Ack error: %v\n", err)
		}
	}
}
