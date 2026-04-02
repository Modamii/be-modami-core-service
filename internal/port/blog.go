package port

import (
	"context"

	"be-modami-core-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// BlogRepository persists blog posts to the blog_posts collection.
type BlogRepository interface {
	Create(ctx context.Context, p *domain.BlogPost) error
	GetByID(ctx context.Context, id bson.ObjectID) (*domain.BlogPost, error)
	GetBySlug(ctx context.Context, slug string) (*domain.BlogPost, error)
	Update(ctx context.Context, p *domain.BlogPost) error
	Delete(ctx context.Context, id bson.ObjectID) error

	// GetFeatured returns the single post flagged as featured and published.
	GetFeatured(ctx context.Context) (*domain.BlogPost, error)

	// List returns a cursor-paginated list of published posts, optionally
	// filtered by postType (empty string = no filter).
	List(ctx context.Context, postType string, cursor string, limit int) ([]domain.BlogPost, string, error)

	// ListByHashtag returns published posts that carry the given hashtag.
	ListByHashtag(ctx context.Context, tag string, cursor string, limit int) ([]domain.BlogPost, string, error)

	// ListTrendReports returns published posts whose post_type is "trend_report",
	// sorted by published_at descending.
	ListTrendReports(ctx context.Context, cursor string, limit int) ([]domain.BlogPost, string, error)
}
