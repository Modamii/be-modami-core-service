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

type Hashtag struct {
	Tag        string    `bson:"_id" json:"tag"`
	UsageCount int64     `bson:"usage_count" json:"usage_count"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}

type CategoryOrder struct {
	ID        string `json:"id"`
	SortOrder int    `json:"sort_order"`
}
