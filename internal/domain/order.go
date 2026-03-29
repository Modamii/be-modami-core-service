package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderStatus string

const (
	StatusCreated   OrderStatus = "created"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
	StatusCompleted OrderStatus = "completed"
)

type Order struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OrderCode string        `bson:"order_code" json:"order_code"`
	Version   int64         `bson:"version" json:"version"`

	BuyerID   bson.ObjectID `bson:"buyer_id" json:"buyer_id"`
	SellerID  bson.ObjectID `bson:"seller_id" json:"seller_id"`
	ProductID bson.ObjectID `bson:"product_id" json:"product_id"`

	Snapshot OrderSnapshot `bson:"snapshot" json:"snapshot"`

	ItemPrice   int64 `bson:"item_price" json:"item_price"`
	ShippingFee int64 `bson:"shipping_fee" json:"shipping_fee"`
	PlatformFee int64 `bson:"platform_fee" json:"platform_fee"`
	TotalPrice  int64 `bson:"total_price" json:"total_price"`

	Shipping         ShippingInfo `bson:"shipping" json:"shipping"`
	TrackingCode     string       `bson:"tracking_code,omitempty" json:"tracking_code,omitempty"`
	ShippingProvider string       `bson:"shipping_provider,omitempty" json:"shipping_provider,omitempty"`

	Status       OrderStatus `bson:"status" json:"status"`
	CancelReason string      `bson:"cancel_reason,omitempty" json:"cancel_reason,omitempty"`
	CancelledBy  string      `bson:"cancelled_by,omitempty" json:"cancelled_by,omitempty"`

	CreatedAt   time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at" json:"updated_at"`
	ConfirmedAt *time.Time `bson:"confirmed_at,omitempty" json:"confirmed_at,omitempty"`
	ShippedAt   *time.Time `bson:"shipped_at,omitempty" json:"shipped_at,omitempty"`
	DeliveredAt *time.Time `bson:"delivered_at,omitempty" json:"delivered_at,omitempty"`
	CancelledAt *time.Time `bson:"cancelled_at,omitempty" json:"cancelled_at,omitempty"`
}

type OrderSnapshot struct {
	Title     string `bson:"title" json:"title"`
	ImageURL  string `bson:"image_url" json:"image_url"`
	Brand     string `bson:"brand" json:"brand"`
	Condition string `bson:"condition" json:"condition"`
	Size      string `bson:"size" json:"size"`
	Category  string `bson:"category" json:"category"`
}

type ShippingInfo struct {
	Name     string `bson:"name" json:"name"`
	Phone    string `bson:"phone" json:"phone"`
	Address  string `bson:"address" json:"address"`
	Province string `bson:"province" json:"province"`
	District string `bson:"district" json:"district"`
	Ward     string `bson:"ward" json:"ward"`
}

type OrderEvent struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OrderID    bson.ObjectID `bson:"order_id" json:"order_id"`
	FromStatus string        `bson:"from_status" json:"from_status"`
	ToStatus   string        `bson:"to_status" json:"to_status"`
	ActorID    bson.ObjectID `bson:"actor_id" json:"actor_id"`
	ActorType  string        `bson:"actor_type" json:"actor_type"`
	Note       string        `bson:"note,omitempty" json:"note,omitempty"`
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`
}

func ValidOrderTransition(from, to OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		StatusCreated:   {StatusConfirmed, StatusCancelled},
		StatusConfirmed: {StatusShipped, StatusCancelled},
		StatusShipped:   {StatusDelivered},
		StatusDelivered: {StatusCompleted},
	}
	for _, allowed := range transitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}
