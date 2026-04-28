package main

import (
	"context"
	"net/http"

	"be-modami-core-service/config"
	_ "be-modami-core-service/docs" // swagger generated
	"be-modami-core-service/internal/adapter/consumer"
	"be-modami-core-service/internal/adapter/handler"
	hmw "be-modami-core-service/internal/adapter/handler/middleware"
	"be-modami-core-service/internal/adapter/producer"
	"be-modami-core-service/internal/adapter/repository"
	"be-modami-core-service/internal/port"
	"be-modami-core-service/internal/service"
	localkafka "be-modami-core-service/pkg/kafka"
	"be-modami-core-service/pkg/mongodb"

	pkgkafka "gitlab.com/lifegoeson-libs/pkg-gokit/kafka"
	pkgredis "gitlab.com/lifegoeson-libs/pkg-gokit/redis"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
	loggingmw "gitlab.com/lifegoeson-libs/pkg-logging/middleware"
	mongodri "go.mongodb.org/mongo-driver/v2/mongo"
)

// Application holds the running HTTP server.
type Application struct {
	HTTPServer *http.Server
}

// appRepos groups all MongoDB repository instances.
type appRepos struct {
	product         port.ProductRepository
	category        port.CategoryRepository
	hashtag         port.HashtagRepository
	favorite        port.FavoriteRepository
	savedProduct    port.SavedProductRepository
	savedCollection port.SavedCollectionRepository
	follow          port.FollowRepository
	review          port.ReviewRepository
	blog            port.BlogRepository
}

// appServices groups all domain service instances.
type appServices struct {
	product    *service.ProductService
	masterdata *service.MasterdataService
	seller     *service.SellerService
	blog       *service.BlogService
	homeFeed   *service.HomeFeedService
}

func newApplication(ctx context.Context, cfg *config.Config, conns *Connections) (*Application, error) {
	mongodb.EnsureIndexes(ctx, conns.DB)

	repos := initRepositories(conns.DB)
	prod := newKafkaProducer(cfg)
	svcs := initServices(repos, conns.Redis, prod)

	routerHandler := buildRouter(cfg, svcs)
	startSyncConsumer(ctx, cfg, conns, repos.product)

	srv := newHTTPServer(cfg, routerHandler)
	logger.Info(ctx, "application routes registered", logging.String("addr", cfg.App.ListenAddr()))

	return &Application{HTTPServer: srv}, nil
}

// initRepositories constructs all MongoDB repository adapters.
func initRepositories(db *mongodri.Database) *appRepos {
	return &appRepos{
		product:         repository.NewProductRepository(db),
		category:        repository.NewCategoryRepository(db),
		hashtag:         repository.NewHashtagRepository(db),
		favorite:        repository.NewFavoriteRepository(db),
		savedProduct:    repository.NewSavedProductRepository(db),
		savedCollection: repository.NewSavedCollectionRepository(db),
		follow:          repository.NewFollowRepository(db),
		review:          repository.NewReviewRepository(db),
		blog:            repository.NewBlogRepository(db),
	}
}

// initServices wires domain services from repositories, cache, and producer.
func initServices(repos *appRepos, cache pkgredis.CachePort, prod port.ProductProducer) *appServices {
	return &appServices{
		product:    service.NewProductService(repos.product, repos.category, cache, prod),
		masterdata: service.NewMasterdataService(repos.category, repos.hashtag),
		seller:     service.NewSellerService(repos.product, repos.favorite, repos.follow, repos.review),
		blog:       service.NewBlogService(repos.blog),
		homeFeed:   service.NewHomeFeedService(repos.product, repos.category, repos.blog),
	}
}

// newKafkaProducer creates a Kafka producer if Kafka is enabled; returns nil otherwise.
func newKafkaProducer(cfg *config.Config) port.ProductProducer {
	if len(cfg.Kafka.Brokers()) == 0 {
		return nil
	}
	kafkaCfg := &pkgkafka.Config{
		Brokers:          cfg.Kafka.Brokers(),
		ClientID:         cfg.Kafka.ClientID + "-producer",
		ProducerOnlyMode: true,
	}
	var opts []pkgkafka.ServiceOption
	if cfg.Kafka.Env != "" {
		resolver := localkafka.NewEnvTopicResolver(cfg.Kafka.Env,
			localkafka.TopicProductCreated,
			localkafka.TopicProductUpdated,
			localkafka.TopicProductDeleted,
		)
		opts = append(opts, pkgkafka.WithTopicResolver(resolver))
	}
	ks, err := pkgkafka.NewKafkaService(kafkaCfg, opts...)
	if err != nil {
		return nil
	}
	return producer.NewProductProducer(ks)
}

// buildRouter constructs the Gin engine with all middleware and routes registered.
func buildRouter(cfg *config.Config, svcs *appServices) http.Handler {
	if !cfg.App.Debug && cfg.Observability.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	productH := handler.NewProductHandler(svcs.product)
	masterdataH := handler.NewMasterdataHandler(svcs.masterdata)
	sellerH := handler.NewSellerHandler(svcs.seller)
	searchH := handler.NewSearchHandler(productH, masterdataH)
	blogH := handler.NewBlogHandler(svcs.blog)
	homeFeedH := handler.NewHomeFeedHandler(svcs.homeFeed)

	auth := hmw.NewAuth(cfg.Keycloak.JWKSUrl)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.App.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: cfg.App.AllowCredentials,
		MaxAge:           300,
	}))

	r.GET("/health", handler.Health)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	registerRoutes(r.Group("/v1/core-services"), auth, productH, masterdataH, sellerH, searchH, blogH, homeFeedH)
	return loggingmw.HTTPMiddleware(cfg.Observability.ServiceName, r, &loggingmw.HttpLoggingOptions{
		ExceptRoutes: []string{"/health", "/swagger", "/swagger/index.html"},
	})
}

// registerRoutes attaches all v1 routes to the provided router group.
func registerRoutes(
	v1 *gin.RouterGroup,
	auth *hmw.Auth,
	productH *handler.ProductHandler,
	masterdataH *handler.MasterdataHandler,
	sellerH *handler.SellerHandler,
	searchH *handler.SearchHandler,
	blogH *handler.BlogHandler,
	homeFeedH *handler.HomeFeedHandler,
) {
	v1.GET("/home-feeds", homeFeedH.GetHomeFeed)

	// Products — public
	v1.GET("/products/feed", productH.Feed)
	v1.GET("/products/featured", productH.Featured)
	v1.GET("/products/select", productH.SelectProducts)
	v1.GET("/products/search", productH.Search)
	v1.GET("/products/slug/:slug", productH.GetBySlug)
	v1.GET("/products/:id", productH.GetByID)
	v1.GET("/products/:id/similar", productH.Similar)
	v1.GET("/products/:id/moderation", productH.GetModeration)
	v1.POST("/products/:id/view", productH.TrackView)

	// Products — authenticated
	authProducts := v1.Group("")
	authProducts.Use(auth.Required())
	authProducts.GET("/products/me", productH.MyProducts)
	authProducts.POST("/products", productH.Create)
	authProducts.PUT("/products/:id", productH.Update)
	authProducts.DELETE("/products/:id", productH.Delete)
	authProducts.POST("/products/:id/submit", productH.Submit)
	authProducts.POST("/products/:id/resubmit", productH.Resubmit)
	authProducts.POST("/products/:id/archive", productH.Archive)
	authProducts.POST("/products/:id/unarchive", productH.Unarchive)

	// Search
	v1.GET("/search", searchH.Search)
	v1.GET("/search/suggest", searchH.Suggest)
	v1.GET("/search/trending", searchH.Trending)
	v1.GET("/hashtags/:tag/products", productH.HashtagProducts)

	// Categories — public
	v1.GET("/categories", masterdataH.ListCategories)
	v1.GET("/categories/:slug", masterdataH.GetCategory)
	v1.GET("/categories/:slug/children", masterdataH.GetCategoryChildren)

	// Hashtags
	v1.GET("/hashtags/trending", masterdataH.TrendingHashtags)
	v1.GET("/hashtags/suggest", masterdataH.SuggestHashtags)

	// Sellers
	v1.GET("/sellers/:id", sellerH.GetProfile)
	v1.GET("/sellers/:id/products", sellerH.GetProducts)
	v1.GET("/sellers/:id/reviews", sellerH.GetReviews)
	v1.GET("/sellers/:id/stats", sellerH.GetStats)

	// Community & Blog — public
	v1.GET("/community", blogH.CommunityFeed)
	v1.GET("/blog/posts", blogH.ListPosts)
	v1.GET("/blog/posts/:slug", blogH.GetPost)
	v1.GET("/blog/reports", blogH.ListTrendReports)
	v1.GET("/blog/hashtags/:tag", blogH.HashtagPosts)

	// Categories & Blog — protected
	protected := v1.Group("")
	protected.Use(auth.Required())
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

// startSyncConsumer launches the Kafka→ES product sync consumer in a goroutine if conditions are met.
func startSyncConsumer(ctx context.Context, cfg *config.Config, conns *Connections, productRepo port.ProductRepository) {
	if len(cfg.Kafka.Brokers()) == 0 || conns.Elasticsearch == nil {
		return
	}
	kafkaCfg := &pkgkafka.Config{
		Brokers:         cfg.Kafka.Brokers(),
		ClientID:        cfg.Kafka.ClientID,
		ConsumerGroupID: cfg.Kafka.ConsumerGroup,
	}
	var opts []pkgkafka.ServiceOption
	if cfg.Kafka.Env != "" {
		resolver := localkafka.NewEnvTopicResolver(cfg.Kafka.Env,
			localkafka.TopicProductCreated,
			localkafka.TopicProductUpdated,
			localkafka.TopicProductDeleted,
		)
		opts = append(opts, pkgkafka.WithTopicResolver(resolver))
	}
	ks, err := pkgkafka.NewKafkaService(kafkaCfg, opts...)
	if err != nil {
		return
	}
	syncConsumer := consumer.NewSyncProductConsumer(conns.Elasticsearch, productRepo)
	go func() {
		if err := ks.StartConsumer(ctx, []pkgkafka.ConsumerHandler{syncConsumer}); err != nil {
			logger.Error(ctx, "sync consumer stopped", err)
		}
	}()
}

// newHTTPServer creates the http.Server with timeouts from config.
func newHTTPServer(cfg *config.Config, h http.Handler) *http.Server {
	return &http.Server{
		Addr:         cfg.App.ListenAddr(),
		Handler:      h,
		ReadTimeout:  cfg.App.GetReadTimeout(),
		WriteTimeout: cfg.App.GetWriteTimeout(),
		IdleTimeout:  cfg.App.GetIdleTimeout(),
	}
}
