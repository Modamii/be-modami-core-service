package service

import (
	"context"

	"be-modami-core-service/internal/domain"
	"be-modami-core-service/internal/port"
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
	followers, _ := s.followRepo.CountFollowers(ctx, sellerID)
	following, _ := s.followRepo.CountFollowing(ctx, sellerID)

	return &SellerProfile{
		SellerID:       sellerID,
		FollowerCount:  followers,
		FollowingCount: following,
	}, nil
}

func (s *SellerService) GetProducts(ctx context.Context, sellerID string, cursor string, limit int) ([]domain.Product, string, error) {
	return s.productRepo.ListBySellerID(ctx, sellerID, string(domain.StatusActive), cursor, limit)
}

func (s *SellerService) GetReviews(ctx context.Context, sellerID string, cursor string, limit int) ([]domain.Review, string, error) {
	return s.reviewRepo.ListBySeller(ctx, sellerID, cursor, limit)
}

func (s *SellerService) GetPublicStats(ctx context.Context, sellerID string) (*SellerPublicStats, error) {
	activeCount, _ := s.productRepo.CountByStatus(ctx, domain.StatusActive)
	soldCount, _ := s.productRepo.CountByStatus(ctx, domain.StatusSold)
	followers, _ := s.followRepo.CountFollowers(ctx, sellerID)

	return &SellerPublicStats{
		TotalProducts: activeCount,
		TotalSold:     soldCount,
		FollowerCount: followers,
	}, nil
}
