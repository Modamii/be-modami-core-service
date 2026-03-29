package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/modami/core-service/config"
	_ "github.com/modami/core-service/docs" // swagger generated
	"github.com/modami/core-service/internal/adapter/handler"
	hmw "github.com/modami/core-service/internal/adapter/handler/middleware"
	"github.com/modami/core-service/internal/adapter/repository"
	"github.com/modami/core-service/internal/service"
	"github.com/modami/core-service/pkg/storage/database/mongodb"
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
	_ = ctx
	db := conns.DB

	mongodb.EnsureIndexes(ctx, db)

	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	packageRepo := repository.NewPackageRepository(db)
	hashtagRepo := repository.NewHashtagRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	savedProductRepo := repository.NewSavedProductRepository(db)
	savedCollectionRepo := repository.NewSavedCollectionRepository(db)
	followRepo := repository.NewFollowRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	productSvc := service.NewProductService(productRepo)
	orderProductReader := service.NewOrderProductReader(productSvc)
	orderSvc := service.NewOrderService(orderRepo, orderProductReader)
	masterdataSvc := service.NewMasterdataService(categoryRepo, packageRepo, hashtagRepo)
	sellerSvc := service.NewSellerService(productRepo, favoriteRepo, followRepo, reviewRepo)

	productH := handler.NewProductHandler(productSvc)
	orderH := handler.NewOrderHandler(orderSvc)
	masterdataH := handler.NewMasterdataHandler(masterdataSvc)
	sellerH := handler.NewSellerHandler(sellerSvc)
	searchH := handler.NewSearchHandler(productH, masterdataH)

	_ = favoriteRepo
	_ = savedProductRepo
	_ = savedCollectionRepo
	_ = reviewRepo

	if !cfg.App.Debug && cfg.Observability.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	auth := hmw.NewAuth(cfg.Security.JWTSecret)

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

	v1 := router.Group("/api/v1")

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
	v1.GET("/sellers/:seller_id/products", productH.SellerProducts)

	v1.GET("/search", searchH.Search)
	v1.GET("/search/suggest", searchH.Suggest)
	v1.GET("/search/trending", searchH.Trending)
	v1.GET("/hashtags/:tag/products", productH.HashtagProducts)

	v1.GET("/categories", masterdataH.ListCategories)
	v1.GET("/categories/:slug", masterdataH.GetCategory)
	v1.GET("/categories/:slug/children", masterdataH.GetCategoryChildren)

	v1.GET("/packages", masterdataH.ListPackages)
	v1.GET("/packages/:code", masterdataH.GetPackage)

	v1.GET("/hashtags/trending", masterdataH.TrendingHashtags)
	v1.GET("/hashtags/suggest", masterdataH.SuggestHashtags)

	v1.GET("/sellers/:id", sellerH.GetProfile)
	v1.GET("/sellers/:id/products", sellerH.GetProducts)
	v1.GET("/sellers/:id/reviews", sellerH.GetReviews)
	v1.GET("/sellers/:id/stats", sellerH.GetStats)

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

		protected.POST("/orders", orderH.CreateOrder)
		protected.GET("/orders/my-purchases", orderH.ListByBuyer)
		protected.GET("/orders/my-sales", orderH.ListBySeller)
		protected.GET("/orders/:id", orderH.GetByID)
		protected.PUT("/orders/:id/confirm", orderH.Confirm)
		protected.PUT("/orders/:id/ship", orderH.Ship)
		protected.PUT("/orders/:id/receive", orderH.Receive)
		protected.PUT("/orders/:id/cancel", orderH.Cancel)
		protected.GET("/orders/:id/events", orderH.ListEvents)
	}

	admin := v1.Group("/admin")
	admin.Use(auth.Required(), hmw.AdminOnly)
	{
		admin.GET("/orders", orderH.AdminListAll)
		admin.GET("/orders/:id", orderH.AdminGetByID)
		admin.PUT("/orders/:id/force-cancel", orderH.ForceCancel)
		admin.PUT("/orders/:id/force-complete", orderH.ForceComplete)

		admin.POST("/categories", masterdataH.AdminCreateCategory)
		admin.PUT("/categories/:id", masterdataH.AdminUpdateCategory)
		admin.PUT("/categories/:id/toggle", masterdataH.AdminToggleCategory)
		admin.DELETE("/categories/:id", masterdataH.AdminDeleteCategory)
		admin.PUT("/categories/reorder", masterdataH.AdminReorderCategories)

		admin.POST("/packages", masterdataH.AdminCreatePackage)
		admin.PUT("/packages/:id", masterdataH.AdminUpdatePackage)
		admin.PUT("/packages/:id/toggle", masterdataH.AdminTogglePackage)
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

	return &Application{HTTPServer: srv}, nil
}
