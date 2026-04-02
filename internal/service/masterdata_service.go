package service

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

	"be-modami-core-service/internal/domain"
	"be-modami-core-service/internal/port"

	apperror "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
)

type MasterdataService struct {
	categories port.CategoryRepository
	hashtags   port.HashtagRepository
}

func NewMasterdataService(cats port.CategoryRepository, tags port.HashtagRepository) *MasterdataService {
	return &MasterdataService{categories: cats, hashtags: tags}
}

func (s *MasterdataService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.ListAll(ctx, true)
}

func (s *MasterdataService) GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	c, err := s.categories.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "lấy danh mục thất bại")
	}
	if c == nil {
		return nil, apperror.New(apperror.CodeNotFound, "không tìm thấy danh mục")
	}
	return c, nil
}

func (s *MasterdataService) GetCategoryChildren(ctx context.Context, slug string) ([]domain.Category, error) {
	parent, err := s.categories.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "lấy danh mục thất bại")
	}
	if parent == nil {
		return nil, apperror.New(apperror.CodeNotFound, "không tìm thấy danh mục")
	}
	return s.categories.ListChildren(ctx, parent.ID)
}

func (s *MasterdataService) CreateCategory(ctx context.Context, c *domain.Category) error {
	return s.categories.Create(ctx, c)
}

func (s *MasterdataService) UpdateCategory(ctx context.Context, c *domain.Category) error {
	return s.categories.Update(ctx, c)
}

func (s *MasterdataService) GetCategoryByID(ctx context.Context, id string) (*domain.Category, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest, "ID danh mục không hợp lệ")
	}
	c, err := s.categories.GetByID(ctx, oid)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "lấy danh mục thất bại")
	}
	if c == nil {
		return nil, apperror.New(apperror.CodeNotFound, "không tìm thấy danh mục")
	}
	return c, nil
}

func (s *MasterdataService) ToggleCategory(ctx context.Context, id string) (*domain.Category, error) {
	c, err := s.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	c.IsActive = !c.IsActive
	if err := s.categories.Update(ctx, c); err != nil {
		return nil, apperror.New(apperror.CodeInternal, "cập nhật trạng thái danh mục thất bại")
	}
	return c, nil
}

func (s *MasterdataService) DeleteCategory(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return apperror.New(apperror.CodeBadRequest, "ID danh mục không hợp lệ")
	}
	return s.categories.Delete(ctx, oid)
}

func (s *MasterdataService) ReorderCategories(ctx context.Context, orders []domain.CategoryOrder) error {
	for _, o := range orders {
		oid, err := bson.ObjectIDFromHex(o.ID)
		if err != nil {
			continue
		}
		c, err := s.categories.GetByID(ctx, oid)
		if err != nil || c == nil {
			continue
		}
		c.SortOrder = o.SortOrder
		_ = s.categories.Update(ctx, c)
	}
	return nil
}

func (s *MasterdataService) ListAllCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.ListAll(ctx, false)
}

func (s *MasterdataService) TrendingHashtags(ctx context.Context, limit int) ([]domain.Hashtag, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.hashtags.ListTrending(ctx, limit)
}

func (s *MasterdataService) SuggestHashtags(ctx context.Context, query string, limit int) ([]domain.Hashtag, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.hashtags.Search(ctx, query, limit)
}

func (s *MasterdataService) UpsertHashtags(ctx context.Context, tags []string, delta int64) {
	for _, tag := range tags {
		_ = s.hashtags.Upsert(ctx, tag, delta)
	}
}
