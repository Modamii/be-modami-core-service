package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Category struct {
	ID           bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	Name         string         `bson:"name" json:"name"`
	NameVI       string         `bson:"name_vi" json:"name_vi"`
	Slug         string         `bson:"slug" json:"slug"`
	Icon         string         `bson:"icon,omitempty" json:"icon,omitempty"`
	ParentID     *bson.ObjectID `bson:"parent_id,omitempty" json:"parent_id,omitempty"`
	SortOrder    int            `bson:"sort_order" json:"sort_order"`
	IsActive     bool           `bson:"is_active" json:"is_active"`
	ProductCount int64          `bson:"product_count" json:"product_count"`
	CreatedAt    time.Time      `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `bson:"updated_at" json:"updated_at"`
}

// Package is a sellable membership tier (Style / Elite), stored in collection `packages`.
type Package struct {
	ID   bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Code string        `bson:"code" json:"code"`
	Name string        `bson:"name" json:"name"`
	Tier int           `bson:"tier" json:"tier"`

	PriceMonthly int64  `bson:"price_monthly" json:"price_monthly"`
	PriceYearly  int64  `bson:"price_yearly" json:"price_yearly"`
	Currency     string `bson:"currency" json:"currency"`

	CreditsPerMonth int    `bson:"credits_per_month" json:"credits_per_month"`
	SearchBoost     bool   `bson:"search_boost" json:"search_boost"`
	SearchPriority  bool   `bson:"search_priority" json:"search_priority"`
	BadgeName       string `bson:"badge_name" json:"badge_name"`
	PrioritySupport bool   `bson:"priority_support" json:"priority_support"`
	FeaturedSlots   int    `bson:"featured_slots" json:"featured_slots"`

	IsActive  bool      `bson:"is_active" json:"is_active"`
	SortOrder int       `bson:"sort_order" json:"sort_order"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type Hashtag struct {
	Tag        string    `bson:"_id" json:"tag"`
	UsageCount int64     `bson:"usage_count" json:"usage_count"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}

// CategoryOrder is the request payload for admin category reorder.
type CategoryOrder struct {
	ID        string `json:"id"`
	SortOrder int    `json:"sort_order"`
}
