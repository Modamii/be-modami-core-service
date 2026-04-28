package migrations

import (
	"context"

	"be-modami-core-service/internal/command"
	mongodb "be-modami-core-service/pkg/mongodb"

	"gitlab.com/lifegoeson-libs/pkg-logging/logger"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func init() {
	command.RegisterMigration(command.Migration{
		Version:     "20260315_002",
		Description: "Create MongoDB indexes for all Modami collections",
		Up: func(ctx context.Context, db *mongo.Database) error {
			l := logger.FromContext(ctx)

			mongodb.EnsureIndexes(ctx, db)

			l.Info("Created MongoDB indexes for all collections")
			return nil
		},
	})
}
