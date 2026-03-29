package service

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/modami/core-service/internal/domain"
	"github.com/modami/core-service/internal/port"
	apperror "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
)

type MasterdataService struct {
	categories port.CategoryRepository
	packages   port.PackageRepository
	hashtags   port.HashtagRepository
}

func NewMasterdataService(cats port.CategoryRepository, pkgs port.PackageRepository, tags port.HashtagRepository) *MasterdataService {
	return &MasterdataService{categories: cats, packages: pkgs, hashtags: tags}
}

// domain.Category methods
func (s *MasterdataService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.ListAll(ctx, true)
}

func (s *MasterdataService) GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	c, err := s.categories.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to get category")
	}
	if c == nil {
		return nil, apperror.New(apperror.CodeNotFound,"category not found")
	}
	return c, nil
}

func (s *MasterdataService) GetCategoryChildren(ctx context.Context, slug string) ([]domain.Category, error) {
	parent, err := s.categories.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to get category")
	}
	if parent == nil {
		return nil, apperror.New(apperror.CodeNotFound,"category not found")
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
		return nil, apperror.New(apperror.CodeBadRequest,"invalid category id")
	}
	c, err := s.categories.GetByID(ctx, oid)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to get category")
	}
	if c == nil {
		return nil, apperror.New(apperror.CodeNotFound,"category not found")
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
		return nil, apperror.New(apperror.CodeInternal,"failed to toggle category")
	}
	return c, nil
}

func (s *MasterdataService) DeleteCategory(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return apperror.New(apperror.CodeBadRequest,"invalid category id")
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

// domain.Package methods
func (s *MasterdataService) ListPackages(ctx context.Context) ([]domain.Package, error) {
	return s.packages.ListActive(ctx)
}

func (s *MasterdataService) GetPackageByCode(ctx context.Context, code string) (*domain.Package, error) {
	p, err := s.packages.GetByCode(ctx, code)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to get package")
	}
	if p == nil {
		return nil, apperror.New(apperror.CodeNotFound,"package not found")
	}
	return p, nil
}

func (s *MasterdataService) GetPackageByID(ctx context.Context, id string) (*domain.Package, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid package id")
	}
	p, err := s.packages.GetByID(ctx, oid)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to get package")
	}
	if p == nil {
		return nil, apperror.New(apperror.CodeNotFound,"package not found")
	}
	return p, nil
}

func (s *MasterdataService) CreatePackage(ctx context.Context, p *domain.Package) error {
	return s.packages.Create(ctx, p)
}

func (s *MasterdataService) UpdatePackage(ctx context.Context, p *domain.Package) error {
	return s.packages.Update(ctx, p)
}

func (s *MasterdataService) TogglePackage(ctx context.Context, id string) (*domain.Package, error) {
	p, err := s.GetPackageByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.IsActive = !p.IsActive
	if err := s.packages.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to toggle package")
	}
	return p, nil
}

// domain.Hashtag methods
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
