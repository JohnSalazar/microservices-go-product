package nats

import (
	"log"

	mongo_listeners "product/src/nats/listeners/mongo"
	postgres_listeners "product/src/nats/listeners/postgres"
	"product/src/nats/subjects"

	mongo_product_command_handler "product/src/application/commands/product/mongo"
	postgres_product_command_handler "product/src/application/commands/product/postgres"

	mongo_store_command_handler "product/src/application/commands/store/mongo"
	postgres_store_command_handler "product/src/application/commands/store/postgres"

	"github.com/JohnSalazar/microservices-go-common/config"
	common_nats "github.com/JohnSalazar/microservices-go-common/nats"
	common_service "github.com/JohnSalazar/microservices-go-common/services"
	"github.com/nats-io/nats.go"
)

type listen struct {
	js nats.JetStreamContext
}

const queueGroupName string = "products-service"

var (
	subscribe          common_nats.Listener
	commandErrorHelper *common_nats.CommandErrorHelper

	mongoProductCreateCommand *mongo_listeners.ProductCreateCommandListener
	mongoProductUpdateCommand *mongo_listeners.ProductUpdateCommandListener

	postgresStoreCreateCommand *postgres_listeners.StoreCreateCommandListener
	mongoStoreCreateCommand    *mongo_listeners.StoreCreateCommandListener

	postgresStoreBookCommand *postgres_listeners.StoreBookCommandListener
	mongoStoreBookCommand    *mongo_listeners.StoreBookCommandListener

	postgresStoreUnbookCommand *postgres_listeners.StoreUnbookCommandListener
	mongoStoreUnbookCommand    *mongo_listeners.StoreUnbookCommandListener

	postgresStorePaymentCommand *postgres_listeners.StorePaymentCommandListener
	mongoStorePaymentCommand    *mongo_listeners.StorePaymentCommandListener
)

func NewListen(
	config *config.Config,
	js nats.JetStreamContext,
	postgresProductCommandHandler *postgres_product_command_handler.ProductCommandHandler,
	mongoProductCommandHandler *mongo_product_command_handler.ProductCommandHandler,
	postgresStoreCommandHandler *postgres_store_command_handler.StoreCommandHandler,
	mongoStoreCommandHandler *mongo_store_command_handler.StoreCommandHandler,
	email common_service.EmailService,
) *listen {
	subscribe = common_nats.NewListener(js)
	commandErrorHelper = common_nats.NewCommandErrorHelper(config, email)

	mongoProductCreateCommand = mongo_listeners.NewProductCreateCommandListener(mongoProductCommandHandler, email, commandErrorHelper)
	mongoProductUpdateCommand = mongo_listeners.NewProductUpdateCommandListener(mongoProductCommandHandler, email, commandErrorHelper)

	postgresStoreCreateCommand = postgres_listeners.NewStoreCreateCommandListener(postgresStoreCommandHandler, email, commandErrorHelper)
	mongoStoreCreateCommand = mongo_listeners.NewStoreCreateCommandListener(mongoStoreCommandHandler, email, commandErrorHelper)

	postgresStoreBookCommand = postgres_listeners.NewStoreBookCommandListener(postgresStoreCommandHandler, email, commandErrorHelper)
	mongoStoreBookCommand = mongo_listeners.NewStoreBookCommandListener(mongoStoreCommandHandler, email, commandErrorHelper)

	postgresStoreUnbookCommand = postgres_listeners.NewStoreUnbookCommandListener(postgresStoreCommandHandler, email, commandErrorHelper)
	mongoStoreUnbookCommand = mongo_listeners.NewStoreUnbookCommandListener(mongoStoreCommandHandler, email, commandErrorHelper)

	postgresStorePaymentCommand = postgres_listeners.NewStorePaymentCommandListener(postgresStoreCommandHandler, email, commandErrorHelper)
	mongoStorePaymentCommand = mongo_listeners.NewStorePaymentCommandListener(mongoStoreCommandHandler, email, commandErrorHelper)
	return &listen{
		js: js,
	}
}

func (l *listen) Listen() {
	go subscribe.Listener(string(subjects.ProductCreateMongo), queueGroupName, queueGroupName+"_0", mongoProductCreateCommand.ProcessProductCreateCommand())

	go subscribe.Listener(string(subjects.ProductUpdateMongo), queueGroupName, queueGroupName+"_1", mongoProductUpdateCommand.ProcessProductUpdateCommand())

	go subscribe.Listener(string(subjects.StoreCreatePostgres), queueGroupName, queueGroupName+"_2", postgresStoreCreateCommand.ProcessStoreCreateCommand())

	go subscribe.Listener(string(subjects.StoreCreateMongo), queueGroupName, queueGroupName+"_3", mongoStoreCreateCommand.ProcessStoreCreateCommand())

	go subscribe.Listener(string(common_nats.StoreBook), queueGroupName, queueGroupName+"_4", postgresStoreBookCommand.ProcessStoreBookCommand())

	go subscribe.Listener(string(subjects.StoreBookMongo), queueGroupName, queueGroupName+"_5", mongoStoreBookCommand.ProcessStoreBookCommand())

	go subscribe.Listener(string(subjects.StoreUnbookPostgres), queueGroupName, queueGroupName+"_6", postgresStoreUnbookCommand.ProcessStoreUnbookCommand())

	go subscribe.Listener(string(subjects.StoreUnbookMongo), queueGroupName, queueGroupName+"_7", mongoStoreUnbookCommand.ProcessStoreUnbookCommand())

	go subscribe.Listener(string(common_nats.StorePayment), queueGroupName, queueGroupName+"_8", postgresStorePaymentCommand.ProcessStorePaymentCommand())

	go subscribe.Listener(string(subjects.StorePaymentMongo), queueGroupName, queueGroupName+"_9", mongoStorePaymentCommand.ProcessStorePaymentCommand())

	log.Printf("Listener on!!!\n")
}
