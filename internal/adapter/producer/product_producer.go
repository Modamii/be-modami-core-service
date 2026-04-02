package producer

import (
	"context"
	"time"

	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"

	"be-modami-core-service/internal/domain"
	internalevents "be-modami-core-service/internal/events"
	"be-modami-core-service/pkg/kafka"
	kafkaevents "be-modami-core-service/pkg/kafka/events"
)

type ProductProducer struct {
	producer kafka.Producer
}

func NewProductProducer(producer kafka.Producer) *ProductProducer {
	return &ProductProducer{producer: producer}
}

func (p *ProductProducer) ProductCreatedWithData(ctx context.Context, product *domain.Product) error {
	catID, catName := categoryFields(product)
	payload := &internalevents.ProductCreatedPayload{
		BaseEventPayload: kafkaevents.BaseEventPayload{
			Type:      internalevents.EventTypeProductCreated,
			Timestamp: time.Now(),
		},
		ProductID:    product.ID.Hex(),
		Slug:         product.Slug,
		Title:        product.Title,
		SellerID:     product.SellerID.Hex(),
		CategoryID:   catID,
		CategoryName: catName,
		Status:       string(product.Status),
		Price:        product.Price,
		Brand:        product.Brand,
		Condition:    product.Condition,
		Hashtags:     product.Hashtags,
		Images:       imageURLs(product),
		CreatedAt:    product.CreatedAt,
	}
	return p.emitAsync(ctx, kafka.TopicProductCreated, product.ID.Hex(), payload, "product created")
}

func (p *ProductProducer) ProductUpdatedWithData(ctx context.Context, product *domain.Product) error {
	catID, catName := categoryFields(product)
	payload := &internalevents.ProductUpdatedPayload{
		BaseEventPayload: kafkaevents.BaseEventPayload{
			Type:      internalevents.EventTypeProductUpdated,
			Timestamp: time.Now(),
		},
		ProductID:    product.ID.Hex(),
		Slug:         product.Slug,
		Title:        product.Title,
		SellerID:     product.SellerID.Hex(),
		CategoryID:   catID,
		CategoryName: catName,
		Status:       string(product.Status),
		Price:        product.Price,
		Brand:        product.Brand,
		Condition:    product.Condition,
		Hashtags:     product.Hashtags,
		Images:       imageURLs(product),
		UpdatedAt:    product.UpdatedAt,
	}
	return p.emitAsync(ctx, kafka.TopicProductUpdated, product.ID.Hex(), payload, "product updated")
}

func (p *ProductProducer) ProductDeleted(ctx context.Context, productID, slug string) error {
	payload := &internalevents.ProductDeletedPayload{
		BaseEventPayload: kafkaevents.BaseEventPayload{
			Type:      internalevents.EventTypeProductDeleted,
			Timestamp: time.Now(),
		},
		ProductID: productID,
		Slug:      slug,
	}
	return p.emitAsync(ctx, kafka.TopicProductDeleted, productID, payload, "product deleted")
}

func (p *ProductProducer) emitAsync(ctx context.Context, topic, key string, payload interface{}, eventName string) error {
	p.producer.EmitAsync(ctx, topic, &kafka.ProducerMessage{
		Key:   key,
		Value: payload,
	})
	logger.Debug(ctx, "published "+eventName+" event", logging.String("productId", key))
	return nil
}

func imageURLs(product *domain.Product) []string {
	urls := make([]string, 0, len(product.Images))
	for _, img := range product.Images {
		urls = append(urls, img.URL)
	}
	return urls
}

func categoryFields(product *domain.Product) (id, name string) {
	if product.Category != nil {
		id = product.Category.ID.Hex()
		name = product.Category.Name
	}
	return
}
