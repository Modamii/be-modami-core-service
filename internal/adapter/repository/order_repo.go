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

type orderMongoRepository struct {
	orders *mongo.Collection
	events *mongo.Collection
}

// NewOrderRepository returns a MongoDB-backed order repository.
func NewOrderRepository(db *mongo.Database) port.OrderRepository {
	return &orderMongoRepository{
		orders: db.Collection("orders"),
		events: db.Collection("order_events"),
	}
}

func (r *orderMongoRepository) Create(ctx context.Context, o *domain.Order) error {
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	o.Version = 1
	if o.Status == "" {
		o.Status = domain.StatusCreated
	}
	result, err := r.orders.InsertOne(ctx, o)
	if err != nil {
		return err
	}
	o.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *orderMongoRepository) GetByID(ctx context.Context, id bson.ObjectID) (*domain.Order, error) {
	var o domain.Order
	err := r.orders.FindOne(ctx, bson.M{"_id": id}).Decode(&o)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *orderMongoRepository) GetByOrderCode(ctx context.Context, code string) (*domain.Order, error) {
	var o domain.Order
	err := r.orders.FindOne(ctx, bson.M{"order_code": code}).Decode(&o)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *orderMongoRepository) Update(ctx context.Context, o *domain.Order) error {
	o.UpdatedAt = time.Now()
	filter := bson.M{"_id": o.ID, "version": o.Version}
	o.Version++
	result, err := r.orders.ReplaceOne(ctx, filter, o)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrOrderVersionConflict
	}
	return nil
}

func (r *orderMongoRepository) ListByBuyer(ctx context.Context, buyerID bson.ObjectID, cursor string, limit int) ([]domain.Order, string, error) {
	filter := bson.M{"buyer_id": buyerID}
	return r.listWithCursor(ctx, filter, cursor, limit)
}

func (r *orderMongoRepository) ListBySeller(ctx context.Context, sellerID bson.ObjectID, cursor string, limit int) ([]domain.Order, string, error) {
	filter := bson.M{"seller_id": sellerID}
	return r.listWithCursor(ctx, filter, cursor, limit)
}

func (r *orderMongoRepository) ListAll(ctx context.Context, status string, cursor string, limit int) ([]domain.Order, string, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}
	return r.listWithCursor(ctx, filter, cursor, limit)
}

func (r *orderMongoRepository) CreateEvent(ctx context.Context, e *domain.OrderEvent) error {
	e.CreatedAt = time.Now()
	result, err := r.events.InsertOne(ctx, e)
	if err != nil {
		return err
	}
	e.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

func (r *orderMongoRepository) ListEvents(ctx context.Context, orderID bson.ObjectID) ([]domain.OrderEvent, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.events.Find(ctx, bson.M{"order_id": orderID}, opts)
	if err != nil {
		return nil, err
	}
	var events []domain.OrderEvent
	if err := cur.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *orderMongoRepository) listWithCursor(ctx context.Context, filter bson.M, cursor string, limit int) ([]domain.Order, string, error) {
	if cursor != "" {
		cursorFilter, err := pagination.CursorFilter(cursor, "created_at")
		if err == nil && len(cursorFilter) > 0 {
			for _, elem := range cursorFilter {
				filter[elem.Key] = elem.Value
			}
		}
	}

	opts := options.Find().
		SetLimit(int64(limit + 1)).
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}})

	cur, err := r.orders.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}

	var orders []domain.Order
	if err := cur.All(ctx, &orders); err != nil {
		return nil, "", err
	}

	var nextCursor string
	hasMore := len(orders) > limit
	if hasMore {
		orders = orders[:limit]
		last := orders[len(orders)-1]
		nextCursor = pagination.EncodeCursor(last.ID.Hex(), last.CreatedAt)
	}

	return orders, nextCursor, nil
}
