package command

import (
	"context"
	"fmt"

	config "be-modami-core-service/config"
	"be-modami-core-service/pkg/elasticsearch"
	pkges "gitlab.com/lifegoeson-libs/pkg-gokit/elasticsearch"
	pkgmongo "gitlab.com/lifegoeson-libs/pkg-gokit/mongodb"
	pkgredis "gitlab.com/lifegoeson-libs/pkg-gokit/redis"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"

	mongodri "go.mongodb.org/mongo-driver/v2/mongo"
)

var cfg *config.Config

func SetConfig(c *config.Config) {
	cfg = c
}

func GetConfig() *config.Config {
	return cfg
}

// CommandContext holds shared connections for commands.
type CommandContext struct {
	Config      *config.Config
	DB          *mongodri.Database
	disconnect  func()
	RedisClient pkgredis.CachePort
	ESClient    *pkges.Client
}

func NewCommandContext() (*CommandContext, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config not initialized")
	}

	ctx := context.Background()
	l := logger.FromContext(ctx)

	cmdCtx := &CommandContext{Config: cfg}

	// Connect MongoDB
	l.Info("Connecting to MongoDB...")
	db, disconnect, err := pkgmongo.Connect(ctx, cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	cmdCtx.DB = db
	cmdCtx.disconnect = disconnect

	// Connect Redis
	l.Info("Connecting to Redis...")
	redisClient, err := pkgredis.NewAdapter(pkgredis.Config{
		Addrs:        []string{cfg.Redis.Addr()},
		Password:     cfg.Redis.Pass,
		DB:           cfg.Redis.Database,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  cfg.Redis.GetDialTimeout(),
		ReadTimeout:  cfg.Redis.GetReadTimeout(),
		WriteTimeout: cfg.Redis.GetWriteTimeout(),
	})
	if err != nil {
		cmdCtx.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	cmdCtx.RedisClient = redisClient

	// Connect Elasticsearch
	l.Info("Connecting to Elasticsearch...")
	esClient, err := pkges.Connect(ctx, pkges.Config{
		URL:      cfg.Elasticsearch.URL,
		Username: cfg.Elasticsearch.Username,
		Password: cfg.Elasticsearch.Password,
		Index:    cfg.Elasticsearch.Index,
	})
	if err != nil {
		l.Warn("Failed to create Elasticsearch client")
	} else {
		if pingErr := elasticsearch.Ping(esClient); pingErr != nil {
			l.Warn("Elasticsearch ping failed")
		} else {
			cmdCtx.ESClient = esClient
			l.Info("Successfully connected to Elasticsearch")
		}
	}

	return cmdCtx, nil
}

func (c *CommandContext) GetMongoDatabase() *mongodri.Database {
	return c.DB
}

func (c *CommandContext) Close() {
	if c.disconnect != nil {
		c.disconnect()
	}
	if c.RedisClient != nil {
		_ = c.RedisClient.Close()
	}
}
