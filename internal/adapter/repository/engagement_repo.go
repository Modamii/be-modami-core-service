package repository

import (
	"context"
	"time"

	"be-modami-core-service/internal/domain"
	"be-modami-core-service/internal/port"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"be-modami-core-service/pkg/storage/database/mongodb/pagination"
)

// ---------------------------------------------------------------------------
// domain.Favorite
// ---------------------------------------------------------------------------

type mongoFavoriteRepo struct {
	col *mongo.Collection
}

func NewFavoriteRepository(db *mongo.Database) port.FavoriteRepository {
	return &mongoFavoriteRepo{col: db.Collection("favorites")}
}

func (r *mongoFavoriteRepo) Add(ctx context.Context, f *domain.Favorite) error {
	f.CreatedAt = time.Now()
	result, err := r.col.InsertOne(ctx, f)
	if err != nil {
		return err
	}
	f.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *mongoFavoriteRepo) Remove(ctx context.Context, userID, productID bson.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"user_id": userID, "product_id": productID})
	return err
}

func (r *mongoFavoriteRepo) ListByUser(ctx context.Context, userID bson.ObjectID, cursor string, limit int) ([]domain.Favorite, string, error) {
	filter := bson.M{"user_id": userID}
	if cursor != "" {
		cf, err := pagination.CursorFilter(cursor, "created_at")
		if err == nil && len(cf) > 0 {
			for _, elem := range cf {
				filter[elem.Key] = elem.Value
			}
		}
	}

	opts := options.Find().
		SetLimit(int64(limit + 1)).
		SetSort(bson.D{{"created_at", -1}, {"_id", -1}})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	var items []domain.Favorite
	if err := cur.All(ctx, &items); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(items) > limit {
		items = items[:limit]
		last := items[len(items)-1]
		nextCursor = pagination.EncodeCursor(last.ID.Hex(), last.CreatedAt)
	}
	return items, nextCursor, nil
}

func (r *mongoFavoriteRepo) Check(ctx context.Context, userID, productID bson.ObjectID) (bool, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"user_id": userID, "product_id": productID})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *mongoFavoriteRepo) CountByProduct(ctx context.Context, productID bson.ObjectID) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{"product_id": productID})
}

// ---------------------------------------------------------------------------
// domain.SavedProduct
// ---------------------------------------------------------------------------

type mongoSavedProductRepo struct {
	col *mongo.Collection
}

func NewSavedProductRepository(db *mongo.Database) port.SavedProductRepository {
	return &mongoSavedProductRepo{col: db.Collection("saved_products")}
}

func (r *mongoSavedProductRepo) Save(ctx context.Context, sp *domain.SavedProduct) error {
	sp.CreatedAt = time.Now()
	result, err := r.col.InsertOne(ctx, sp)
	if err != nil {
		return err
	}
	sp.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *mongoSavedProductRepo) Remove(ctx context.Context, userID, productID bson.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"user_id": userID, "product_id": productID})
	return err
}

func (r *mongoSavedProductRepo) ListByUser(ctx context.Context, userID bson.ObjectID, cursor string, limit int) ([]domain.SavedProduct, string, error) {
	filter := bson.M{"user_id": userID}
	if cursor != "" {
		cf, err := pagination.CursorFilter(cursor, "created_at")
		if err == nil && len(cf) > 0 {
			for _, elem := range cf {
				filter[elem.Key] = elem.Value
			}
		}
	}

	opts := options.Find().
		SetLimit(int64(limit + 1)).
		SetSort(bson.D{{"created_at", -1}, {"_id", -1}})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	var items []domain.SavedProduct
	if err := cur.All(ctx, &items); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(items) > limit {
		items = items[:limit]
		last := items[len(items)-1]
		nextCursor = pagination.EncodeCursor(last.ID.Hex(), last.CreatedAt)
	}
	return items, nextCursor, nil
}

func (r *mongoSavedProductRepo) Check(ctx context.Context, userID, productID bson.ObjectID) (bool, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"user_id": userID, "product_id": productID})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *mongoSavedProductRepo) MoveToCollection(ctx context.Context, userID, productID bson.ObjectID, collectionID *bson.ObjectID) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"user_id": userID, "product_id": productID},
		bson.M{"$set": bson.M{"collection_id": collectionID}},
	)
	return err
}

// ---------------------------------------------------------------------------
// domain.SavedCollection
// ---------------------------------------------------------------------------

type mongoSavedCollectionRepo struct {
	col *mongo.Collection
}

func NewSavedCollectionRepository(db *mongo.Database) port.SavedCollectionRepository {
	return &mongoSavedCollectionRepo{col: db.Collection("saved_collections")}
}

func (r *mongoSavedCollectionRepo) Create(ctx context.Context, sc *domain.SavedCollection) error {
	now := time.Now()
	sc.CreatedAt = now
	sc.UpdatedAt = now
	result, err := r.col.InsertOne(ctx, sc)
	if err != nil {
		return err
	}
	sc.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *mongoSavedCollectionRepo) List(ctx context.Context, userID bson.ObjectID) ([]domain.SavedCollection, error) {
	opts := options.Find().SetSort(bson.D{{"created_at", -1}})
	cur, err := r.col.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	var items []domain.SavedCollection
	if err := cur.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *mongoSavedCollectionRepo) Update(ctx context.Context, id, userID bson.ObjectID, name string) error {
	result, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id, "user_id": userID},
		bson.M{"$set": bson.M{"name": name, "updated_at": time.Now()}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *mongoSavedCollectionRepo) Delete(ctx context.Context, id, userID bson.ObjectID) error {
	result, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// ---------------------------------------------------------------------------
// domain.Follow
// ---------------------------------------------------------------------------

type mongoFollowRepo struct {
	col *mongo.Collection
}

func NewFollowRepository(db *mongo.Database) port.FollowRepository {
	return &mongoFollowRepo{col: db.Collection("follows")}
}

func (r *mongoFollowRepo) Follow(ctx context.Context, f *domain.Follow) error {
	f.CreatedAt = time.Now()
	result, err := r.col.InsertOne(ctx, f)
	if err != nil {
		return err
	}
	f.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *mongoFollowRepo) Unfollow(ctx context.Context, followerID, sellerID bson.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"follower_id": followerID, "seller_id": sellerID})
	return err
}

func (r *mongoFollowRepo) ListFollowing(ctx context.Context, followerID bson.ObjectID, cursor string, limit int) ([]domain.Follow, string, error) {
	filter := bson.M{"follower_id": followerID}
	return r.listWithCursor(ctx, filter, cursor, limit)
}

func (r *mongoFollowRepo) ListFollowers(ctx context.Context, sellerID bson.ObjectID, cursor string, limit int) ([]domain.Follow, string, error) {
	filter := bson.M{"seller_id": sellerID}
	return r.listWithCursor(ctx, filter, cursor, limit)
}

func (r *mongoFollowRepo) Check(ctx context.Context, followerID, sellerID bson.ObjectID) (bool, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"follower_id": followerID, "seller_id": sellerID})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *mongoFollowRepo) CountFollowers(ctx context.Context, sellerID bson.ObjectID) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{"seller_id": sellerID})
}

func (r *mongoFollowRepo) CountFollowing(ctx context.Context, followerID bson.ObjectID) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{"follower_id": followerID})
}

func (r *mongoFollowRepo) listWithCursor(ctx context.Context, filter bson.M, cursor string, limit int) ([]domain.Follow, string, error) {
	if cursor != "" {
		cf, err := pagination.CursorFilter(cursor, "created_at")
		if err == nil && len(cf) > 0 {
			for _, elem := range cf {
				filter[elem.Key] = elem.Value
			}
		}
	}

	opts := options.Find().
		SetLimit(int64(limit + 1)).
		SetSort(bson.D{{"created_at", -1}, {"_id", -1}})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	var items []domain.Follow
	if err := cur.All(ctx, &items); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(items) > limit {
		items = items[:limit]
		last := items[len(items)-1]
		nextCursor = pagination.EncodeCursor(last.ID.Hex(), last.CreatedAt)
	}
	return items, nextCursor, nil
}

// ---------------------------------------------------------------------------
// domain.Review
// ---------------------------------------------------------------------------

type mongoReviewRepo struct {
	col *mongo.Collection
}

func NewReviewRepository(db *mongo.Database) port.ReviewRepository {
	return &mongoReviewRepo{col: db.Collection("reviews")}
}

func (r *mongoReviewRepo) Create(ctx context.Context, rv *domain.Review) error {
	now := time.Now()
	rv.CreatedAt = now
	rv.UpdatedAt = now
	result, err := r.col.InsertOne(ctx, rv)
	if err != nil {
		return err
	}
	rv.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *mongoReviewRepo) ListBySeller(ctx context.Context, sellerID bson.ObjectID, cursor string, limit int) ([]domain.Review, string, error) {
	filter := bson.M{"seller_id": sellerID}
	return r.listWithCursor(ctx, filter, cursor, limit)
}

func (r *mongoReviewRepo) ListByProduct(ctx context.Context, productID bson.ObjectID, cursor string, limit int) ([]domain.Review, string, error) {
	filter := bson.M{"product_id": productID}
	return r.listWithCursor(ctx, filter, cursor, limit)
}

func (r *mongoReviewRepo) ListByBuyer(ctx context.Context, buyerID bson.ObjectID, cursor string, limit int) ([]domain.Review, string, error) {
	filter := bson.M{"buyer_id": buyerID}
	return r.listWithCursor(ctx, filter, cursor, limit)
}

func (r *mongoReviewRepo) GetByOrderID(ctx context.Context, orderID bson.ObjectID) (*domain.Review, error) {
	var rv domain.Review
	err := r.col.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&rv)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &rv, nil
}

func (r *mongoReviewRepo) listWithCursor(ctx context.Context, filter bson.M, cursor string, limit int) ([]domain.Review, string, error) {
	if cursor != "" {
		cf, err := pagination.CursorFilter(cursor, "created_at")
		if err == nil && len(cf) > 0 {
			for _, elem := range cf {
				filter[elem.Key] = elem.Value
			}
		}
	}

	opts := options.Find().
		SetLimit(int64(limit + 1)).
		SetSort(bson.D{{"created_at", -1}, {"_id", -1}})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	var items []domain.Review
	if err := cur.All(ctx, &items); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(items) > limit {
		items = items[:limit]
		last := items[len(items)-1]
		nextCursor = pagination.EncodeCursor(last.ID.Hex(), last.CreatedAt)
	}
	return items, nextCursor, nil
}
