package service

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/modami/core-service/internal/domain"
	"github.com/modami/core-service/internal/port"
	apperror "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
)

type SellerService struct {
	productRepo  port.ProductRepository
	favoriteRepo port.FavoriteRepository
	followRepo   port.FollowRepository
	reviewRepo   port.ReviewRepository
}

func NewSellerService(
	productRepo port.ProductRepository,
	favoriteRepo port.FavoriteRepository,
	followRepo port.FollowRepository,
	reviewRepo port.ReviewRepository,
) *SellerService {
	return &SellerService{
		productRepo:  productRepo,
		favoriteRepo: favoriteRepo,
		followRepo:   followRepo,
		reviewRepo:   reviewRepo,
	}
}

type SellerProfile struct {
	SellerID       string  `json:"seller_id"`
	ProductCount   int     `json:"product_count"`
	FollowerCount  int64   `json:"follower_count"`
	FollowingCount int64   `json:"following_count"`
	AvgRating      float64 `json:"avg_rating"`
	ReviewCount    int64   `json:"review_count"`
}

type SellerPublicStats struct {
	TotalProducts int64   `json:"total_products"`
	TotalSold     int64   `json:"total_sold"`
	AvgRating     float64 `json:"avg_rating"`
	ReviewCount   int64   `json:"review_count"`
	FollowerCount int64   `json:"follower_count"`
}

func (s *SellerService) GetProfile(ctx context.Context, sellerID string) (*SellerProfile, error) {
	oid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid seller id")
	}

	followers, _ := s.followRepo.CountFollowers(ctx, oid)
	following, _ := s.followRepo.CountFollowing(ctx, oid)

	return &SellerProfile{
		SellerID:       sellerID,
		FollowerCount:  followers,
		FollowingCount: following,
	}, nil
}

func (s *SellerService) GetProducts(ctx context.Context, sellerID string, cursor string, limit int) ([]domain.Product, string, error) {
	oid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, "", apperror.New(apperror.CodeBadRequest,"invalid seller id")
	}
	return s.productRepo.ListBySellerID(ctx, oid, string(domain.StatusActive), cursor, limit)
}

func (s *SellerService) GetReviews(ctx context.Context, sellerID string, cursor string, limit int) ([]domain.Review, string, error) {
	oid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, "", apperror.New(apperror.CodeBadRequest,"invalid seller id")
	}
	return s.reviewRepo.ListBySeller(ctx, oid, cursor, limit)
}

func (s *SellerService) GetPublicStats(ctx context.Context, sellerID string) (*SellerPublicStats, error) {
	oid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid seller id")
	}

	activeCount, _ := s.productRepo.CountByStatus(ctx, domain.StatusActive)
	soldCount, _ := s.productRepo.CountByStatus(ctx, domain.StatusSold)
	followers, _ := s.followRepo.CountFollowers(ctx, oid)

	return &SellerPublicStats{
		TotalProducts: activeCount,
		TotalSold:     soldCount,
		FollowerCount: followers,
	}, nil
}
