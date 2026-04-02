package repository

import (
	"context"
	"fmt"
	"time"

	"be-modami-core-service/internal/domain"
	"be-modami-core-service/internal/port"

	apperror "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// domain.Category MongoDB implementation
type mongoCategoryRepo struct {
	col *mongo.Collection
}

func NewCategoryRepository(db *mongo.Database) port.CategoryRepository {
	return &mongoCategoryRepo{col: db.Collection("categories")}
}

func (r *mongoCategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	c.IsActive = true
	result, err := r.col.InsertOne(ctx, c)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return apperror.New(apperror.CodeConflict, "slug danh mục đã tồn tại")
		}
		return apperror.New(apperror.CodeInternal, fmt.Sprintf("tạo danh mục thất bại: %v", err))
	}
	c.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *mongoCategoryRepo) GetByID(ctx context.Context, id bson.ObjectID) (*domain.Category, error) {
	var c domain.Category
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *mongoCategoryRepo) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var c domain.Category
	err := r.col.FindOne(ctx, bson.M{"slug": slug}).Decode(&c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *mongoCategoryRepo) Update(ctx context.Context, c *domain.Category) error {
	c.UpdatedAt = time.Now()
	_, err := r.col.ReplaceOne(ctx, bson.M{"_id": c.ID}, c)
	return err
}

func (r *mongoCategoryRepo) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *mongoCategoryRepo) ListAll(ctx context.Context, activeOnly bool) ([]domain.Category, error) {
	filter := bson.M{}
	if activeOnly {
		filter["is_active"] = true
	}
	opts := options.Find().SetSort(bson.D{{"sort_order", 1}})
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var cats []domain.Category
	if err := cur.All(ctx, &cats); err != nil {
		return nil, err
	}
	return cats, nil
}

func (r *mongoCategoryRepo) ListChildren(ctx context.Context, parentID bson.ObjectID) ([]domain.Category, error) {
	opts := options.Find().SetSort(bson.D{{"sort_order", 1}})
	cur, err := r.col.Find(ctx, bson.M{"parent_id": parentID, "is_active": true}, opts)
	if err != nil {
		return nil, err
	}
	var cats []domain.Category
	if err := cur.All(ctx, &cats); err != nil {
		return nil, err
	}
	return cats, nil
}

func (r *mongoCategoryRepo) IncrementProductCount(ctx context.Context, id bson.ObjectID, delta int64) error {
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"product_count": delta}})
	return err
}

// domain.Hashtag MongoDB implementation
type mongoHashtagRepo struct {
	col *mongo.Collection
}

func NewHashtagRepository(db *mongo.Database) port.HashtagRepository {
	return &mongoHashtagRepo{col: db.Collection("hashtags")}
}

func (r *mongoHashtagRepo) Upsert(ctx context.Context, tag string, delta int64) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": tag},
		bson.M{
			"$inc": bson.M{"usage_count": delta},
			"$set": bson.M{"updated_at": time.Now()},
		},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (r *mongoHashtagRepo) ListTrending(ctx context.Context, limit int) ([]domain.Hashtag, error) {
	opts := options.Find().SetSort(bson.D{{"usage_count", -1}}).SetLimit(int64(limit))
	cur, err := r.col.Find(ctx, bson.M{"usage_count": bson.M{"$gt": 0}}, opts)
	if err != nil {
		return nil, err
	}
	var tags []domain.Hashtag
	if err := cur.All(ctx, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *mongoHashtagRepo) Search(ctx context.Context, query string, limit int) ([]domain.Hashtag, error) {
	filter := bson.M{"_id": bson.M{"$regex": query, "$options": "i"}}
	opts := options.Find().SetSort(bson.D{{"usage_count", -1}}).SetLimit(int64(limit))
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var tags []domain.Hashtag
	if err := cur.All(ctx, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}
