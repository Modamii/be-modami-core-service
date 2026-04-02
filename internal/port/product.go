package port

import (
	"context"

	"be-modami-core-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ProductProducer publishes product domain events to the message broker.
type ProductProducer interface {
	ProductCreatedWithData(ctx context.Context, product *domain.Product) error
	ProductUpdatedWithData(ctx context.Context, product *domain.Product) error
	ProductDeleted(ctx context.Context, productID, slug string) error
}

// ProductRepository persists catalog products and related product_* collections.
type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	GetByID(ctx context.Context, id bson.ObjectID) (*domain.Product, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Product, error)
	Update(ctx context.Context, p *domain.Product) error
	SoftDelete(ctx context.Context, id bson.ObjectID) error

	ListBySellerID(ctx context.Context, sellerID string, status string, cursor string, limit int) ([]domain.Product, string, error)
	ListFeed(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error)
	ListFeatured(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error)
	ListSelect(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error)
	ListSimilar(ctx context.Context, productID bson.ObjectID, limit int) ([]domain.Product, error)
	ListByIDs(ctx context.Context, ids []bson.ObjectID) ([]domain.Product, error)

	InitStats(ctx context.Context, productID bson.ObjectID) error
	GetStats(ctx context.Context, productID bson.ObjectID) (*domain.ProductStats, error)
	IncrementStat(ctx context.Context, productID bson.ObjectID, field string, value int64) error

	CreateModeration(ctx context.Context, m *domain.ProductModeration) error
	ListModerations(ctx context.Context, productID bson.ObjectID) ([]domain.ProductModeration, error)
	GetLatestModeration(ctx context.Context, productID bson.ObjectID) (*domain.ProductModeration, error)
	ListPendingProducts(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error)
	CountByStatus(ctx context.Context, status domain.ProductStatus) (int64, error)

	CreateSelectProduct(ctx context.Context, sp *domain.SelectProduct) error
	GetSelectProduct(ctx context.Context, productID bson.ObjectID) (*domain.SelectProduct, error)

	Search(ctx context.Context, query string, params domain.ProductListParams, cursor string, limit int) ([]domain.Product, string, error)
	ListByHashtag(ctx context.Context, tag string, cursor string, limit int) ([]domain.Product, string, error)
}
