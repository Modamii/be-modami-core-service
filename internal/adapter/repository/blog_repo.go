package repository

import (
	"context"
	"time"

	"be-modami-core-service/internal/domain"
	"be-modami-core-service/internal/port"
	"be-modami-core-service/pkg/storage/database/mongodb/pagination"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const trendReportPostType = "trend_report"

type blogMongoRepository struct {
	posts *mongo.Collection
}

// NewBlogRepository returns a MongoDB-backed blog repository.
func NewBlogRepository(db *mongo.Database) port.BlogRepository {
	return &blogMongoRepository{
		posts: db.Collection("blog_posts"),
	}
}

func (r *blogMongoRepository) Create(ctx context.Context, p *domain.BlogPost) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	if p.Status == "" {
		p.Status = domain.PostStatusDraft
	}
	result, err := r.posts.InsertOne(ctx, p)
	if err != nil {
		return err
	}
	p.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *blogMongoRepository) GetByID(ctx context.Context, id bson.ObjectID) (*domain.BlogPost, error) {
	var p domain.BlogPost
	err := r.posts.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *blogMongoRepository) GetBySlug(ctx context.Context, slug string) (*domain.BlogPost, error) {
	var p domain.BlogPost
	err := r.posts.FindOne(ctx, bson.M{"slug": slug, "status": domain.PostStatusPublished}).Decode(&p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *blogMongoRepository) Update(ctx context.Context, p *domain.BlogPost) error {
	p.UpdatedAt = time.Now()
	_, err := r.posts.ReplaceOne(ctx, bson.M{"_id": p.ID}, p)
	return err
}

func (r *blogMongoRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.posts.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *blogMongoRepository) GetFeatured(ctx context.Context) (*domain.BlogPost, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "published_at", Value: -1}})
	var p domain.BlogPost
	err := r.posts.FindOne(ctx, bson.M{
		"status":      domain.PostStatusPublished,
		"is_featured": true,
	}, opts).Decode(&p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *blogMongoRepository) List(ctx context.Context, postType string, cursor string, limit int) ([]domain.BlogPost, string, error) {
	filter := bson.M{"status": domain.PostStatusPublished}
	if postType != "" {
		filter["post_type"] = postType
	}
	return r.listWithCursor(ctx, filter, cursor, limit)
}

func (r *blogMongoRepository) ListByHashtag(ctx context.Context, tag string, cursor string, limit int) ([]domain.BlogPost, string, error) {
	filter := bson.M{
		"status":   domain.PostStatusPublished,
		"hashtags": tag,
	}
	return r.listWithCursor(ctx, filter, cursor, limit)
}

func (r *blogMongoRepository) ListTrendReports(ctx context.Context, cursor string, limit int) ([]domain.BlogPost, string, error) {
	filter := bson.M{
		"status":    domain.PostStatusPublished,
		"post_type": trendReportPostType,
	}
	return r.listWithCursor(ctx, filter, cursor, limit)
}

// listWithCursor implements the shared cursor-pagination pattern used across the codebase.
// It sorts by published_at DESC, _id DESC and fetches limit+1 items to detect a next page.
func (r *blogMongoRepository) listWithCursor(ctx context.Context, filter bson.M, cursor string, limit int) ([]domain.BlogPost, string, error) {
	if cursor != "" {
		cursorFilter, err := pagination.CursorFilter(cursor, "published_at")
		if err == nil && len(cursorFilter) > 0 {
			for _, elem := range cursorFilter {
				filter[elem.Key] = elem.Value
			}
		}
	}

	opts := options.Find().
		SetLimit(int64(limit + 1)).
		SetSort(bson.D{{Key: "published_at", Value: -1}, {Key: "_id", Value: -1}})

	cur, err := r.posts.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}

	var posts []domain.BlogPost
	if err := cur.All(ctx, &posts); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(posts) > limit {
		posts = posts[:limit]
		last := posts[len(posts)-1]
		t := last.CreatedAt
		if last.PublishedAt != nil {
			t = *last.PublishedAt
		}
		nextCursor = pagination.EncodeCursor(last.ID.Hex(), t)
	}

	return posts, nextCursor, nil
}
