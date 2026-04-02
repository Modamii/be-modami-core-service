package command

import (
	"context"
	"fmt"

	config "github.com/modami/core-service/config"
	es "github.com/modami/core-service/pkg/elasticsearch"
	mongodb "github.com/modami/core-service/pkg/storage/database/mongodb"
	redisStorage "github.com/modami/core-service/pkg/storage/redis"

	"gitlab.com/lifegoeson-libs/pkg-logging/logger"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
	DB          *mongo.Database
	disconnect  func()
	RedisClient *redis.Client
	ESClient    *es.Client
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
	db, disconnect, err := mongodb.Connect(ctx, cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	cmdCtx.DB = db
	cmdCtx.disconnect = disconnect

	// Connect Redis
	l.Info("Connecting to Redis...")
	redisClient, err := redisStorage.NewRedisClient(redisStorage.RedisConfig{
		Addr:         cfg.Redis.Addr(),
		Password:     cfg.Redis.Pass,
		DB:           cfg.Redis.Database,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	if err != nil {
		cmdCtx.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	cmdCtx.RedisClient = redisClient

	// Connect Elasticsearch
	l.Info("Connecting to Elasticsearch...")
	esClient, err := es.NewClient(&es.Config{
		URL:      cfg.Elasticsearch.URL,
		Username: cfg.Elasticsearch.Username,
		Password: cfg.Elasticsearch.Password,
		Index:    cfg.Elasticsearch.Index,
	})
	if err != nil {
		l.Warn("Failed to create Elasticsearch client")
	} else {
		if pingErr := esClient.Ping(); pingErr != nil {
			l.Warn("Elasticsearch ping failed")
		} else {
			cmdCtx.ESClient = esClient
			l.Info("Successfully connected to Elasticsearch")
		}
	}

	return cmdCtx, nil
}

func (c *CommandContext) GetMongoDatabase() *mongo.Database {
	return c.DB
}

func (c *CommandContext) Close() {
	if c.disconnect != nil {
		c.disconnect()
	}
	if c.RedisClient != nil {
		c.RedisClient.Close()
	}
}
