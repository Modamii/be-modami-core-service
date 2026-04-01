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
	repo     port.ProductRepository
	catRepo  port.CategoryRepository
}

func NewProductService(repo port.ProductRepository, catRepo port.CategoryRepository) *ProductService {
	return &ProductService{repo: repo, catRepo: catRepo}
}

func (s *ProductService) Create(ctx context.Context, sellerID string, req dto.CreateProductRequest) (*domain.Product, error) {
	sid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"seller_id không hợp lệ")
	}

	catID, err := bson.ObjectIDFromHex(req.CategoryID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest, "category_id không hợp lệ")
	}
	cat, err := s.catRepo.GetByID(ctx, catID)
	if err != nil || cat == nil {
		return nil, apperror.New(apperror.CodeBadRequest, "không tìm thấy danh mục")
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
		Category:    cat,
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
		return nil, apperror.New(apperror.CodeInternal,"tạo sản phẩm thất bại")
	}

	_ = s.repo.InitStats(ctx, p.ID)
	return p, nil
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"ID sản phẩm không hợp lệ")
	}
	p, err := s.repo.GetByID(ctx, oid)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal,"lấy sản phẩm thất bại")
	}
	if p == nil {
		return nil, apperror.New(apperror.CodeNotFound,"không tìm thấy sản phẩm")
	}
	return p, nil
}

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	p, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal,"lấy sản phẩm thất bại")
	}
	if p == nil {
		return nil, apperror.New(apperror.CodeNotFound,"không tìm thấy sản phẩm")
	}
	return p, nil
}

func (s *ProductService) Update(ctx context.Context, id string, sellerID string, req dto.UpdateProductRequest) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.SellerID.Hex() != sellerID {
		return nil, apperror.New(apperror.CodeForbidden,"bạn chỉ có thể cập nhật sản phẩm của mình")
	}
	if p.Status != domain.StatusDraft && p.Status != domain.StatusArchived {
		return nil, apperror.New(apperror.CodeBadRequest,"chỉ có thể cập nhật sản phẩm ở trạng thái nháp hoặc đã lưu trữ")
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
			return nil, apperror.New(apperror.CodeBadRequest, "category_id không hợp lệ")
		}
		cat, err := s.catRepo.GetByID(ctx, catID)
		if err != nil || cat == nil {
			return nil, apperror.New(apperror.CodeBadRequest, "không tìm thấy danh mục")
		}
		p.Category = cat
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
			return nil, apperror.New(apperror.CodeConflict, "sản phẩm đã bị thay đổi bởi yêu cầu khác, vui lòng thử lại")
		}
		return nil, apperror.New(apperror.CodeInternal, "cập nhật sản phẩm thất bại")
	}
	return p, nil
}

func (s *ProductService) Delete(ctx context.Context, id string, sellerID string) error {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if p.SellerID.Hex() != sellerID {
		return apperror.New(apperror.CodeForbidden,"bạn chỉ có thể xóa sản phẩm của mình")
	}
	if err := s.repo.SoftDelete(ctx, p.ID); err != nil {
		return apperror.New(apperror.CodeInternal,"xóa sản phẩm thất bại")
	}
	return nil
}

func (s *ProductService) Submit(ctx context.Context, id string, sellerID string) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.SellerID.Hex() != sellerID {
		return nil, apperror.New(apperror.CodeForbidden,"bạn chỉ có thể gửi duyệt sản phẩm của mình")
	}
	if p.Status != domain.StatusDraft {
		return nil, apperror.New(apperror.CodeBadRequest,"chỉ có thể gửi duyệt sản phẩm ở trạng thái nháp")
	}

	p.Status = domain.StatusPending
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"gửi duyệt sản phẩm thất bại")
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
		return nil, apperror.New(apperror.CodeForbidden,"bạn chỉ có thể gửi lại sản phẩm của mình")
	}
	if p.Status != domain.StatusDraft {
		return nil, apperror.New(apperror.CodeBadRequest,"chỉ có thể gửi lại sản phẩm bị từ chối (trạng thái nháp)")
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
			return nil, apperror.New(apperror.CodeBadRequest, "category_id không hợp lệ")
		}
		cat, err := s.catRepo.GetByID(ctx, catID)
		if err != nil || cat == nil {
			return nil, apperror.New(apperror.CodeBadRequest, "không tìm thấy danh mục")
		}
		p.Category = cat
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
		return nil, apperror.New(apperror.CodeInternal,"gửi lại sản phẩm thất bại")
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
		return nil, apperror.New(apperror.CodeForbidden,"bạn chỉ có thể lưu trữ sản phẩm của mình")
	}
	if p.Status != domain.StatusActive {
		return nil, apperror.New(apperror.CodeBadRequest,"chỉ có thể lưu trữ sản phẩm đang hoạt động")
	}
	p.Status = domain.StatusArchived
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"lưu trữ sản phẩm thất bại")
	}
	return p, nil
}

func (s *ProductService) Unarchive(ctx context.Context, id string, sellerID string) (*domain.Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.SellerID.Hex() != sellerID {
		return nil, apperror.New(apperror.CodeForbidden,"bạn chỉ có thể khôi phục sản phẩm của mình")
	}
	if p.Status != domain.StatusArchived {
		return nil, apperror.New(apperror.CodeBadRequest,"chỉ có thể khôi phục sản phẩm đã lưu trữ")
	}
	p.Status = domain.StatusActive
	now := time.Now()
	p.PublishedAt = &now
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"khôi phục sản phẩm thất bại")
	}
	return p, nil
}

func (s *ProductService) GetModeration(ctx context.Context, id string) ([]domain.ProductModeration, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"ID sản phẩm không hợp lệ")
	}
	return s.repo.ListModerations(ctx, oid)
}

func (s *ProductService) MyProducts(ctx context.Context, sellerID string, status string, cursor string, limit int) ([]domain.Product, string, error) {
	sid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, "", apperror.New(apperror.CodeBadRequest,"seller_id không hợp lệ")
	}
	return s.repo.ListBySellerID(ctx, sid, status, cursor, limit)
}

func (s *ProductService) SellerProducts(ctx context.Context, sellerID string, cursor string, limit int) ([]domain.Product, string, error) {
	sid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, "", apperror.New(apperror.CodeBadRequest,"seller_id không hợp lệ")
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
		return nil, apperror.New(apperror.CodeBadRequest,"ID sản phẩm không hợp lệ")
	}
	return s.repo.ListSimilar(ctx, oid, limit)
}

func (s *ProductService) TrackView(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return apperror.New(apperror.CodeBadRequest,"ID sản phẩm không hợp lệ")
	}
	return s.repo.IncrementStat(ctx, oid, "view_count", 1)
}

func (s *ProductService) GetStats(ctx context.Context, id string) (*domain.ProductStats, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"ID sản phẩm không hợp lệ")
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
		return apperror.New(apperror.CodeBadRequest,"ID sản phẩm không hợp lệ")
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
