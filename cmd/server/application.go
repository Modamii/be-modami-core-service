package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"be-modami-core-service/config"
	_ "be-modami-core-service/docs" // swagger generated
	"be-modami-core-service/internal/adapter/consumer"
	"be-modami-core-service/internal/adapter/handler"
	hmw "be-modami-core-service/internal/adapter/handler/middleware"
	"be-modami-core-service/internal/adapter/producer"
	"be-modami-core-service/internal/adapter/repository"
	"be-modami-core-service/internal/port"
	"be-modami-core-service/internal/service"
	kafkapkg "be-modami-core-service/pkg/kafka"
	"be-modami-core-service/pkg/storage/database/mongodb"
	redisstorage "be-modami-core-service/pkg/storage/redis"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
	loggingmw "gitlab.com/lifegoeson-libs/pkg-logging/middleware"
)

type Application struct {
	HTTPServer *http.Server
}

func newApplication(ctx context.Context, cfg *config.Config, conns *Connections) (*Application, error) {
	db := conns.DB

	mongodb.EnsureIndexes(ctx, db)

	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	hashtagRepo := repository.NewHashtagRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	savedProductRepo := repository.NewSavedProductRepository(db)
	savedCollectionRepo := repository.NewSavedCollectionRepository(db)
	followRepo := repository.NewFollowRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	blogRepo := repository.NewBlogRepository(db)

	var redisCache redisstorage.RedisCacheService
	if conns.Redis != nil {
		redisCache = redisstorage.NewRedisCacheService(conns.Redis)
	}

	var kafkaProducer kafkapkg.Producer
	if cfg.Kafka.Enable && len(cfg.Kafka.Brokers()) > 0 {
		ks, err := kafkapkg.NewKafkaService(&kafkapkg.KafkaConfig{
			Brokers:          cfg.Kafka.Brokers(),
			ClientID:         cfg.Kafka.ClientID + "-producer",
			ProducerOnlyMode: true,
		}, cfg.Kafka.Env)
		if err == nil {
			kafkaProducer = ks
		}
	}

	var productProducer port.ProductProducer
	if kafkaProducer != nil {
		productProducer = producer.NewProductProducer(kafkaProducer)
	}
	productSvc := service.NewProductService(productRepo, categoryRepo, redisCache, productProducer)
	masterdataSvc := service.NewMasterdataService(categoryRepo, hashtagRepo)
	sellerSvc := service.NewSellerService(productRepo, favoriteRepo, followRepo, reviewRepo)
	blogSvc := service.NewBlogService(blogRepo)

	productH := handler.NewProductHandler(productSvc)
	masterdataH := handler.NewMasterdataHandler(masterdataSvc)
	sellerH := handler.NewSellerHandler(sellerSvc)
	searchH := handler.NewSearchHandler(productH, masterdataH)
	blogH := handler.NewBlogHandler(blogSvc)

	_ = favoriteRepo
	_ = savedProductRepo
	_ = savedCollectionRepo
	_ = reviewRepo

	if !cfg.App.Debug && cfg.Observability.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	auth := hmw.NewAuth(cfg.Keycloak.JWKSUrl)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.GET("/health", handler.Health)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/v1/core-services")

	v1.GET("/products/feed", productH.Feed)
	v1.GET("/products/featured", productH.Featured)
	v1.GET("/products/select", productH.SelectProducts)
	v1.GET("/products/search", productH.Search)
	v1.GET("/products/slug/:slug", productH.GetBySlug)

	authProducts := v1.Group("")
	authProducts.Use(auth.Required())
	authProducts.GET("/products/me", productH.MyProducts)

	v1.GET("/products/:id", productH.GetByID)
	v1.GET("/products/:id/similar", productH.Similar)
	v1.GET("/products/:id/moderation", productH.GetModeration)
	v1.POST("/products/:id/view", productH.TrackView)

	v1.GET("/search", searchH.Search)
	v1.GET("/search/suggest", searchH.Suggest)
	v1.GET("/search/trending", searchH.Trending)
	v1.GET("/hashtags/:tag/products", productH.HashtagProducts)

	v1.GET("/categories", masterdataH.ListCategories)
	v1.GET("/categories/:slug", masterdataH.GetCategory)
	v1.GET("/categories/:slug/children", masterdataH.GetCategoryChildren)

	v1.GET("/hashtags/trending", masterdataH.TrendingHashtags)
	v1.GET("/hashtags/suggest", masterdataH.SuggestHashtags)

	v1.GET("/sellers/:id", sellerH.GetProfile)
	v1.GET("/sellers/:id/products", sellerH.GetProducts)
	v1.GET("/sellers/:id/reviews", sellerH.GetReviews)
	v1.GET("/sellers/:id/stats", sellerH.GetStats)

	// Community & Blog — public routes
	v1.GET("/community", blogH.CommunityFeed)
	v1.GET("/blog/posts", blogH.ListPosts)
	v1.GET("/blog/posts/:slug", blogH.GetPost)
	v1.GET("/blog/reports", blogH.ListTrendReports)
	v1.GET("/blog/hashtags/:tag", blogH.HashtagPosts)

	protected := v1.Group("")
	protected.Use(auth.Required())
	{
		protected.POST("/products", productH.Create)
		protected.PUT("/products/:id", productH.Update)
		protected.DELETE("/products/:id", productH.Delete)
		protected.POST("/products/:id/submit", productH.Submit)
		protected.POST("/products/:id/resubmit", productH.Resubmit)
		protected.POST("/products/:id/archive", productH.Archive)
		protected.POST("/products/:id/unarchive", productH.Unarchive)

		protected.POST("/categories", hmw.RequirePermission("category.create"), masterdataH.CreateCategory)
		protected.PUT("/categories/:id", hmw.RequirePermission("category.update"), masterdataH.UpdateCategory)
		protected.PUT("/categories/:id/toggle", hmw.RequirePermission("category.manage"), masterdataH.ToggleCategory)
		protected.DELETE("/categories/:id", hmw.RequirePermission("category.delete"), masterdataH.DeleteCategory)
		protected.PUT("/categories/reorder", hmw.RequirePermission("category.manage"), masterdataH.ReorderCategories)

		protected.POST("/blog/posts", hmw.RequirePermission("blog.create"), blogH.CreatePost)
		protected.PUT("/blog/posts/:id", hmw.RequirePermission("blog.update"), blogH.UpdatePost)
		protected.DELETE("/blog/posts/:id", hmw.RequirePermission("blog.delete"), blogH.DeletePost)
		protected.POST("/blog/posts/:id/publish", hmw.RequirePermission("blog.publish"), blogH.PublishPost)
	}

	serviceName := strings.TrimSpace(cfg.Observability.ServiceName)
	if serviceName == "" {
		serviceName = strings.TrimSpace(cfg.App.Name)
	}
	if serviceName == "" {
		serviceName = "core-service"
	}
	httpHandler := loggingmw.HTTPMiddleware(serviceName, router, &loggingmw.HttpLoggingOptions{
		ExceptRoutes: []string{"/health", "/swagger", "/swagger/index.html"},
	})

	addr := cfg.App.ListenAddr()
	readTO := cfg.App.ReadTimeout
	if readTO == 0 {
		readTO = 30 * time.Second
	}
	writeTO := cfg.App.WriteTimeout
	if writeTO == 0 {
		writeTO = 30 * time.Second
	}
	idleTO := cfg.App.IdleTimeout
	if idleTO == 0 {
		idleTO = 120 * time.Second
	}
	srv := &http.Server{
		Addr:         addr,
		Handler:      httpHandler,
		ReadTimeout:  readTO,
		WriteTimeout: writeTO,
		IdleTimeout:  idleTO,
	}

	logger.Info(context.Background(), "application routes registered", logging.String("addr", addr))

	if cfg.Kafka.Enable && len(cfg.Kafka.Brokers()) > 0 && conns.Elasticsearch != nil {
		ks, err := kafkapkg.NewKafkaService(&kafkapkg.KafkaConfig{
			Brokers:         cfg.Kafka.Brokers(),
			ClientID:        cfg.Kafka.ClientID,
			ConsumerGroupID: cfg.Kafka.ConsumerGroup,
		}, cfg.Kafka.Env)
		if err == nil {
			syncConsumer := consumer.NewSyncProductConsumer(conns.Elasticsearch, productRepo)
			go func() {
				if err := ks.StartConsumer(ctx, []kafkapkg.ConsumerHandler{syncConsumer}); err != nil {
					logger.Error(ctx, "sync consumer stopped", err)
				}
			}()
		}
	}

	return &Application{HTTPServer: srv}, nil
}
