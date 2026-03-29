package port

import (
	"context"

	"github.com/modami/core-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// OrderRepository persists orders and order events.
type OrderRepository interface {
	Create(ctx context.Context, o *domain.Order) error
	GetByID(ctx context.Context, id bson.ObjectID) (*domain.Order, error)
	GetByOrderCode(ctx context.Context, code string) (*domain.Order, error)
	Update(ctx context.Context, o *domain.Order) error
	ListByBuyer(ctx context.Context, buyerID bson.ObjectID, cursor string, limit int) ([]domain.Order, string, error)
	ListBySeller(ctx context.Context, sellerID bson.ObjectID, cursor string, limit int) ([]domain.Order, string, error)
	ListAll(ctx context.Context, status string, cursor string, limit int) ([]domain.Order, string, error)
	CreateEvent(ctx context.Context, e *domain.OrderEvent) error
	ListEvents(ctx context.Context, orderID bson.ObjectID) ([]domain.OrderEvent, error)
}

// OrderProductReader loads product data needed to create an order snapshot.
type OrderProductReader interface {
	GetByIDForOrder(ctx context.Context, id string) (OrderProductSnapshot, error)
}

// OrderProductSnapshot is denormalized product data for order creation.
type OrderProductSnapshot struct {
	ID        bson.ObjectID
	SellerID  bson.ObjectID
	Title     string
	ImageURL  string
	Brand     string
	Condition string
	Size      string
	Category  string
	Price     int64
}
