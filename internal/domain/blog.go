package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
)

type PostDepth string

const (
	PostDepthQuick PostDepth = "quick"
	PostDepthDeep  PostDepth = "deep"
)

// BlogAuthor holds the embedded author metadata for a blog post.
type BlogAuthor struct {
	Name  string `bson:"name"  json:"name"`
	Title string `bson:"title" json:"title"`
	Bio   string `bson:"bio"   json:"bio"`
}

// BlogPost is the core domain model for the Community & Blog feature.
type BlogPost struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"id"`

	Slug string `bson:"slug" json:"slug"`

	// Series metadata
	SeriesName    string `bson:"series_name"    json:"series_name"`    // e.g. "MODAMI INSIGHT"
	SeriesNo      int    `bson:"series_no"      json:"series_no"`      // e.g. 12
	SeriesQuarter string `bson:"series_quarter" json:"series_quarter"` // e.g. "Q4/2025"

	// Categorisation
	PostType string    `bson:"post_type" json:"post_type"` // e.g. "XU HƯỚNG TIÊU ĐIỂM"
	Depth    PostDepth `bson:"depth"     json:"depth"`     // "quick" | "deep"

	// Content
	Title        string `bson:"title"                   json:"title"`
	Subtitle     string `bson:"subtitle"                json:"subtitle"`
	Body         string `bson:"body"                    json:"body"`
	CoverImage   string `bson:"cover_image"             json:"cover_image"`
	CoverCaption string `bson:"cover_caption,omitempty" json:"cover_caption,omitempty"`

	ReadTimeMin int `bson:"read_time_min" json:"read_time_min"`
	WordCount   int `bson:"word_count"    json:"word_count"`

	Author     BlogAuthor `bson:"author"              json:"author"`
	KeyPoints  []string   `bson:"key_points,omitempty"  json:"key_points,omitempty"`
	References []string   `bson:"references,omitempty"  json:"references,omitempty"`
	Hashtags   []string   `bson:"hashtags,omitempty"    json:"hashtags,omitempty"`
	CTALink    string     `bson:"cta_link,omitempty"    json:"cta_link,omitempty"`

	IsFeatured bool       `bson:"is_featured" json:"is_featured"`
	Status     PostStatus `bson:"status"      json:"status"`

	PublishedAt *time.Time `bson:"published_at,omitempty" json:"published_at,omitempty"`
	UpdatedAt   time.Time  `bson:"updated_at"             json:"updated_at"`
	CreatedAt   time.Time  `bson:"created_at"             json:"created_at"`
}
