package mongo_listeners

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	command "product/src/application/commands/product"
	mongo_command_handler "product/src/application/commands/product/mongo"

	"github.com/nats-io/nats.go"
	common_nats "github.com/oceano-dev/microservices-go-common/nats"
	common_service "github.com/oceano-dev/microservices-go-common/services"

	trace "github.com/oceano-dev/microservices-go-common/trace/otel"
)

type ProductCreateCommandListener struct {
	mongoProductCommandHandler *mongo_command_handler.ProductCommandHandler
	email                      common_service.EmailService
	errorHelper                *common_nats.CommandErrorHelper
}

func NewProductCreateCommandListener(
	mongoProductCommandHandler *mongo_command_handler.ProductCommandHandler,
	email common_service.EmailService,
	errorHelper *common_nats.CommandErrorHelper,
) *ProductCreateCommandListener {
	return &ProductCreateCommandListener{
		mongoProductCommandHandler: mongoProductCommandHandler,
		email:                      email,
		errorHelper:                errorHelper,
	}
}

func (c *ProductCreateCommandListener) ProcessProductCreateCommand() nats.MsgHandler {
	return func(msg *nats.Msg) {
		ctx := context.Background()
		_, span := trace.NewSpan(ctx, fmt.Sprintf("publish.%s\n", msg.Subject))
		defer span.End()

		productCommand := &command.CreateProductCommand{}
		err := json.Unmarshal(msg.Data, productCommand)
		if c.errorHelper.CheckUnmarshal(msg, err) == nil {
			err = c.mongoProductCommandHandler.CreateProductCommandHandler(ctx, productCommand)
			c.errorHelper.CheckCommandError(span, msg, err)
		}

		err = msg.Ack()
		if err != nil {
			log.Printf("stan msg.Ack error: %v", err)
		}
	}
}
