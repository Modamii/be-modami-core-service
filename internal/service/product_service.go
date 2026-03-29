package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/modami/core-service/internal/domain"
	"github.com/modami/core-service/internal/dto"
	"github.com/modami/core-service/internal/port"
	apperror "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
)

type ProductService struct {
	repo port.ProductRepository
}

func NewProductService(repo port.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, sellerID string, req dto.CreateProductRequest) (*domain.Product, error) {
	sid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid seller_id")
	}

	catID, err := bson.ObjectIDFromHex(req.CategoryID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid category_id")
	}

	images := make([]domain.ProductImage, len(req.Images))
	for i, img := range req.Images {
		images[i] = domain.ProductImage{
			URL:      img.URL,
			Position: img.Position,
			Width:    img.Width,
			Height:   img.Height,
		}
	}

	p := &domain.Product{
		SellerID:    sid,
		Status:      domain.StatusDraft,
		Title:       req.Title,
		Slug:        generateSlug(req.Title),
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  catID,
		Condition:   req.Condition,
		Size:        req.Size,
		Brand:       req.Brand,
		Color:       req.Color,
		Material:    req.Material,
		Images:      images,
		Hashtags:    req.Hashtags,
		CreditCost:  req.CreditCost,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to create product")
	}

	_ = s.repo.InitStats(ctx, p.ID)
	return p, nil
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid product id")
	}
	p, err := s.repo.GetByID(ctx, oid)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to get product")
	}
	if p == nil {
		return nil, apperror.New(apperror.CodeNotFound,"product not found")
	}
	return p, nil
}

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	p, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to get product")
	}
	if p == nil {
		return nil, apperror.New(apperror.CodeNotFound,"product not found")
	}
	return p, nil
}

func (s *ProductService) Update(ctx context.Context, id string, sellerID string, req dto.UpdateProductRequest) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.SellerID.Hex() != sellerID {
		return nil, apperror.New(apperror.CodeForbidden,"you can only update your own products")
	}
	if p.Status != domain.StatusDraft && p.Status != domain.StatusArchived {
		return nil, apperror.New(apperror.CodeBadRequest,"can only update products in draft or archived status")
	}

	if req.Title != nil {
		p.Title = *req.Title
		p.Slug = generateSlug(*req.Title)
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.CategoryID != nil {
		catID, err := bson.ObjectIDFromHex(*req.CategoryID)
		if err != nil {
			return nil, apperror.New(apperror.CodeBadRequest,"invalid category_id")
		}
		p.CategoryID = catID
	}
	if req.Condition != nil {
		p.Condition = *req.Condition
	}
	if req.Size != nil {
		p.Size = *req.Size
	}
	if req.Brand != nil {
		p.Brand = *req.Brand
	}
	if req.Color != nil {
		p.Color = *req.Color
	}
	if req.Material != nil {
		p.Material = *req.Material
	}
	if req.Images != nil {
		images := make([]domain.ProductImage, len(req.Images))
		for i, img := range req.Images {
			images[i] = domain.ProductImage{URL: img.URL, Position: img.Position, Width: img.Width, Height: img.Height}
		}
		p.Images = images
	}
	if req.Hashtags != nil {
		p.Hashtags = req.Hashtags
	}
	if req.CreditCost != nil {
		p.CreditCost = *req.CreditCost
	}

	if err := s.repo.Update(ctx, p); err != nil {
		if err == domain.ErrProductVersionConflict {
			return nil, apperror.New(apperror.CodeConflict,"product was modified by another request, please retry")
		}
		return nil, apperror.New(apperror.CodeInternal,"failed to update product")
	}
	return p, nil
}

func (s *ProductService) Delete(ctx context.Context, id string, sellerID string) error {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if p.SellerID.Hex() != sellerID {
		return apperror.New(apperror.CodeForbidden,"you can only delete your own products")
	}
	if err := s.repo.SoftDelete(ctx, p.ID); err != nil {
		return apperror.New(apperror.CodeInternal,"failed to delete product")
	}
	return nil
}

func (s *ProductService) Submit(ctx context.Context, id string, sellerID string) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.SellerID.Hex() != sellerID {
		return nil, apperror.New(apperror.CodeForbidden,"you can only submit your own products")
	}
	if p.Status != domain.StatusDraft {
		return nil, apperror.New(apperror.CodeBadRequest,"can only submit products in draft status")
	}

	p.Status = domain.StatusPending
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to submit product")
	}

	latestMod, _ := s.repo.GetLatestModeration(ctx, p.ID)
	round := 1
	if latestMod != nil {
		round = latestMod.Round + 1
	}

	_ = s.repo.CreateModeration(ctx, &domain.ProductModeration{
		ProductID: p.ID,
		Round:     round,
		Action:    "submitted",
	})

	return p, nil
}

func (s *ProductService) Resubmit(ctx context.Context, id string, sellerID string, req dto.ResubmitRequest) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.SellerID.Hex() != sellerID {
		return nil, apperror.New(apperror.CodeForbidden,"you can only resubmit your own products")
	}
	if p.Status != domain.StatusDraft {
		return nil, apperror.New(apperror.CodeBadRequest,"can only resubmit products that were rejected (draft status)")
	}

	// Apply updates
	if req.Title != nil {
		p.Title = *req.Title
		p.Slug = generateSlug(*req.Title)
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.CategoryID != nil {
		catID, err := bson.ObjectIDFromHex(*req.CategoryID)
		if err != nil {
			return nil, apperror.New(apperror.CodeBadRequest,"invalid category_id")
		}
		p.CategoryID = catID
	}
	if req.Condition != nil {
		p.Condition = *req.Condition
	}
	if req.Size != nil {
		p.Size = *req.Size
	}
	if req.Brand != nil {
		p.Brand = *req.Brand
	}
	if req.Color != nil {
		p.Color = *req.Color
	}
	if req.Material != nil {
		p.Material = *req.Material
	}
	if req.Images != nil {
		images := make([]domain.ProductImage, len(req.Images))
		for i, img := range req.Images {
			images[i] = domain.ProductImage{URL: img.URL, Position: img.Position, Width: img.Width, Height: img.Height}
		}
		p.Images = images
	}
	if req.Hashtags != nil {
		p.Hashtags = req.Hashtags
	}

	p.Status = domain.StatusPending
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to resubmit product")
	}

	latestMod, _ := s.repo.GetLatestModeration(ctx, p.ID)
	round := 1
	if latestMod != nil {
		round = latestMod.Round + 1
	}

	_ = s.repo.CreateModeration(ctx, &domain.ProductModeration{
		ProductID: p.ID,
		Round:     round,
		Action:    "submitted",
		Note:      req.Note,
	})

	return p, nil
}

func (s *ProductService) Archive(ctx context.Context, id string, sellerID string) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.SellerID.Hex() != sellerID {
		return nil, apperror.New(apperror.CodeForbidden,"you can only archive your own products")
	}
	if p.Status != domain.StatusActive {
		return nil, apperror.New(apperror.CodeBadRequest,"can only archive active products")
	}
	p.Status = domain.StatusArchived
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to archive product")
	}
	return p, nil
}

func (s *ProductService) Unarchive(ctx context.Context, id string, sellerID string) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.SellerID.Hex() != sellerID {
		return nil, apperror.New(apperror.CodeForbidden,"you can only unarchive your own products")
	}
	if p.Status != domain.StatusArchived {
		return nil, apperror.New(apperror.CodeBadRequest,"can only unarchive archived products")
	}
	p.Status = domain.StatusActive
	now := time.Now()
	p.PublishedAt = &now
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to unarchive product")
	}
	return p, nil
}

func (s *ProductService) GetModeration(ctx context.Context, id string) ([]domain.ProductModeration, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid product id")
	}
	return s.repo.ListModerations(ctx, oid)
}

func (s *ProductService) MyProducts(ctx context.Context, sellerID string, status string, cursor string, limit int) ([]domain.Product, string, error) {
	sid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, "", apperror.New(apperror.CodeBadRequest,"invalid seller_id")
	}
	return s.repo.ListBySellerID(ctx, sid, status, cursor, limit)
}

func (s *ProductService) SellerProducts(ctx context.Context, sellerID string, cursor string, limit int) ([]domain.Product, string, error) {
	sid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, "", apperror.New(apperror.CodeBadRequest,"invalid seller_id")
	}
	return s.repo.ListBySellerID(ctx, sid, string(domain.StatusActive), cursor, limit)
}

func (s *ProductService) Feed(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error) {
	return s.repo.ListFeed(ctx, cursor, limit)
}

func (s *ProductService) Featured(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error) {
	return s.repo.ListFeatured(ctx, cursor, limit)
}

func (s *ProductService) Select(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error) {
	return s.repo.ListSelect(ctx, cursor, limit)
}

func (s *ProductService) Similar(ctx context.Context, id string, limit int) ([]domain.Product, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid product id")
	}
	return s.repo.ListSimilar(ctx, oid, limit)
}

func (s *ProductService) TrackView(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return apperror.New(apperror.CodeBadRequest,"invalid product id")
	}
	return s.repo.IncrementStat(ctx, oid, "view_count", 1)
}

func (s *ProductService) GetStats(ctx context.Context, id string) (*domain.ProductStats, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid product id")
	}
	return s.repo.GetStats(ctx, oid)
}

// Admin methods
func (s *ProductService) Approve(ctx context.Context, id string, moderatorID string) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status != domain.StatusPending {
		return nil, apperror.New(apperror.CodeBadRequest,"can only approve pending products")
	}
	p.Status = domain.StatusActive
	now := time.Now()
	p.PublishedAt = &now
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to approve product")
	}

	modID, _ := bson.ObjectIDFromHex(moderatorID)
	_ = s.repo.CreateModeration(ctx, &domain.ProductModeration{
		ProductID:   p.ID,
		Action:      "approved",
		ModeratorID: &modID,
	})
	return p, nil
}

func (s *ProductService) Reject(ctx context.Context, id string, moderatorID string, rejectCode string, reason string, suggestion string) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status != domain.StatusPending {
		return nil, apperror.New(apperror.CodeBadRequest,"can only reject pending products")
	}
	p.Status = domain.StatusDraft
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to reject product")
	}

	modID, _ := bson.ObjectIDFromHex(moderatorID)
	_ = s.repo.CreateModeration(ctx, &domain.ProductModeration{
		ProductID:   p.ID,
		Action:      "rejected",
		RejectCode:  rejectCode,
		Reason:      reason,
		Suggestion:  suggestion,
		ModeratorID: &modID,
	})
	return p, nil
}

func (s *ProductService) SetFeatured(ctx context.Context, id string, featured bool) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.IsFeatured = featured
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to update featured status")
	}
	return p, nil
}

func (s *ProductService) SetVerified(ctx context.Context, id string) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.IsVerified = true
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to verify product")
	}
	return p, nil
}

func (s *ProductService) SetSelect(ctx context.Context, id string, sp *domain.SelectProduct) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.IsSelect = true
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to set select")
	}
	sp.ProductID = p.ID
	_ = s.repo.CreateSelectProduct(ctx, sp)
	return p, nil
}

func (s *ProductService) HardDelete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return apperror.New(apperror.CodeBadRequest,"invalid product id")
	}
	return s.repo.SoftDelete(ctx, oid)
}

func (s *ProductService) ListPending(ctx context.Context, cursor string, limit int) ([]domain.Product, string, error) {
	return s.repo.ListPendingProducts(ctx, cursor, limit)
}

func (s *ProductService) CountByStatus(ctx context.Context, status domain.ProductStatus) (int64, error) {
	return s.repo.CountByStatus(ctx, status)
}

func (s *ProductService) Search(ctx context.Context, query string, params domain.ProductListParams, cursor string, limit int) ([]domain.Product, string, error) {
	return s.repo.Search(ctx, query, params, cursor, limit)
}

func (s *ProductService) ListByHashtag(ctx context.Context, tag string, cursor string, limit int) ([]domain.Product, string, error) {
	return s.repo.ListByHashtag(ctx, tag, cursor, limit)
}

func (s *ProductService) ListByIDs(ctx context.Context, ids []bson.ObjectID) ([]domain.Product, error) {
	return s.repo.ListByIDs(ctx, ids)
}

func (s *ProductService) IncrementStat(ctx context.Context, productID bson.ObjectID, field string, value int64) error {
	return s.repo.IncrementStat(ctx, productID, field, value)
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	var b strings.Builder
	for _, r := range slug {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' {
			b.WriteRune('-')
		}
	}
	result := b.String()
	// Remove consecutive dashes
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	result = strings.Trim(result, "-")
	// Append timestamp suffix for uniqueness
	return fmt.Sprintf("%s-%d", result, time.Now().UnixMilli()%100000)
}
