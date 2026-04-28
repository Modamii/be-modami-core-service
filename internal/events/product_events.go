package events

import (
	"time"

	kafkaevents "gitlab.com/lifegoeson-libs/pkg-gokit/kafka/events"
)

const (
	EventTypeProductCreated = "modami.product.created"
	EventTypeProductUpdated = "modami.product.updated"
	EventTypeProductDeleted = "modami.product.deleted"
)

type ProductCreatedPayload struct {
	kafkaevents.BaseEventPayload
	ProductID    string    `json:"productId"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	SellerID     string    `json:"sellerId"`
	CategoryID   string    `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Status       string    `json:"status"`
	Price        int64     `json:"price"`
	Brand        string    `json:"brand,omitempty"`
	Condition    string    `json:"condition"`
	Hashtags     []string  `json:"hashtags,omitempty"`
	Images       []string  `json:"images,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (p *ProductCreatedPayload) GetType() string { return EventTypeProductCreated }
func (p *ProductCreatedPayload) Validate() error { return nil }

type ProductUpdatedPayload struct {
	kafkaevents.BaseEventPayload
	ProductID    string    `json:"productId"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	SellerID     string    `json:"sellerId"`
	CategoryID   string    `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Status       string    `json:"status"`
	Price        int64     `json:"price"`
	Brand        string    `json:"brand,omitempty"`
	Condition    string    `json:"condition"`
	Hashtags     []string  `json:"hashtags,omitempty"`
	Images       []string  `json:"images,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (p *ProductUpdatedPayload) GetType() string { return EventTypeProductUpdated }
func (p *ProductUpdatedPayload) Validate() error { return nil }

type ProductDeletedPayload struct {
	kafkaevents.BaseEventPayload
	ProductID string `json:"productId"`
	Slug      string `json:"slug"`
}

func (p *ProductDeletedPayload) GetType() string { return EventTypeProductDeleted }
func (p *ProductDeletedPayload) Validate() error { return nil }
