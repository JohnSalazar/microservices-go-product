package routers

import (
	"fmt"
	"product/src/controllers"

	"github.com/JohnSalazar/microservices-go-common/config"
	"github.com/JohnSalazar/microservices-go-common/middlewares"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	common_service "github.com/JohnSalazar/microservices-go-common/services"
)

type Router struct {
	config            *config.Config
	serviceMetrics    common_service.Metrics
	authentication    *middlewares.Authentication
	productController *controllers.ProductController
}

func NewRouter(
	config *config.Config,
	serviceMetrics common_service.Metrics,
	authentication *middlewares.Authentication,
	productController *controllers.ProductController,
) *Router {
	return &Router{
		config:            config,
		serviceMetrics:    serviceMetrics,
		authentication:    authentication,
		productController: productController,
	}
}

func (r *Router) RouterSetup() *gin.Engine {
	router := r.initRouter()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middlewares.CORS())
	router.Use(location.Default())
	router.Use(otelgin.Middleware(r.config.Jaeger.ServiceName))
	router.Use(middlewares.Metrics(r.serviceMetrics))

	router.GET("/healthy", middlewares.Healthy())
	router.GET("/metrics", middlewares.MetricsHandler())

	v1 := router.Group(fmt.Sprintf("/api/%s", r.config.ApiVersion))

	v1.GET("/:name/:page/:size", r.productController.GetAll)
	v1.GET("/id/:id", r.productController.GetProductById)
	v1.GET("/slug/:slug", r.productController.GetProductBySlug)
	v1.GET("/refresh", r.authentication.Verify(),
		middlewares.Authorization("admin", "update"),
		r.productController.Refresh)
	v1.POST("/", r.authentication.Verify(),
		middlewares.Authorization("product", "create"),
		r.productController.AddProduct)
	v1.POST("/book", r.authentication.Verify(), r.productController.Book)
	v1.PUT("/:id", r.authentication.Verify(),
		middlewares.Authorization("product", "update"),
		r.productController.UpdateProduct)
	v1.PUT("/payment", r.authentication.Verify(), r.productController.Payment)

	return router
}

func (r *Router) initRouter() *gin.Engine {
	if r.config.Production {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	return gin.New()
}
