package repository

import (
	"context"
	"time"

	"github.com/modami/core-service/internal/domain"
	"github.com/modami/core-service/internal/port"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type analyticsMongoRepository struct {
	daily  *mongo.Collection
	seller *mongo.Collection
}

func NewAnalyticsRepository(db *mongo.Database) port.AnalyticsRepository {
	return &analyticsMongoRepository{
		daily:  db.Collection("daily_stats"),
		seller: db.Collection("seller_stats_snapshots"),
	}
}

func (r *analyticsMongoRepository) UpsertDailyStat(ctx context.Context, date string, field string, value int) error {
	_, err := r.daily.UpdateOne(ctx,
		bson.M{"date": date},
		bson.M{
			"$inc":         bson.M{field: value},
			"$setOnInsert": bson.M{"created_at": time.Now()},
		},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (r *analyticsMongoRepository) GetDailyStat(ctx context.Context, date string) (*domain.DailyStat, error) {
	var s domain.DailyStat
	err := r.daily.FindOne(ctx, bson.M{"date": date}).Decode(&s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &domain.DailyStat{Date: date}, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *analyticsMongoRepository) ListDailyStats(ctx context.Context, from, to string) ([]domain.DailyStat, error) {
	filter := bson.M{}
	if from != "" {
		filter["date"] = bson.M{"$gte": from}
	}
	if to != "" {
		if _, ok := filter["date"]; ok {
			filter["date"].(bson.M)["$lte"] = to
		} else {
			filter["date"] = bson.M{"$lte": to}
		}
	}
	opts := options.Find().SetSort(bson.D{{"date", -1}})
	cur, err := r.daily.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var stats []domain.DailyStat
	if err := cur.All(ctx, &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *analyticsMongoRepository) CreateSellerSnapshot(ctx context.Context, s *domain.SellerStatsSnapshot) error {
	s.CreatedAt = time.Now()
	_, err := r.seller.InsertOne(ctx, s)
	return err
}

func (r *analyticsMongoRepository) GetSellerSnapshot(ctx context.Context, sellerID bson.ObjectID, period string) (*domain.SellerStatsSnapshot, error) {
	var s domain.SellerStatsSnapshot
	err := r.seller.FindOne(ctx, bson.M{"seller_id": sellerID, "period": period}).Decode(&s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *analyticsMongoRepository) ListTopSellers(ctx context.Context, period string, limit int) ([]domain.SellerStatsSnapshot, error) {
	opts := options.Find().
		SetSort(bson.D{{"total_revenue", -1}}).
		SetLimit(int64(limit))
	cur, err := r.seller.Find(ctx, bson.M{"period": period}, opts)
	if err != nil {
		return nil, err
	}
	var snapshots []domain.SellerStatsSnapshot
	if err := cur.All(ctx, &snapshots); err != nil {
		return nil, err
	}
	return snapshots, nil
}
