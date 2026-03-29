package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Subscription struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"user_id" json:"user_id"`
	PackageID bson.ObjectID `bson:"package_id" json:"package_id"`

	BillingCycle     string `bson:"billing_cycle" json:"billing_cycle"`
	PricePaid        int64  `bson:"price_paid" json:"price_paid"`
	CreditsAllocated int    `bson:"credits_allocated" json:"credits_allocated"`
	CreditsUsed      int    `bson:"credits_used" json:"credits_used"`

	Status       string     `bson:"status" json:"status"`
	AutoRenew    bool       `bson:"auto_renew" json:"auto_renew"`
	StartDate    time.Time  `bson:"start_date" json:"start_date"`
	EndDate      time.Time  `bson:"end_date" json:"end_date"`
	CancelledAt  *time.Time `bson:"cancelled_at,omitempty" json:"cancelled_at,omitempty"`
	CancelReason string     `bson:"cancel_reason,omitempty" json:"cancel_reason,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
