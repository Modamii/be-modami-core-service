package port

import (
	"context"

	"github.com/modami/core-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CategoryRepository interface {
	Create(ctx context.Context, c *domain.Category) error
	GetByID(ctx context.Context, id bson.ObjectID) (*domain.Category, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Category, error)
	Update(ctx context.Context, c *domain.Category) error
	Delete(ctx context.Context, id bson.ObjectID) error
	ListAll(ctx context.Context, activeOnly bool) ([]domain.Category, error)
	ListChildren(ctx context.Context, parentID bson.ObjectID) ([]domain.Category, error)
	IncrementProductCount(ctx context.Context, id bson.ObjectID, delta int64) error
}

type PackageRepository interface {
	Create(ctx context.Context, p *domain.Package) error
	GetByID(ctx context.Context, id bson.ObjectID) (*domain.Package, error)
	GetByCode(ctx context.Context, code string) (*domain.Package, error)
	Update(ctx context.Context, p *domain.Package) error
	ListActive(ctx context.Context) ([]domain.Package, error)
}

type HashtagRepository interface {
	Upsert(ctx context.Context, tag string, delta int64) error
	ListTrending(ctx context.Context, limit int) ([]domain.Hashtag, error)
	Search(ctx context.Context, query string, limit int) ([]domain.Hashtag, error)
}
