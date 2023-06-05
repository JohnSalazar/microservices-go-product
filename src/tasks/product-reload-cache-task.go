package tasks

import (
	"context"
	"fmt"
	"log"
	product_repository "product/src/data/repositories/interfaces"
	redis_product_repository "product/src/data/repositories/redis"
	"sync"
	"time"

	common_helpers "github.com/oceano-dev/microservices-go-common/helpers"
	common_service "github.com/oceano-dev/microservices-go-common/services"

	trace "github.com/oceano-dev/microservices-go-common/trace/otel"
)

type ProductReloadCacheTask struct {
	mongoRepository product_repository.ProductRepository
	redisRepository redis_product_repository.ProductRepository
	email           common_service.EmailService
}

var (
	timeToReload = "00:03:00"
	loadingCache bool
	mLoading     sync.Mutex
)

func NewProductReloadCacheTask(
	mongoRepository product_repository.ProductRepository,
	redisRepository redis_product_repository.ProductRepository,
	email common_service.EmailService,
) *ProductReloadCacheTask {
	return &ProductReloadCacheTask{
		mongoRepository: mongoRepository,
		redisRepository: redisRepository,
		email:           email,
	}
}

func (task *ProductReloadCacheTask) Run() {
	ticker := time.NewTicker(2 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if loadingCache {
					ticker.Reset(15 * time.Second)
					break
				}

				mLoading.Lock()
				loadingCache = true
				mLoading.Unlock()

				ctx := context.Background()
				products, err := task.mongoRepository.GetAll(ctx, "", 0, 0)
				if err != nil {
					_, span := trace.NewSpan(ctx, "tasks.ProductReloadTask")
					defer span.End()
					msg := fmt.Sprintf("error task product reload mongo getall: %s", err.Error())
					trace.FailSpan(span, msg)
					log.Print(msg)
					go task.email.SendSupportMessage(msg)
					ticker.Reset(15 * time.Second)
					break
				}

				err = task.redisRepository.Refresh(ctx, products)
				if err != nil {
					_, span := trace.NewSpan(ctx, "tasks.ProductReloadTask")
					defer span.End()
					msg := fmt.Sprintf("error task product redis refresh : %s", err.Error())
					trace.FailSpan(span, msg)
					log.Print(msg)
					go task.email.SendSupportMessage(msg)
					ticker.Reset(15 * time.Second)
					break
				}

				nextTime, err := common_helpers.NextTime(timeToReload)
				if err != nil {
					_, span := trace.NewSpan(ctx, "tasks.ProductReloadTask")
					defer span.End()
					msg := fmt.Sprintf("error next time: %s", err.Error())
					trace.FailSpan(span, msg)
					log.Print(msg)
					go task.email.SendSupportMessage(msg)
					ticker.Stop()
					break
				}

				fmt.Printf("product redis refresh successfully: %s\n", time.Now().UTC())
				fmt.Printf("next refresh: %s\n", nextTime)

				rest := time.Until(nextTime).Seconds()

				mLoading.Lock()
				loadingCache = false
				mLoading.Unlock()

				ticker.Reset(time.Duration(rest) * time.Second)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	mLoading.Lock()
	loadingCache = false
	mLoading.Unlock()
}
