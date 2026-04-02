package migrations

import (
	"context"
	"fmt"

	"github.com/modami/core-service/internal/command"
	es "github.com/modami/core-service/pkg/elasticsearch"

	"gitlab.com/lifegoeson-libs/pkg-logging/logger"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func init() {
	command.RegisterMigration(command.Migration{
		Version:     "20260315_001",
		Description: "Create Elasticsearch products index with mappings",
		Up: func(ctx context.Context, db *mongo.Database) error {
			l := logger.FromContext(ctx)

			cfg := command.GetConfig()
			if cfg == nil {
				return fmt.Errorf("config not available")
			}

			esClient, err := es.NewClient(&es.Config{
				URL:      cfg.Elasticsearch.URL,
				Username: cfg.Elasticsearch.Username,
				Password: cfg.Elasticsearch.Password,
				Index:    cfg.Elasticsearch.Index,
			})
			if err != nil {
				return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
			}

			if err := esClient.Ping(); err != nil {
				return fmt.Errorf("elasticsearch is not reachable: %w", err)
			}

			if err := esClient.EnsureProductIndices(ctx); err != nil {
				return fmt.Errorf("failed to create ES products index: %w", err)
			}

			l.Info("Elasticsearch products index created successfully")
			return nil
		},
	})
}
