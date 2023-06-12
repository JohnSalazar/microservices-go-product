package controllers

import (
	"context"
	"net/http"
	command_product "product/src/application/commands/product"
	postgres_product_command_handler "product/src/application/commands/product/postgres"
	command_store "product/src/application/commands/store"
	postgres_store_command_handler "product/src/application/commands/store/postgres"
	repository_interface "product/src/data/repositories/interfaces"
	redis_repository_interface "product/src/data/repositories/redis"
	"product/src/decorators"
	"product/src/dtos"
	"strconv"

	"strings"

	"github.com/JohnSalazar/microservices-go-common/httputil"
	common_nats "github.com/JohnSalazar/microservices-go-common/nats"
	trace "github.com/JohnSalazar/microservices-go-common/trace/otel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductController struct {
	productRepositoryDecorator    decorators.ProductRepositoryDecorator
	productMongoRepository        repository_interface.ProductRepository
	productPostgresRepository     repository_interface.ProductRepository
	productRedisRepository        redis_repository_interface.ProductRepository
	productPostgresCommandHandler *postgres_product_command_handler.ProductCommandHandler
	storePostgresCommandHandler   *postgres_store_command_handler.StoreCommandHandler
	publisher                     common_nats.Publisher
}

func NewProductController(
	productRepositoryDecorator decorators.ProductRepositoryDecorator,
	productMongoRepository repository_interface.ProductRepository,
	productPostgresRepository repository_interface.ProductRepository,
	productRedisRepository redis_repository_interface.ProductRepository,
	productPostgresCommandHandler *postgres_product_command_handler.ProductCommandHandler,
	storePostgresCommandHandler *postgres_store_command_handler.StoreCommandHandler,
	publisher common_nats.Publisher,
) *ProductController {
	return &ProductController{
		productRepositoryDecorator:    productRepositoryDecorator,
		productMongoRepository:        productMongoRepository,
		productPostgresRepository:     productPostgresRepository,
		productRedisRepository:        productRedisRepository,
		productPostgresCommandHandler: productPostgresCommandHandler,
		storePostgresCommandHandler:   storePostgresCommandHandler,
		publisher:                     publisher,
	}
}

func (product *ProductController) GetAll(c *gin.Context) {
	_, span := trace.NewSpan(c.Request.Context(), "ProductController.GetAll")
	defer span.End()

	name := c.Param("name")

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, "page is required")
		return
	}

	size, err := strconv.Atoi(c.Param("size"))
	if err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, "size is required")
		return
	}

	//products, err := product.productMongoRepository.GetAll(c.Request.Context(), page, size)
	products, err := product.productRepositoryDecorator.GetAll(c.Request.Context(), name, page, size)
	if err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, "products get error")
		return
	}

	c.JSON(http.StatusOK, products)
}

func (product *ProductController) GetProductById(c *gin.Context) {
	_, span := trace.NewSpan(c.Request.Context(), "ProductController.GetProductById")
	defer span.End()

	ID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, "invalid id")
		return
	}

	//_product, err := product.productMongoRepository.FindByID(c.Request.Context(), ID)
	_product, err := product.productRepositoryDecorator.FindByID(c.Request.Context(), ID)
	if _product == nil || err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, "products not found")
		return
	}

	c.JSON(http.StatusOK, _product)
}

func (product *ProductController) GetProductBySlug(c *gin.Context) {
	_, span := trace.NewSpan(c.Request.Context(), "ProductController.GetProductBySlug")
	defer span.End()

	slug := c.Param("slug")
	if strings.Trim(slug, "") == "" {
		httputil.NewResponseError(c, http.StatusBadRequest, "invalid slug")
		return
	}

	//_product, err := product.productMongoRepository.FindByID(c.Request.Context(), ID)
	_product, err := product.productRepositoryDecorator.FindBySlug(c.Request.Context(), slug)
	if _product == nil || err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, "products not found")
		return
	}

	c.JSON(http.StatusOK, _product)
}

func (product *ProductController) AddProduct(c *gin.Context) {
	ctx, span := trace.NewSpan(c.Request.Context(), "ProductController.AddProduct")
	defer span.End()

	createProductPostgresCommand := &command_product.CreateProductCommand{}
	err := c.BindJSON(createProductPostgresCommand)
	if err != nil {
		trace.FailSpan(span, "Error json parse")
		httputil.NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	if createProductPostgresCommand.ID == uuid.Nil {
		createProductPostgresCommand.ID = uuid.New()
	}

	productModel, err := product.productPostgresCommandHandler.CreateProductCommandHandler(ctx, createProductPostgresCommand)
	if err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusCreated, productModel)
}

func (product *ProductController) UpdateProduct(c *gin.Context) {
	ctx, span := trace.NewSpan(c.Request.Context(), "ProductController.UpdateProduct")
	defer span.End()

	ID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, "invalid product id")
		return
	}

	updateProductPostgresCommand := &command_product.UpdateProductCommand{}
	err = c.BindJSON(updateProductPostgresCommand)
	if err != nil {
		trace.FailSpan(span, "Error json parse")
		httputil.NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	if updateProductPostgresCommand.ID != ID {
		trace.FailSpan(span, "Error divergent product id")
		httputil.NewResponseError(c, http.StatusBadRequest, "Error divergent product id")
		return
	}

	productModel, err := product.productPostgresCommandHandler.UpdateProductCommandHandler(ctx, updateProductPostgresCommand)
	if err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, productModel)
}

func (product *ProductController) Book(c *gin.Context) {
	ctx, span := trace.NewSpan(c.Request.Context(), "ProductController.Book")
	defer span.End()

	bookStoreCommand := &command_store.BookStoreCommand{}
	err := c.BindJSON(bookStoreCommand)
	if err != nil {
		trace.FailSpan(span, "Error json parse")
		httputil.NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	err = product.storePostgresCommandHandler.BookStoreCommandHandler(ctx, bookStoreCommand)
	if err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	// dataPostgres, err := json.Marshal(bookStorePostgresCommand)
	// if err != nil {
	// 	trace.FailSpan(span, "error json parse")
	// 	httputil.NewResponseError(c, http.StatusBadRequest, err.Error())
	// 	return
	// }

	// err = product.publisher.Publish(string(common_nats.StoreBook), dataPostgres)
	// if err != nil {
	// 	httputil.NewResponseError(c, http.StatusBadRequest, fmt.Sprintf("%s error: %s", common_nats.StoreBook, err.Error()))
	// 	return
	// }

	bookStoreDTO := &dtos.BookStore{
		Products: bookStoreCommand.Products,
	}

	c.JSON(http.StatusOK, bookStoreDTO)
}

func (product *ProductController) Payment(c *gin.Context) {
	ctx, span := trace.NewSpan(c.Request.Context(), "ProductController.Payment")
	defer span.End()

	paymentStoreCommand := &command_store.PaymentStoreCommand{}
	err := c.BindJSON(paymentStoreCommand)
	if err != nil {
		trace.FailSpan(span, "Error json parse")
		httputil.NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	stores, err := product.storePostgresCommandHandler.PaymentStoreCommandHandler(ctx, paymentStoreCommand)
	if err != nil {
		httputil.NewResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	// dataPostgres, err := json.Marshal(paymentStorePostgresCommand)
	// if err != nil {
	// 	trace.FailSpan(span, "error json parse")
	// 	httputil.NewResponseError(c, http.StatusBadRequest, err.Error())
	// 	return
	// }

	// err = product.publisher.Publish(string(common_nats.StorePaymentPostgres), dataPostgres)
	// if err != nil {
	// 	httputil.NewResponseError(c, http.StatusBadRequest, fmt.Sprintf("%s error: %s", common_nats.StorePaymentPostgres, err.Error()))
	// 	return
	// }

	paymentsStoreDTO := []*dtos.PaymentStore{}
	for _, store := range stores {
		paymentStoreDTO := &dtos.PaymentStore{
			ID:   store.ID,
			Sold: store.Sold,
		}

		paymentsStoreDTO = append(paymentsStoreDTO, paymentStoreDTO)

	}

	c.JSON(http.StatusOK, paymentsStoreDTO)
}

func (product *ProductController) Refresh(c *gin.Context) {
	ctx := context.Background()
	go func(ctx context.Context) {
		_, span := trace.NewSpan(c.Request.Context(), "ProductController.Refresh")
		defer span.End()

		products, err := product.productMongoRepository.GetAll(ctx, "", 0, 0)
		if err != nil {
			httputil.NewResponseError(c, http.StatusBadRequest, "products get error")
			return
		}

		err = product.productRedisRepository.Refresh(ctx, products)
		if err != nil {
			httputil.NewResponseError(c, http.StatusBadRequest, "products refresh error")
			return
		}
	}(ctx)

	c.JSON(http.StatusOK, "refresh requested")
}
