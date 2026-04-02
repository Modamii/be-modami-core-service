package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"be-modami-core-service/internal/domain"
	"be-modami-core-service/internal/events"
	"be-modami-core-service/internal/port"
	"be-modami-core-service/pkg/elasticsearch"
	"be-modami-core-service/pkg/kafka"
)

type SyncProductConsumer struct {
	esClient    *elasticsearch.Client
	productRepo port.ProductRepository
}

func NewSyncProductConsumer(esClient *elasticsearch.Client, productRepo port.ProductRepository) *SyncProductConsumer {
	return &SyncProductConsumer{esClient: esClient, productRepo: productRepo}
}

func (c *SyncProductConsumer) GetTopics() []string {
	return []string{
		kafka.TopicProductCreated,
		kafka.TopicProductUpdated,
		kafka.TopicProductDeleted,
	}
}

func (c *SyncProductConsumer) HandleMessage(ctx context.Context, record *kgo.Record) error {
	switch record.Topic {
	case kafka.TopicProductCreated, kafka.TopicProductUpdated:
		return c.handleUpsert(ctx, record)
	case kafka.TopicProductDeleted:
		return c.handleDelete(ctx, record)
	default:
		return nil
	}
}

func (c *SyncProductConsumer) handleUpsert(ctx context.Context, record *kgo.Record) error {
	// Both created and updated events share the productId field at top level
	var base struct {
		ProductID string `json:"productId"`
	}
	if err := json.Unmarshal(record.Value, &base); err != nil {
		return fmt.Errorf("unmarshal product event: %w", err)
	}

	oid, err := bson.ObjectIDFromHex(base.ProductID)
	if err != nil {
		return fmt.Errorf("invalid product id: %w", err)
	}

	product, err := c.productRepo.GetByID(ctx, oid)
	if err != nil || product == nil {
		logger.Warn(ctx, "sync: product not found", logging.String("id", base.ProductID))
		return nil
	}

	// Only index active products
	if product.Status != domain.StatusActive {
		return nil
	}

	if err := c.esClient.IndexProduct(ctx, buildProductDocument(product)); err != nil {
		logger.Error(ctx, "sync: es index failed", err, logging.String("id", product.ID.Hex()))
		return err
	}

	logger.Info(ctx, "sync: product indexed", logging.String("id", product.ID.Hex()))
	return nil
}

func (c *SyncProductConsumer) handleDelete(ctx context.Context, record *kgo.Record) error {
	var payload events.ProductDeletedPayload
	if err := json.Unmarshal(record.Value, &payload); err != nil {
		return fmt.Errorf("unmarshal product deleted event: %w", err)
	}

	if err := c.esClient.DeleteProduct(ctx, payload.ProductID); err != nil {
		logger.Error(ctx, "sync: es delete failed", err, logging.String("id", payload.ProductID))
		return err
	}

	logger.Info(ctx, "sync: product deleted from index", logging.String("id", payload.ProductID))
	return nil
}

func buildProductDocument(product *domain.Product) *elasticsearch.ProductDocument {
	images := make([]string, 0, len(product.Images))
	for _, img := range product.Images {
		images = append(images, img.URL)
	}

	var catID, catName string
	if product.Category != nil {
		catID = product.Category.ID.Hex()
		catName = product.Category.Name
	}

	return &elasticsearch.ProductDocument{
		ID:           product.ID.Hex(),
		Slug:         product.Slug,
		Title:        product.Title,
		Description:  product.Description,
		Price:        product.Price,
		Brand:        product.Brand,
		Condition:    product.Condition,
		CategoryID:   catID,
		CategoryName: catName,
		Status:       string(product.Status),
		SellerID:     product.SellerID.Hex(),
		Images:       images,
		Hashtags:     product.Hashtags,
		IsVerified:   product.IsVerified,
		IsFeatured:   product.IsFeatured,
		IsSelect:     product.IsSelect,
		PublishedAt:  product.PublishedAt,
		CreatedAt:    product.CreatedAt,
	}
}
