package main

import (
	"context"
	"fmt"
	"time"

	"be-modami-core-service/config"
	localkafka "be-modami-core-service/pkg/kafka"

	pkges "gitlab.com/lifegoeson-libs/pkg-gokit/elasticsearch"
	pkgkafka "gitlab.com/lifegoeson-libs/pkg-gokit/kafka"
	pkgmongo "gitlab.com/lifegoeson-libs/pkg-gokit/mongodb"
	pkgredis "gitlab.com/lifegoeson-libs/pkg-gokit/redis"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
	mongodri "go.mongodb.org/mongo-driver/v2/mongo"
)

// Connections holds infrastructure clients. Optional fields are nil when disabled or on error (caller may log).
type Connections struct {
	DB            *mongodri.Database
	Redis         pkgredis.CachePort
	Kafka         *pkgkafka.KafkaService
	Elasticsearch *pkges.Client
	closeMongo func()
}

func newConnections(ctx context.Context, cfg *config.Config) (*Connections, error) {
	db, disconnectMongo, err := pkgmongo.Connect(ctx, cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		return nil, fmt.Errorf("mongo: %w", err)
	}
	c := &Connections{
		DB:         db,
		closeMongo: disconnectMongo,
	}

	// Redis
	if cfg.Redis.Host != "" {
		redisCfg := pkgredis.Config{
			Addrs:       []string{fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)},
			Password:    cfg.Redis.Pass,
			DB:          cfg.Redis.Database,
			PoolSize:    cfg.Redis.PoolSize,
			DialTimeout: 5 * time.Second,
		}
		adapter, err := pkgredis.NewAdapter(redisCfg)
		if err != nil {
			logger.Warn(ctx, "failed to connect to Redis, cache features will be disabled", logging.Any("error", err.Error()))
		} else {
			c.Redis = adapter
			logger.Info(ctx, "Redis connected", logging.String("addr", fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)))
		}
	} else {
		logger.Warn(ctx, "redis host not set, cache features will be disabled")
	}

	if len(cfg.Kafka.Brokers()) > 0 {
		kafkaCfg := &pkgkafka.Config{
			Brokers:          cfg.Kafka.Brokers(),
			ClientID:         cfg.Kafka.ClientID,
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
			c.closeAll(ctx)
			return nil, fmt.Errorf("kafka: %w", err)
		}
		c.Kafka = ks
	}

	if cfg.Elasticsearch.URL != "" {
		es, err := pkges.Connect(ctx, pkges.Config{
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
		_ = c.Redis.Close()
		c.Redis = nil
	}
	if c.Kafka != nil {
		_ = c.Kafka.Close()
		c.Kafka = nil
	}
	c.Elasticsearch = nil
}

// Disconnect releases all resources.
func (c *Connections) Disconnect(ctx context.Context) {
	c.closeAll(ctx)
}
