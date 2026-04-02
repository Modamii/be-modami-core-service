package main

import (
	"context"
	"fmt"

	"be-modami-core-service/config"
	"be-modami-core-service/pkg/elasticsearch"
	"be-modami-core-service/pkg/storage/database/mongodb"
	redisstorage "be-modami-core-service/pkg/storage/redis"

	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Connections holds infrastructure clients. Optional fields are nil when disabled or on error (caller may log).
type Connections struct {
	DB            *mongo.Database
	Redis         *redis.Client
	Kafka         *kgo.Client
	Elasticsearch *elasticsearch.Client

	closeMongo func()
}

func newConnections(ctx context.Context, cfg *config.Config) (*Connections, error) {
	db, disconnectMongo, err := mongodb.Connect(ctx, cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		return nil, fmt.Errorf("mongo: %w", err)
	}
	c := &Connections{
		DB:         db,
		closeMongo: disconnectMongo,
	}

	if cfg.Redis.Host != "" {
		rcfg := redisstorage.RedisConfig{
			Addr:         cfg.Redis.Addr(),
			Password:     cfg.Redis.Pass,
			DB:           cfg.Redis.Database,
			PoolSize:     cfg.Redis.PoolSize,
			DialTimeout:  cfg.Redis.DialTimeout,
			ReadTimeout:  cfg.Redis.ReadTimeout,
			WriteTimeout: cfg.Redis.WriteTimeout,
		}
		rcli, err := redisstorage.NewRedisClient(rcfg)
		if err != nil {
			disconnectMongo()
			return nil, fmt.Errorf("redis: %w", err)
		}
		c.Redis = rcli
	}

	if cfg.Kafka.Enable && len(cfg.Kafka.Brokers()) > 0 {
		kcl, err := kgo.NewClient(
			kgo.SeedBrokers(cfg.Kafka.Brokers()...),
			kgo.ClientID(cfg.Kafka.ClientID),
		)
		if err != nil {
			c.closeAll(ctx)
			return nil, fmt.Errorf("kafka: %w", err)
		}
		c.Kafka = kcl
	}

	if cfg.Elasticsearch.Enable && cfg.Elasticsearch.URL != "" {
		es, err := elasticsearch.NewClient(&elasticsearch.Config{
			URL:      cfg.Elasticsearch.URL,
			Username: cfg.Elasticsearch.Username,
			Password: cfg.Elasticsearch.Password,
			Index:    cfg.Elasticsearch.Index,
		})
		if err != nil {
			c.closeAll(ctx)
			return nil, fmt.Errorf("elasticsearch: %w", err)
		}
		c.Elasticsearch = es
	}

	return c, nil
}

func (c *Connections) closeAll(ctx context.Context) {
	if c == nil {
		return
	}
	if c.closeMongo != nil {
		c.closeMongo()
		c.closeMongo = nil
	}
	if c.Redis != nil {
		_ = redisstorage.CloseRedis(ctx, c.Redis)
		c.Redis = nil
	}
	if c.Kafka != nil {
		c.Kafka.Close()
		c.Kafka = nil
	}
	c.Elasticsearch = nil
}

// Disconnect releases all resources (Mongo, Redis, Kafka).
func (c *Connections) Disconnect(ctx context.Context) {
	if c == nil {
		return
	}
	if c.closeMongo != nil {
		c.closeMongo()
		c.closeMongo = nil
	}
	if c.Redis != nil {
		_ = redisstorage.CloseRedis(ctx, c.Redis)
		c.Redis = nil
	}
	if c.Kafka != nil {
		c.Kafka.Close()
		c.Kafka = nil
	}
	c.Elasticsearch = nil
}
