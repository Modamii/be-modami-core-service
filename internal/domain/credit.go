package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CreditWallet struct {
	UserID      bson.ObjectID `bson:"_id" json:"user_id"`
	Balance     int           `bson:"balance" json:"balance"`
	TotalEarned int           `bson:"total_earned" json:"total_earned"`
	TotalSpent  int           `bson:"total_spent" json:"total_spent"`
	Version     int64         `bson:"version" json:"version"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updated_at"`
}

type CreditTransaction struct {
	ID           bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID       bson.ObjectID  `bson:"user_id" json:"user_id"`
	Amount       int            `bson:"amount" json:"amount"`
	Type         string         `bson:"type" json:"type"`
	RefType      string         `bson:"ref_type,omitempty" json:"ref_type,omitempty"`
	RefID        *bson.ObjectID `bson:"ref_id,omitempty" json:"ref_id,omitempty"`
	BalanceAfter int            `bson:"balance_after" json:"balance_after"`
	Description  string         `bson:"description" json:"description"`
	CreatedAt    time.Time      `bson:"created_at" json:"created_at"`
}

type ContactUnlock struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	BuyerID    bson.ObjectID `bson:"buyer_id" json:"buyer_id"`
	ProductID  bson.ObjectID `bson:"product_id" json:"product_id"`
	SellerID   bson.ObjectID `bson:"seller_id" json:"seller_id"`
	CreditTxID bson.ObjectID `bson:"credit_tx_id" json:"credit_tx_id"`
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`
}
