package repository

import (
	"context"
	"time"

	"github.com/modami/core-service/internal/domain"
	"github.com/modami/core-service/internal/port"
	"github.com/modami/core-service/pkg/storage/database/mongodb/pagination"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type productMongoRepository struct {
	products    *mongo.Collection
	stats       *mongo.Collection
	moderations *mongo.Collection
	selects     *mongo.Collection
}

// NewProductRepository returns a MongoDB-backed product catalog repository.
func NewProductRepository(db *mongo.Database) port.ProductRepository {
	return &productMongoRepository{
		products:    db.Collection("products"),
		stats:       db.Collection("product_stats"),
		moderations: db.Collection("product_moderations"),
		selects:     db.Collection("select_products"),
	}
}

func (r *productMongoRepository) Create(ctx context.Context, p *domain.Product) error {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	p.Version = 1
	if p.Status == "" {
		p.Status = domain.StatusDraft
	}
	result, err := r.products.InsertOne(ctx, p)
	if err != nil {
		return err
	}
	p.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *productMongoRepository) GetByID(ctx context.Context, id bson.ObjectID) (*domain.Product, error) {
	var p domain.Product
	err := r.products.FindOne(ctx, bson.M{"_id": id, "deleted_at": nil}).Decode(&p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *productMongoRepository) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	var p domain.Product
	err := r.products.FindOne(ctx, bson.M{"slug": slug, "deleted_at": nil}).Decode(&p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *productMongoRepository) Update(ctx context.Context, p *domain.Product) error {
	p.UpdatedAt = time.Now()
	filter := bson.M{"_id": p.ID, "version": p.Version}
	p.Version++
	result, err := r.products.ReplaceOne(ctx, filter, p)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrProductVersionConflict
	}
	return nil
}

func (r *productMongoRepository) SoftDelete(ctx context.Context, id bson.ObjectID) error {
	now := time.Now()
	_, err := r.products.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
	})
	return err
}

func (r *productMongoRepository) ListBySellerID(ctx context.Context, sellerID bson.ObjectID, status string, cursor string, limit int) ([]domain.Product, string, error) {
	filter := bson.M{"seller_id": sellerID, "deleted_at": nil}
	if status != "" {
		filter["status"] = status
	}
	return r.listWithCursor(ctx, filter, cursor, limit, "created_at")
}

func (r *productMongoRepository) ListFeed(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error) {
	filter := bson.M{"status": domain.StatusActive, "deleted_at": nil}
	return r.listWithCursor(ctx, filter, cursor, limit, "published_at")
}

func (r *productMongoRepository) ListFeatured(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error) {
	filter := bson.M{"status": domain.StatusActive, "is_featured": true, "deleted_at": nil}
	return r.listWithCursor(ctx, filter, cursor, limit, "published_at")
}

func (r *productMongoRepository) ListSelect(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error) {
	filter := bson.M{"status": domain.StatusActive, "is_select": true, "deleted_at": nil}
	return r.listWithCursor(ctx, filter, cursor, limit, "published_at")
}

func (r *productMongoRepository) ListSimilar(ctx context.Context, productID bson.ObjectID, limit int) ([]domain.Product, error) {
	p, err := r.GetByID(ctx, productID)
	if err != nil || p == nil {
		return nil, err
	}

	catFilter := bson.M{"$exists": false}
	if p.Category != nil {
		catFilter = bson.M{"$eq": p.Category.ID}
	}
	filter := bson.M{
		"status":       domain.StatusActive,
		"category._id": catFilter,
		"_id":          bson.M{"$ne": productID},
		"deleted_at":   nil,
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "published_at", Value: -1}})
	cur, err := r.products.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var products []domain.Product
	if err := cur.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productMongoRepository) ListByIDs(ctx context.Context, ids []bson.ObjectID) ([]domain.Product, error) {
	cur, err := r.products.Find(ctx, bson.M{"_id": bson.M{"$in": ids}, "deleted_at": nil})
	if err != nil {
		return nil, err
	}
	var products []domain.Product
	if err := cur.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productMongoRepository) InitStats(ctx context.Context, productID bson.ObjectID) error {
	_, err := r.stats.InsertOne(ctx, &domain.ProductStats{
		ProductID: productID,
		UpdatedAt: time.Now(),
	})
	return err
}

func (r *productMongoRepository) GetStats(ctx context.Context, productID bson.ObjectID) (*domain.ProductStats, error) {
	var s domain.ProductStats
	err := r.stats.FindOne(ctx, bson.M{"_id": productID}).Decode(&s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &domain.ProductStats{ProductID: productID}, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *productMongoRepository) IncrementStat(ctx context.Context, productID bson.ObjectID, field string, value int64) error {
	_, err := r.stats.UpdateOne(ctx,
		bson.M{"_id": productID},
		bson.M{
			"$inc": bson.M{field: value},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *productMongoRepository) CreateModeration(ctx context.Context, m *domain.ProductModeration) error {
	m.CreatedAt = time.Now()
	result, err := r.moderations.InsertOne(ctx, m)
	if err != nil {
		return err
	}
	m.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *productMongoRepository) ListModerations(ctx context.Context, productID bson.ObjectID) ([]domain.ProductModeration, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.moderations.Find(ctx, bson.M{"product_id": productID}, opts)
	if err != nil {
		return nil, err
	}
	var mods []domain.ProductModeration
	if err := cur.All(ctx, &mods); err != nil {
		return nil, err
	}
	return mods, nil
}

func (r *productMongoRepository) GetLatestModeration(ctx context.Context, productID bson.ObjectID) (*domain.ProductModeration, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	var m domain.ProductModeration
	err := r.moderations.FindOne(ctx, bson.M{"product_id": productID}, opts).Decode(&m)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *productMongoRepository) ListPendingProducts(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error) {
	filter := bson.M{"status": domain.StatusPending, "deleted_at": nil}
	return r.listWithCursor(ctx, filter, cursor, limit, "created_at")
}

func (r *productMongoRepository) CountByStatus(ctx context.Context, status domain.ProductStatus) (int64, error) {
	return r.products.CountDocuments(ctx, bson.M{"status": status, "deleted_at": nil})
}

func (r *productMongoRepository) CreateSelectProduct(ctx context.Context, sp *domain.SelectProduct) error {
	sp.CreatedAt = time.Now()
	_, err := r.selects.InsertOne(ctx, sp)
	return err
}

func (r *productMongoRepository) GetSelectProduct(ctx context.Context, productID bson.ObjectID) (*domain.SelectProduct, error) {
	var sp domain.SelectProduct
	err := r.selects.FindOne(ctx, bson.M{"_id": productID}).Decode(&sp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &sp, nil
}

func (r *productMongoRepository) Search(ctx context.Context, query string, params domain.ProductListParams, cursor string, limit int) ([]domain.Product, string, error) {
	filter := bson.M{"status": domain.StatusActive, "deleted_at": nil}

	if query != "" {
		filter["$or"] = bson.A{
			bson.M{"title": bson.M{"$regex": query, "$options": "i"}},
			bson.M{"description": bson.M{"$regex": query, "$options": "i"}},
			bson.M{"brand": bson.M{"$regex": query, "$options": "i"}},
		}
	}

	if params.CategoryID != "" {
		if oid, err := bson.ObjectIDFromHex(params.CategoryID); err == nil {
			filter["category._id"] = oid
		}
	}
	if params.Condition != "" {
		filter["condition"] = params.Condition
	}
	if params.Brand != "" {
		filter["brand"] = bson.M{"$regex": params.Brand, "$options": "i"}
	}
	if params.MinPrice > 0 {
		filter["price"] = bson.M{"$gte": params.MinPrice}
	}
	if params.MaxPrice > 0 {
		if _, ok := filter["price"]; ok {
			filter["price"].(bson.M)["$lte"] = params.MaxPrice
		} else {
			filter["price"] = bson.M{"$lte": params.MaxPrice}
		}
	}

	return r.listWithCursor(ctx, filter, cursor, limit, "published_at")
}

func (r *productMongoRepository) ListByHashtag(ctx context.Context, tag string, cursor string, limit int) ([]domain.Product, string, error) {
	filter := bson.M{"status": domain.StatusActive, "hashtags": tag, "deleted_at": nil}
	return r.listWithCursor(ctx, filter, cursor, limit, "published_at")
}

func (r *productMongoRepository) listWithCursor(ctx context.Context, filter bson.M, cursor string, limit int, sortField string) ([]domain.Product, string, error) {
	if cursor != "" {
		cursorFilter, err := pagination.CursorFilter(cursor, sortField)
		if err == nil && len(cursorFilter) > 0 {
			for _, elem := range cursorFilter {
				filter[elem.Key] = elem.Value
			}
		}
	}

	opts := options.Find().
		SetLimit(int64(limit + 1)).
		SetSort(bson.D{{Key: sortField, Value: -1}, {Key: "_id", Value: -1}})

	cur, err := r.products.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}

	var products []domain.Product
	if err := cur.All(ctx, &products); err != nil {
		return nil, "", err
	}

	var nextCursor string
	hasMore := len(products) > limit
	if hasMore {
		products = products[:limit]
		last := products[len(products)-1]
		t := last.CreatedAt
		if sortField == "published_at" && last.PublishedAt != nil {
			t = *last.PublishedAt
		}
		nextCursor = pagination.EncodeCursor(last.ID.Hex(), t)
	}

	return products, nextCursor, nil
}
