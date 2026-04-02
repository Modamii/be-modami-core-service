package port

import (
	"context"

	"be-modami-core-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type FavoriteRepository interface {
	Add(ctx context.Context, f *domain.Favorite) error
	Remove(ctx context.Context, userID string, productID bson.ObjectID) error
	ListByUser(ctx context.Context, userID string, cursor string, limit int) ([]domain.Favorite, string, error)
	Check(ctx context.Context, userID string, productID bson.ObjectID) (bool, error)
	CountByProduct(ctx context.Context, productID bson.ObjectID) (int64, error)
}

type SavedProductRepository interface {
	Save(ctx context.Context, sp *domain.SavedProduct) error
	Remove(ctx context.Context, userID string, productID bson.ObjectID) error
	ListByUser(ctx context.Context, userID string, cursor string, limit int) ([]domain.SavedProduct, string, error)
	Check(ctx context.Context, userID string, productID bson.ObjectID) (bool, error)
	MoveToCollection(ctx context.Context, userID string, productID bson.ObjectID, collectionID *bson.ObjectID) error
}

type SavedCollectionRepository interface {
	Create(ctx context.Context, sc *domain.SavedCollection) error
	List(ctx context.Context, userID string) ([]domain.SavedCollection, error)
	Update(ctx context.Context, id bson.ObjectID, userID string, name string) error
	Delete(ctx context.Context, id bson.ObjectID, userID string) error
}

type FollowRepository interface {
	Follow(ctx context.Context, f *domain.Follow) error
	Unfollow(ctx context.Context, followerID, sellerID string) error
	ListFollowing(ctx context.Context, followerID string, cursor string, limit int) ([]domain.Follow, string, error)
	ListFollowers(ctx context.Context, sellerID string, cursor string, limit int) ([]domain.Follow, string, error)
	Check(ctx context.Context, followerID, sellerID string) (bool, error)
	CountFollowers(ctx context.Context, sellerID string) (int64, error)
	CountFollowing(ctx context.Context, followerID string) (int64, error)
}

type ReviewRepository interface {
	Create(ctx context.Context, rv *domain.Review) error
	ListBySeller(ctx context.Context, sellerID string, cursor string, limit int) ([]domain.Review, string, error)
	ListByProduct(ctx context.Context, productID bson.ObjectID, cursor string, limit int) ([]domain.Review, string, error)
	ListByBuyer(ctx context.Context, buyerID string, cursor string, limit int) ([]domain.Review, string, error)
	GetByOrderID(ctx context.Context, orderID bson.ObjectID) (*domain.Review, error)
}
