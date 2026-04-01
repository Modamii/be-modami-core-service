package mongodb

import (
	"context"
	"time"

	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func EnsureIndexes(ctx context.Context, db *mongo.Database) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// products
	createIndexes(ctx, db.Collection("products"), []mongo.IndexModel{
		{Keys: bson.D{{"seller_id", 1}, {"status", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"status", 1}, {"category._id", 1}, {"published_at", -1}}},
		{Keys: bson.D{{"status", 1}, {"is_featured", 1}, {"published_at", -1}}},
		{Keys: bson.D{{"status", 1}, {"is_select", 1}, {"published_at", -1}}},
		{Keys: bson.D{{"status", 1}, {"is_verified", 1}, {"published_at", -1}}},
		{Keys: bson.D{{"slug", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"hashtags", 1}, {"status", 1}}},
		{Keys: bson.D{{"price", 1}, {"status", 1}}},
		{Keys: bson.D{{"brand", 1}, {"status", 1}}},
		{
			Keys:    bson.D{{"deleted_at", 1}},
			Options: options.Index().SetPartialFilterExpression(bson.D{{"deleted_at", bson.D{{"$exists", true}}}}),
		},
	})

	// product_moderations
	createIndexes(ctx, db.Collection("product_moderations"), []mongo.IndexModel{
		{Keys: bson.D{{"product_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"action", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"moderator_id", 1}, {"created_at", -1}}},
	})

	// categories
	createIndexes(ctx, db.Collection("categories"), []mongo.IndexModel{
		{Keys: bson.D{{"slug", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"parent_id", 1}, {"sort_order", 1}}},
		{Keys: bson.D{{"is_active", 1}, {"sort_order", 1}}},
	})

	// favorites
	createIndexes(ctx, db.Collection("favorites"), []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}, {"product_id", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"user_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"product_id", 1}}},
	})

	// saved_products
	createIndexes(ctx, db.Collection("saved_products"), []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}, {"product_id", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"user_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"user_id", 1}, {"collection_id", 1}, {"created_at", -1}}},
	})

	// saved_collections
	createIndexes(ctx, db.Collection("saved_collections"), []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}, {"created_at", -1}}},
	})

	// follows
	createIndexes(ctx, db.Collection("follows"), []mongo.IndexModel{
		{Keys: bson.D{{"follower_id", 1}, {"seller_id", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"seller_id", 1}}},
		{Keys: bson.D{{"follower_id", 1}, {"created_at", -1}}},
	})

	// reviews
	createIndexes(ctx, db.Collection("reviews"), []mongo.IndexModel{
		{Keys: bson.D{{"order_id", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"seller_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"product_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"buyer_id", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"seller_id", 1}, {"rating", 1}}},
	})

	// reports
	createIndexes(ctx, db.Collection("reports"), []mongo.IndexModel{
		{Keys: bson.D{{"status", 1}, {"created_at", -1}}},
		{Keys: bson.D{{"target_type", 1}, {"target_id", 1}}},
		{Keys: bson.D{{"reporter_id", 1}, {"created_at", -1}}},
	})

	// hashtags
	createIndexes(ctx, db.Collection("hashtags"), []mongo.IndexModel{
		{Keys: bson.D{{"usage_count", -1}}},
	})

	// blog_posts
	createIndexes(ctx, db.Collection("blog_posts"), []mongo.IndexModel{
		{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "published_at", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "is_featured", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "post_type", Value: 1}, {Key: "published_at", Value: -1}}},
		{Keys: bson.D{{Key: "hashtags", Value: 1}, {Key: "status", Value: 1}}},
	})

	logger.Info(ctx, "all indexes ensured")
}

func createIndexes(ctx context.Context, col *mongo.Collection, models []mongo.IndexModel) {
	_, err := col.Indexes().CreateMany(ctx, models)
	if err != nil {
		logger.Warn(ctx, "failed to create indexes", logging.String("collection", col.Name()), logging.String("error", err.Error()))
	}
}
