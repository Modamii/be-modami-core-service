package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"be-modami-core-service/internal/domain"
	"be-modami-core-service/internal/events"
	"be-modami-core-service/internal/port"
	"be-modami-core-service/pkg/elasticsearch"
	localkafka "be-modami-core-service/pkg/kafka"
	pkges "gitlab.com/lifegoeson-libs/pkg-gokit/elasticsearch"
	pkgkafka "gitlab.com/lifegoeson-libs/pkg-gokit/kafka"
	pkgredis "gitlab.com/lifegoeson-libs/pkg-gokit/redis"
)

type SyncProductConsumer struct {
	esClient    *pkges.Client
	cacheClient pkgredis.CachePort
	productRepo port.ProductRepository
}

func NewSyncProductConsumer(esClient *pkges.Client, productRepo port.ProductRepository) *SyncProductConsumer {
	return &SyncProductConsumer{esClient: esClient, productRepo: productRepo}
}

func (c *SyncProductConsumer) GetTopics() []string {
	return []string{
		localkafka.TopicProductCreated,
		localkafka.TopicProductUpdated,
		localkafka.TopicProductDeleted,
	}
}

func (c *SyncProductConsumer) HandleMessage(ctx context.Context, msg *pkgkafka.Message) error {
	switch msg.Topic {
	case localkafka.TopicProductCreated, localkafka.TopicProductUpdated:
		c.handleESUpsert(ctx, msg)
		c.handlerCacheDelete(ctx, msg)
		return nil
	case localkafka.TopicProductDeleted:
		c.handleESDelete(ctx, msg)
		c.handlerCacheDelete(ctx, msg)
		return nil
	default:
		return nil
	}
}

func (c *SyncProductConsumer) handleESUpsert(ctx context.Context, msg *pkgkafka.Message) error {
	var base struct {
		ProductID string `json:"productId"`
	}
	if err := json.Unmarshal(msg.Value, &base); err != nil {
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

	if product.Status != domain.StatusActive {
		return nil
	}

	if err := elasticsearch.IndexProduct(ctx, c.esClient, buildProductDocument(product)); err != nil {
		logger.Error(ctx, "sync: es index failed", err, logging.String("id", product.ID.Hex()))
		return err
	}

	logger.Info(ctx, "sync: product indexed", logging.String("id", product.ID.Hex()))
	return nil
}

func (c *SyncProductConsumer) handlerCacheDelete(ctx context.Context, msg *pkgkafka.Message) error {
	if c.cacheClient == nil {
		return nil
	}
	var payload events.ProductDeletedPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return fmt.Errorf("unmarshal product deleted event: %w", err)
	}

	if err := c.cacheClient.Delete(ctx, payload.ProductID); err != nil {
		logger.Error(ctx, "sync: cache delete failed", err, logging.String("id", payload.ProductID))
		return err
	}

	logger.Info(ctx, "sync: product deleted from cache", logging.String("id", payload.ProductID))
	return nil
}

func (c *SyncProductConsumer) handleESDelete(ctx context.Context, msg *pkgkafka.Message) error {
	var payload events.ProductDeletedPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return fmt.Errorf("unmarshal product deleted event: %w", err)
	}

	if err := elasticsearch.DeleteProduct(ctx, c.esClient, payload.ProductID); err != nil {
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
		SellerID:     product.SellerID,
		Images:       images,
		Hashtags:     product.Hashtags,
		IsVerified:   product.IsVerified,
		IsFeatured:   product.IsFeatured,
		IsSelect:     product.IsSelect,
		PublishedAt:  product.PublishedAt,
		CreatedAt:    product.CreatedAt,
	}
}
