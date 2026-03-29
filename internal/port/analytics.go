package port

import (
	"context"

	"github.com/modami/core-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AnalyticsRepository interface {
	UpsertDailyStat(ctx context.Context, date string, field string, value int) error
	GetDailyStat(ctx context.Context, date string) (*domain.DailyStat, error)
	ListDailyStats(ctx context.Context, from, to string) ([]domain.DailyStat, error)

	CreateSellerSnapshot(ctx context.Context, s *domain.SellerStatsSnapshot) error
	GetSellerSnapshot(ctx context.Context, sellerID bson.ObjectID, period string) (*domain.SellerStatsSnapshot, error)
	ListTopSellers(ctx context.Context, period string, limit int) ([]domain.SellerStatsSnapshot, error)
}
