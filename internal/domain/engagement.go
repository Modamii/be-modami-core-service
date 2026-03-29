package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Favorite struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"user_id" json:"user_id"`
	ProductID bson.ObjectID `bson:"product_id" json:"product_id"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}

type SavedProduct struct {
	ID           bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID       bson.ObjectID  `bson:"user_id" json:"user_id"`
	ProductID    bson.ObjectID  `bson:"product_id" json:"product_id"`
	CollectionID *bson.ObjectID `bson:"collection_id,omitempty" json:"collection_id,omitempty"`
	Note         string         `bson:"note,omitempty" json:"note,omitempty"`
	CreatedAt    time.Time      `bson:"created_at" json:"created_at"`
}

type SavedCollection struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"user_id" json:"user_id"`
	Name      string        `bson:"name" json:"name"`
	ItemCount int           `bson:"item_count" json:"item_count"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
}

type Follow struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	FollowerID bson.ObjectID `bson:"follower_id" json:"follower_id"`
	SellerID   bson.ObjectID `bson:"seller_id" json:"seller_id"`
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`
}

type Review struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OrderID   bson.ObjectID `bson:"order_id" json:"order_id"`
	ProductID bson.ObjectID `bson:"product_id" json:"product_id"`
	BuyerID   bson.ObjectID `bson:"buyer_id" json:"buyer_id"`
	SellerID  bson.ObjectID `bson:"seller_id" json:"seller_id"`
	Rating    int           `bson:"rating" json:"rating"`
	Comment   string        `bson:"comment,omitempty" json:"comment,omitempty"`
	Images    []string      `bson:"images,omitempty" json:"images,omitempty"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
}
