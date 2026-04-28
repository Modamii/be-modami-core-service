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
		{Keys: bson.D{{Key: "seller_id", Value: 1}, {Key: "status", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "category._id", Value: 1}, {Key: "published_at", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "is_featured", Value: 1}, {Key: "published_at", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "is_select", Value: 1}, {Key: "published_at", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "is_verified", Value: 1}, {Key: "published_at", Value: -1}}},
		{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "hashtags", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "price", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "brand", Value: 1}, {Key: "status", Value: 1}}},
		{
			Keys:    bson.D{{Key: "deleted_at", Value: 1}},
			Options: options.Index().SetPartialFilterExpression(bson.D{{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: true}}}}),
		},
	})

	// product_moderations
	createIndexes(ctx, db.Collection("product_moderations"), []mongo.IndexModel{
		{Keys: bson.D{{Key: "product_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "action", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "moderator_id", Value: 1}, {Key: "created_at", Value: -1}}},
	})

	// categories
	createIndexes(ctx, db.Collection("categories"), []mongo.IndexModel{
		{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "parent_id", Value: 1}, {Key: "sort_order", Value: 1}}},
		{Keys: bson.D{{Key: "is_active", Value: 1}, {Key: "sort_order", Value: 1}}},
	})

	// favorites
	createIndexes(ctx, db.Collection("favorites"), []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "product_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
	})

	// saved_products
	createIndexes(ctx, db.Collection("saved_products"), []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "product_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "collection_id", Value: 1}, {Key: "created_at", Value: -1}}},
	})

	// saved_collections
	createIndexes(ctx, db.Collection("saved_collections"), []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
	})

	// follows
	createIndexes(ctx, db.Collection("follows"), []mongo.IndexModel{
		{Keys: bson.D{{Key: "follower_id", Value: 1}, {Key: "seller_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "seller_id", Value: 1}}},
		{Keys: bson.D{{Key: "follower_id", Value: 1}, {Key: "created_at", Value: -1}}},
	})

	// reviews
	createIndexes(ctx, db.Collection("reviews"), []mongo.IndexModel{
		{Keys: bson.D{{Key: "order_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "seller_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "product_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "buyer_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "seller_id", Value: 1}, {Key: "rating", Value: 1}}},
	})

	// reports
	createIndexes(ctx, db.Collection("reports"), []mongo.IndexModel{
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "target_type", Value: 1}, {Key: "target_id", Value: 1}}},
		{Keys: bson.D{{Key: "reporter_id", Value: 1}, {Key: "created_at", Value: -1}}},
	})

	// hashtags
	createIndexes(ctx, db.Collection("hashtags"), []mongo.IndexModel{
		{Keys: bson.D{{Key: "usage_count", Value: -1}}},
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
