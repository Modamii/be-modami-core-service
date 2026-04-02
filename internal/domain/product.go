package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductStatus string

const (
	StatusDraft    ProductStatus = "draft"
	StatusPending  ProductStatus = "pending"
	StatusActive   ProductStatus = "active"
	StatusSold     ProductStatus = "sold"
	StatusArchived ProductStatus = "archived"
)

type Product struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id"`
	SellerID bson.ObjectID `bson:"seller_id" json:"seller_id"`
	Status   ProductStatus `bson:"status" json:"status"`
	Version  int64         `bson:"version" json:"version"`

	Title       string `bson:"title" json:"title"`
	Slug        string `bson:"slug" json:"slug"`
	Description string `bson:"description" json:"description"`
	Price       int64  `bson:"price" json:"price"`

	Category *Category `bson:"category" json:"category"`
	Condition  string        `bson:"condition" json:"condition"`
	Size       string        `bson:"size" json:"size"`
	Brand      string        `bson:"brand,omitempty" json:"brand,omitempty"`
	Color      string        `bson:"color,omitempty" json:"color,omitempty"`
	Material   string        `bson:"material,omitempty" json:"material,omitempty"`

	Images []ProductImage `bson:"images" json:"images"`

	IsVerified bool `bson:"is_verified" json:"is_verified"`
	IsFeatured bool `bson:"is_featured" json:"is_featured"`
	IsSelect   bool `bson:"is_select" json:"is_select"`

	CreditCost int `bson:"credit_cost" json:"credit_cost"`

	Hashtags []string `bson:"hashtags,omitempty" json:"hashtags,omitempty"`

	CreatedAt   time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at" json:"updated_at"`
	PublishedAt *time.Time `bson:"published_at,omitempty" json:"published_at,omitempty"`
	SoldAt      *time.Time `bson:"sold_at,omitempty" json:"sold_at,omitempty"`
	DeletedAt   *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type ProductImage struct {
	URL      string `bson:"url" json:"url"`
	Position int    `bson:"position" json:"position"`
	Width    int    `bson:"width,omitempty" json:"width,omitempty"`
	Height   int    `bson:"height,omitempty" json:"height,omitempty"`
}

type ProductStats struct {
	ProductID     bson.ObjectID `bson:"_id" json:"product_id"`
	ViewCount     int64         `bson:"view_count" json:"view_count"`
	LikeCount     int64         `bson:"like_count" json:"like_count"`
	CommentCount  int64         `bson:"comment_count" json:"comment_count"`
	FavoriteCount int64         `bson:"favorite_count" json:"favorite_count"`
	SaveCount     int64         `bson:"save_count" json:"save_count"`
	UnlockCount   int64         `bson:"unlock_count" json:"unlock_count"`
	ShareCount    int64         `bson:"share_count" json:"share_count"`
	UpdatedAt     time.Time     `bson:"updated_at" json:"updated_at"`
}

// ProductStatsSummary is the public-facing API stats shape.
type ProductStatsSummary struct {
	TotalView    int64 `json:"totalView"`
	TotalLike    int64 `json:"totalLike"`
	TotalComment int64 `json:"totalComment"`
}

// ProductDetail holds a product with its stats, used for detail API responses.
type ProductDetail struct {
	Product *Product            `json:"product"`
	Stats   ProductStatsSummary `json:"stats"`
}

type ProductModeration struct {
	ID          bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	ProductID   bson.ObjectID  `bson:"product_id" json:"product_id"`
	Round       int            `bson:"round" json:"round"`
	Action      string         `bson:"action" json:"action"`
	RejectCode  string         `bson:"reject_code,omitempty" json:"reject_code,omitempty"`
	Reason      string         `bson:"reason,omitempty" json:"reason,omitempty"`
	Note        string         `bson:"note,omitempty" json:"note,omitempty"`
	Suggestion  string         `bson:"suggestion,omitempty" json:"suggestion,omitempty"`
	ModeratorID *bson.ObjectID `bson:"moderator_id,omitempty" json:"moderator_id,omitempty"`
	CreatedAt   time.Time      `bson:"created_at" json:"created_at"`
}

type SelectProduct struct {
	ProductID     bson.ObjectID `bson:"_id" json:"product_id"`
	Campaign      string        `bson:"campaign" json:"campaign"`
	Story         string        `bson:"story" json:"story"`
	Provenance    string        `bson:"provenance" json:"provenance"`
	Year          int           `bson:"year,omitempty" json:"year,omitempty"`
	CertificateID string        `bson:"certificate_id" json:"certificate_id"`
	VerifiedBy    bson.ObjectID `bson:"verified_by" json:"verified_by"`
	VerifiedAt    time.Time     `bson:"verified_at" json:"verified_at"`
	CreatedAt     time.Time     `bson:"created_at" json:"created_at"`
}

// ProductListParams holds filter/sort options for listing and search.
type ProductListParams struct {
	Status     string
	CategoryID string
	Condition  string
	MinPrice   int64
	MaxPrice   int64
	Brand      string
	Sort       string
}

// ValidProductTransition checks if a product status transition is allowed.
func ValidProductTransition(from, to ProductStatus) bool {
	transitions := map[ProductStatus][]ProductStatus{
		StatusDraft:    {StatusPending},
		StatusPending:  {StatusActive, StatusDraft},
		StatusActive:   {StatusSold, StatusArchived},
		StatusArchived: {StatusActive},
	}
	for _, allowed := range transitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}
