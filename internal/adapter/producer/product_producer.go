package producer

import (
	"context"
	"fmt"
	"techinsight-api/internal/entity"
	"techinsight-api/pkg/kafka"
	"techinsight-api/pkg/kafka/events"

	logging "gitlab.com/lifegoeson-libs/pkg-logging"

	"gitlab.com/lifegoeson-libs/pkg-logging/logger"
)

type ArticleProducer struct {
	kafkaService *kafka.KafkaService
}

func NewArticleProducer(kafkaService *kafka.KafkaService, ctx context.Context) *ArticleProducer {
	return &ArticleProducer{
		kafkaService: kafkaService,
	}
}

func (p *ArticleProducer) PublishArticleCreatedEvent(ctx context.Context, event *events.ArticleCreatedEvent) error {
	return p.publishEvent(ctx, kafka.GetKafkaTopics().Article.Created, event.ArticleID, event)
}

func (p *ArticleProducer) PublishArticleDeletedEvent(ctx context.Context, event *events.ArticleDeletedEvent) error {
	return p.publishEvent(ctx, kafka.GetKafkaTopics().Article.Deleted, event.ArticleID, event)
}

func (p *ArticleProducer) PublishArticleLikedEvent(ctx context.Context, event *events.ArticleLikedEvent) error {
	return p.publishEvent(ctx, kafka.GetKafkaTopics().Article.Liked, event.ArticleID, event)
}

func (p *ArticleProducer) PublishArticleSharedEvent(ctx context.Context, event *events.ArticleSharedEvent) error {
	return p.publishEvent(ctx, kafka.GetKafkaTopics().Article.Shared, event.ArticleID, event)
}

func (p *ArticleProducer) PublishArticleUpdatedEvent(ctx context.Context, event *events.ArticleUpdatedEvent) error {
	return p.publishEvent(ctx, kafka.GetKafkaTopics().Article.Updated, event.ArticleID, event)
}

func (p *ArticleProducer) ArticleCreated(ctx context.Context, articleID, categoryID, authorID, title, description string) error {
	event := events.NewArticleCreatedEvent(articleID, categoryID, authorID, title, description)
	return p.PublishArticleCreatedEvent(ctx, event)
}

func (p *ArticleProducer) ArticleDeleted(ctx context.Context, articleID, categoryID, authorID string) error {
	event := events.NewArticleDeletedEvent(articleID, categoryID, authorID)
	return p.PublishArticleDeletedEvent(ctx, event)
}

func (p *ArticleProducer) ArticleViewed(ctx context.Context, articleID string) {
	event := events.NewArticleViewedEvent(articleID)
	p.publishEventAsync(ctx, kafka.GetKafkaTopics().Article.Viewed, event.ArticleID, event)
}

func (p *ArticleProducer) ArticleLiked(ctx context.Context, articleID, authorID string) {
	event := events.NewArticleLikedEvent(articleID, authorID)
	p.publishEventAsync(ctx, kafka.GetKafkaTopics().Article.Liked, event.ArticleID, event)
}

func (p *ArticleProducer) ArticleShared(ctx context.Context, articleID, authorID string) {
	event := events.NewArticleSharedEvent(articleID, authorID)
	p.publishEventAsync(ctx, kafka.GetKafkaTopics().Article.Shared, event.ArticleID, event)
}

func (p *ArticleProducer) ArticleCreatedWithData(ctx context.Context, article *entity.Article) error {
	eventData := p.convertArticleToEventData(article)
	event := events.NewArticleCreatedEventWithData(eventData)
	return p.PublishArticleCreatedEvent(ctx, event)
}

func (p *ArticleProducer) ArticleUpdated(ctx context.Context, article *entity.Article) error {
	eventData := p.convertArticleToEventData(article)
	event := events.NewArticleUpdatedEvent(eventData)
	return p.publishEvent(ctx, kafka.GetKafkaTopics().Article.Updated, event.ArticleID, event)
}

func (p *ArticleProducer) publishEvent(ctx context.Context, topic, key string, event interface{}) error {
	message := &kafka.ProducerMessage{
		Key:   key,
		Value: event,
	}
	if err := p.kafkaService.Emit(ctx, topic, message); err != nil {
		logger.FromContext(ctx).Error("Failed to publish article event", err,
			logging.String("topic", topic),
			logging.String("key", key),
		)
		return fmt.Errorf("failed to publish article event: %w", err)
	}
	logger.FromContext(ctx).Debug("Published article event",
		logging.String("topic", topic),
		logging.String("key", key),
	)
	return nil
}

func (p *ArticleProducer) publishEventAsync(ctx context.Context, topic, key string, event interface{}) {
	message := &kafka.ProducerMessage{
		Key:   key,
		Value: event,
	}
	p.kafkaService.EmitAsync(ctx, topic, message)
	logger.FromContext(ctx).Debug("Published article event async",
		logging.String("topic", topic),
		logging.String("key", key),
	)
}

func (p *ArticleProducer) convertArticleToEventData(article *entity.Article) *events.ArticleEventData {
	eventData := &events.ArticleEventData{
		ID:          article.IDHex,
		Slug:        article.Slug,
		Title:       article.Title,
		Description: article.Description,
		Content:     article.Content,
		Image:       article.Image,
		Status:      string(article.Status),
		IsPremium:   article.IsPremium,
		IsFeatured:  article.IsFeatured,
		ReadTime:    article.ReadTime,
		PublishedAt: article.PublishedAt,
		ScheduledAt: article.ScheduledAt,
		CreatedAt:   article.CreatedAt,
		UpdatedAt:   article.UpdatedAt,
	}
	eventData.Author = events.ArticleAuthorEventData{
		ID:        article.Author.IDHex,
		FirstName: article.Author.FirstName,
		LastName:  article.Author.LastName,
		Image:     article.Author.Image,
		Bio:       article.Author.Bio,
	}
	eventData.Category = events.ArticleCategoryEventData{
		ID:          article.Category.IDHex,
		Name:        article.Category.Name,
		Slug:        article.Category.Slug,
		Image:       article.Category.Image,
		Description: article.Category.Description,
	}
	var tags []events.ArticleTagEventData
	for _, tag := range article.Tags {
		tags = append(tags, events.ArticleTagEventData{
			ID:   tag.IDHex,
			Name: tag.Name,
		})
	}
	eventData.Tags = tags
	if article.SEO != nil {
		eventData.SEO = &events.ArticleSEOEventData{
			MetaTitle:       article.SEO.MetaTitle,
			MetaDescription: article.SEO.MetaDescription,
			MetaKeywords:    article.SEO.MetaKeywords,
			OGTitle:         article.SEO.OGTitle,
			OGDescription:   article.SEO.OGDescription,
			OGImage:         article.SEO.OGImage,
			TwitterCard:     article.SEO.TwitterCard,
			CanonicalURL:    article.SEO.CanonicalURL,
		}
	}
	return eventData
}
