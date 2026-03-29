package mongodb

import (
	"context"
	"time"

	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Connect opens a MongoDB database (driver v2) and returns a disconnect function.
func Connect(ctx context.Context, uri, dbName string) (*mongo.Database, func(), error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, err
	}

	logger.Info(ctx, "connected to MongoDB", logging.String("db", dbName))

	disconnect := func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dcancel()
		if err := client.Disconnect(dctx); err != nil {
			logger.Error(dctx, "failed to disconnect from MongoDB", err)
		}
	}

	return client.Database(dbName), disconnect, nil
}
