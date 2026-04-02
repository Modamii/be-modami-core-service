package events

import (
	"time"

	kafkaevents "github.com/modami/core-service/pkg/kafka/events"
)

const EventTypeProductCreated = "modami.product.created"

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
