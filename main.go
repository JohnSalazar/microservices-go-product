package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"product/src/controllers"
	"product/src/decorators"
	product_nats "product/src/nats"
	"product/src/nats/subjects"

	"product/src/routers"
	"product/src/tasks"
	"syscall"
	"time"

	mongo_repository "product/src/data/repositories/mongo"
	postgres_repository "product/src/data/repositories/postgres"
	redis_repository "product/src/data/repositories/redis"

	"github.com/JohnSalazar/microservices-go-common/config"
	common_grpc_client "github.com/JohnSalazar/microservices-go-common/grpc/email/client"
	"github.com/JohnSalazar/microservices-go-common/helpers"
	"github.com/JohnSalazar/microservices-go-common/httputil"
	"github.com/JohnSalazar/microservices-go-common/middlewares"
	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/mongo"

	common_log "github.com/JohnSalazar/microservices-go-common/logs"
	common_nats "github.com/JohnSalazar/microservices-go-common/nats"
	common_repositories "github.com/JohnSalazar/microservices-go-common/repositories"
	common_security "github.com/JohnSalazar/microservices-go-common/security"
	common_services "github.com/JohnSalazar/microservices-go-common/services"
	common_tasks "github.com/JohnSalazar/microservices-go-common/tasks"
	common_validator "github.com/JohnSalazar/microservices-go-common/validators"

	provider "github.com/JohnSalazar/microservices-go-common/trace/otel/jaeger"

	migrate "product/src/migrations"

	mongo_product_command_handler "product/src/application/commands/product/mongo"
	postgres_product_command_handler "product/src/application/commands/product/postgres"
	mongo_store_command_handler "product/src/application/commands/store/mongo"
	postgres_store_command_handler "product/src/application/commands/store/postgres"

	mongo_product_events_handler "product/src/application/events/product/mongo"
	postgres_product_events_handler "product/src/application/events/product/postgres"
	mongo_store_events_handler "product/src/application/events/store/mongo"
	postgres_store_events_handler "product/src/application/events/store/postgres"

	seedProduct "product/src/seed"

	common_consul "github.com/JohnSalazar/microservices-go-common/consul"
	consul "github.com/hashicorp/consul/api"
)

type Main struct {
	config              *config.Config
	client              *mongo.Client
	natsConn            *nats.Conn
	securityKeyService  common_services.SecurityKeysService
	managerCertificates common_security.ManagerCertificates
	adminMongoDbService *common_services.AdminMongoDbService
	verifyStoreTask     tasks.VerifyStoreTask
	postgresDatabase    *sql.DB
	redisDatabase       *redis.Client
	productReloadCache  *tasks.ProductReloadCacheTask
	httpServer          httputil.HttpServer
	consulClient        *consul.Client
	serviceID           string
}

func NewMain(
	config *config.Config,
	client *mongo.Client,
	natsConn *nats.Conn,
	securityKeyService common_services.SecurityKeysService,
	managerCertificates common_security.ManagerCertificates,
	adminMongoDbService *common_services.AdminMongoDbService,
	verifyStoreTask tasks.VerifyStoreTask,
	postgresDatabase *sql.DB,
	redisDatabase *redis.Client,
	productReloadCache *tasks.ProductReloadCacheTask,
	httpServer httputil.HttpServer,
	consulClient *consul.Client,
	serviceID string,
) *Main {
	return &Main{
		config:              config,
		client:              client,
		natsConn:            natsConn,
		securityKeyService:  securityKeyService,
		managerCertificates: managerCertificates,
		adminMongoDbService: adminMongoDbService,
		verifyStoreTask:     verifyStoreTask,
		postgresDatabase:    postgresDatabase,
		redisDatabase:       redisDatabase,
		productReloadCache:  productReloadCache,
		httpServer:          httpServer,
		consulClient:        consulClient,
		serviceID:           serviceID,
	}
}

var production *bool
var disableTrace *bool
var runMigrations *bool
var disableProductReloadCache *bool
var seed *bool

func main() {
	production = flag.Bool("prod", false, "use -prod=true to run in production mode")
	disableTrace = flag.Bool("disable-trace", false, "use disable-trace=true if you want to disable tracing completly")
	runMigrations = flag.Bool("migrations", false, "use migrations=true if you want to run migrations")
	disableProductReloadCache = flag.Bool("disable-product-reload-cache", false, "use disable-product-reload-cache=true if you want to disable product reload cache")
	seed = flag.Bool("seed", false, "use seed=true if you want to enable product recharge")

	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	app, err := startup(ctx)
	if err != nil {
		panic(err)
	}

	err = app.client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer app.client.Disconnect(ctx)

	err = app.client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	defer app.natsConn.Close()

	defer app.postgresDatabase.Close()
	err = app.postgresDatabase.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to Postgres")

	err = app.redisDatabase.Ping(ctx).Err()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to Redis")

	providerTracer, err := provider.NewProvider(provider.ProviderConfig{
		JaegerEndpoint: app.config.Jaeger.JaegerEndpoint,
		ServiceName:    app.config.Jaeger.ServiceName,
		ServiceVersion: app.config.Jaeger.ServiceVersion,
		Production:     *production,
		Disabled:       *disableTrace,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer providerTracer.Close(ctx)
	log.Println("Connected to Jaegger")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	userMongoExporter, err := app.adminMongoDbService.VerifyMongoDBExporterUser()
	if err != nil {
		log.Fatal(err)
	}

	if !userMongoExporter {
		log.Fatal("MongoDB Exporter user not found!")
	}

	app.verifyStoreTask.Run()

	if !*disableProductReloadCache {
		app.productReloadCache.Run()
	}

	app.httpServer.RunTLSServer()

	<-done
	err = app.consulClient.Agent().ServiceDeregister(app.serviceID)
	if err != nil {
		log.Printf("consul deregister error: %s", err)
	}

	log.Print("Server Stopped")
	os.Exit(0)
}

func startup(ctx context.Context) (*Main, error) {
	logger := common_log.NewLogger()
	config := config.LoadConfig(*production, "./config/")
	helpers.CreateFolder(config.Folders)
	common_validator.NewValidator("en")

	consulClient, serviceID, err := common_consul.NewConsulClient(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	checkServiceName := common_tasks.NewCheckServiceNameTask()

	certificateServiceNameDone := make(chan bool)
	go checkServiceName.ReloadServiceName(
		ctx,
		config,
		consulClient,
		config.Certificates.ServiceName,
		common_consul.CertificatesAndSecurityKeys,
		certificateServiceNameDone)
	<-certificateServiceNameDone

	emailsServiceNameDone := make(chan bool)
	go checkServiceName.ReloadServiceName(
		ctx,
		config,
		consulClient,
		config.EmailService.ServiceName,
		common_consul.EmailService,
		emailsServiceNameDone)
	<-emailsServiceNameDone

	metricService, err := common_services.NewMetricsService(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	client, err := mongo_repository.NewMongoClient(config)
	if err != nil {
		return nil, err
	}

	certificatesService := common_services.NewCertificatesService(config)
	managerCertificates := common_security.NewManagerCertificates(config, certificatesService)
	emailService := common_grpc_client.NewEmailServiceClientGrpc(config, certificatesService)

	checkCertificates := common_tasks.NewCheckCertificatesTask(config, managerCertificates, emailService)
	certsDone := make(chan bool)
	go checkCertificates.Start(ctx, certsDone)
	<-certsDone

	nc, err := common_nats.NewNats(config, certificatesService)
	if err != nil {
		log.Fatalf("Nats connect error: %+v", err)
	}
	log.Printf("Nats Connected Status: %+v	", nc.Status().String())

	productSubjects := subjects.GetProductSubjects()
	_, err = common_nats.NewJetStream(nc, "product", productSubjects)
	if err != nil {
		log.Fatalf("Nats JetStream create error: %+v", err)
	}

	storeSubjects := subjects.GetStoreSubjects()
	storeSubjects = append(storeSubjects, string(common_nats.StoreBook))
	storeSubjects = append(storeSubjects, string(common_nats.StorePayment))
	js, err := common_nats.NewJetStream(nc, "store", storeSubjects)
	if err != nil {
		log.Fatalf("Nats JetStream create error: %+v", err)
	}

	natsPublisher := common_nats.NewPublisher(js)

	mongoDatabase := mongo_repository.NewMongoDatabase(config.MongoDB.Database, client)
	adminMongoDbRepository := common_repositories.NewAdminMongoDbRepository(mongoDatabase)
	adminMongoDbService := common_services.NewAdminMongoDbService(config, adminMongoDbRepository)

	productMongoRepository := mongo_repository.NewProductRepository(mongoDatabase)
	storeMongoRepository := mongo_repository.NewStoreRepository(mongoDatabase)

	mongoEventSourcingDatabase := mongo_repository.NewMongoDatabase("event-sourcing", client)
	eventSourcingMongoRepository := mongo_repository.NewEventSourcingRepository(mongoEventSourcingDatabase)

	postgresDatabase, err := postgres_repository.NewPostgresDatabase(config)
	if err != nil {
		log.Fatal(err)
	}

	productPostgresRepository := postgres_repository.NewProductRepository(postgresDatabase)
	storePostgresRepository := postgres_repository.NewStoreRepository(postgresDatabase)

	redisDatabase := redis_repository.NewRedisClient(config)
	productRedisRepository := redis_repository.NewProductRepository(redisDatabase)
	productRepositoryDecorator := decorators.NewProductRepositoryDecorator(productMongoRepository, productPostgresRepository, productRedisRepository, natsPublisher)

	storeTask := tasks.NewStoreTask(storePostgresRepository, emailService, natsPublisher)

	postgresProductEventsHandler := postgres_product_events_handler.NewProductEventHandler(natsPublisher)
	mongoProductEventsHandler := mongo_product_events_handler.NewProductEventHandler(productRedisRepository, natsPublisher)

	postgresStoreEventsHandler := postgres_store_events_handler.NewStoreEventHandler(storeTask, natsPublisher)
	mongoStoreEventsHandler := mongo_store_events_handler.NewStoreEventHandler()

	postgresProductCommandHandler := postgres_product_command_handler.NewProductCommandHandler(productPostgresRepository, eventSourcingMongoRepository, postgresProductEventsHandler)
	mongoProductCommandHandler := mongo_product_command_handler.NewProductCommandHandler(productMongoRepository, mongoProductEventsHandler)

	postgresStoreCommandHandler := postgres_store_command_handler.NewStoreCommandHandler(storePostgresRepository, eventSourcingMongoRepository, postgresStoreEventsHandler, natsPublisher)
	mongoStoreCommandHandler := mongo_store_command_handler.NewStoreCommandHandler(storeMongoRepository, mongoStoreEventsHandler)

	securityKeysService := common_services.NewSecurityKeysService(config, certificatesService)
	managerSecurityKeys := common_security.NewManagerSecurityKeys(config, securityKeysService)
	managerTokens := common_security.NewManagerTokens(config, managerSecurityKeys)

	listens := product_nats.NewListen(
		config,
		js,
		postgresProductCommandHandler,
		mongoProductCommandHandler,
		postgresStoreCommandHandler,
		mongoStoreCommandHandler,
		emailService)

	authentication := middlewares.NewAuthentication(logger, managerTokens)

	productController := controllers.NewProductController(
		productRepositoryDecorator,
		productMongoRepository,
		productPostgresRepository,
		productRedisRepository,
		postgresProductCommandHandler,
		postgresStoreCommandHandler,
		natsPublisher,
	)
	router := routers.NewRouter(config, metricService, authentication, productController)
	productReloadCache := tasks.NewProductReloadCacheTask(productMongoRepository, productRedisRepository, emailService)
	httpServer := httputil.NewHttpServer(config, router.RouterSetup(), certificatesService)
	app := NewMain(
		config,
		client,
		nc,
		securityKeysService,
		managerCertificates,
		adminMongoDbService,
		storeTask,
		postgresDatabase,
		redisDatabase,
		productReloadCache,
		httpServer,
		consulClient,
		serviceID,
	)

	if *runMigrations {
		migrate.Run(config)
	}

	if *seed {
		seedProduct.Run(postgresProductCommandHandler)
	}

	listens.Listen()

	return app, nil
}
