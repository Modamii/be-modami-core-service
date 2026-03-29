package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type DailyStat struct {
	ID   bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Date string        `bson:"date" json:"date"`

	NewProducts         int `bson:"new_products" json:"new_products"`
	ApprovedProducts    int `bson:"approved_products" json:"approved_products"`
	RejectedProducts    int `bson:"rejected_products" json:"rejected_products"`
	SoldProducts        int `bson:"sold_products" json:"sold_products"`
	TotalActiveProducts int `bson:"total_active_products" json:"total_active_products"`

	NewOrders       int   `bson:"new_orders" json:"new_orders"`
	CompletedOrders int   `bson:"completed_orders" json:"completed_orders"`
	CancelledOrders int   `bson:"cancelled_orders" json:"cancelled_orders"`
	TotalGMV        int64 `bson:"total_gmv" json:"total_gmv"`

	NewUsers         int `bson:"new_users" json:"new_users"`
	ActiveUsers      int `bson:"active_users" json:"active_users"`
	NewSubscriptions int `bson:"new_subscriptions" json:"new_subscriptions"`

	CreditsPurchased int `bson:"credits_purchased" json:"credits_purchased"`
	CreditsSpent     int `bson:"credits_spent" json:"credits_spent"`
	UnlockCount      int `bson:"unlock_count" json:"unlock_count"`

	SubscriptionRevenue int64 `bson:"subscription_revenue" json:"subscription_revenue"`
	PlatformFeeRevenue  int64 `bson:"platform_fee_revenue" json:"platform_fee_revenue"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type SellerStatsSnapshot struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id"`
	SellerID        bson.ObjectID `bson:"seller_id" json:"seller_id"`
	Period          string        `bson:"period" json:"period"`
	ProductsListed  int           `bson:"products_listed" json:"products_listed"`
	ProductsSold    int           `bson:"products_sold" json:"products_sold"`
	TotalRevenue    int64         `bson:"total_revenue" json:"total_revenue"`
	OrdersCompleted int           `bson:"orders_completed" json:"orders_completed"`
	OrdersCancelled int           `bson:"orders_cancelled" json:"orders_cancelled"`
	AvgRating       float64       `bson:"avg_rating" json:"avg_rating"`
	NewReviews      int           `bson:"new_reviews" json:"new_reviews"`
	NewFollowers    int           `bson:"new_followers" json:"new_followers"`
	ProfileViews    int           `bson:"profile_views" json:"profile_views"`
	UnlocksReceived int           `bson:"unlocks_received" json:"unlocks_received"`
	CreatedAt       time.Time     `bson:"created_at" json:"created_at"`
}
