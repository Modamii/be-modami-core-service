package service

import (
	"context"
	"sync"

	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"

	"be-modami-core-service/internal/domain"
	"be-modami-core-service/internal/port"
)

type HomeFeedResponse struct {
	News       []domain.Product  `json:"news"`
	Categories []domain.Category `json:"categories"`
	Near       []domain.Product  `json:"near"`
	Blogs      []domain.BlogPost `json:"blogs"`
}

type HomeFeedService struct {
	productRepo  port.ProductRepository
	categoryRepo port.CategoryRepository
	blogRepo     port.BlogRepository
}

func NewHomeFeedService(
	productRepo port.ProductRepository,
	categoryRepo port.CategoryRepository,
	blogRepo port.BlogRepository,
) *HomeFeedService {
	return &HomeFeedService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		blogRepo:     blogRepo,
	}
}

func (s *HomeFeedService) GetHomeFeed(ctx context.Context) *HomeFeedResponse {
	var (
		news       []domain.Product
		categories []domain.Category
		near       []domain.Product
		blogs      []domain.BlogPost
	)

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		items, _, err := s.productRepo.ListFeed(ctx, "", 10)
		if err != nil {
			logger.FromContext(ctx).Error("home feed: fetch news failed", err)
			news = []domain.Product{}
			return
		}
		news = items
	}()

	go func() {
		defer wg.Done()
		all, err := s.categoryRepo.ListAll(ctx, true)
		if err != nil {
			logger.FromContext(ctx).Error("home feed: fetch categories failed", err)
			categories = []domain.Category{}
			return
		}
		categories = topLevelCategories(all, 4)
	}()

	go func() {
		defer wg.Done()
		// Near: featured active products — placeholder for future geo-based logic.
		items, _, err := s.productRepo.ListFeatured(ctx, "", 4)
		if err != nil {
			logger.FromContext(ctx).Error("home feed: fetch near failed", err,
				logging.String("section", "near"),
			)
			near = []domain.Product{}
			return
		}
		near = items
	}()

	go func() {
		defer wg.Done()
		items, _, err := s.blogRepo.List(ctx, "", "", 10)
		if err != nil {
			logger.FromContext(ctx).Error("home feed: fetch blogs failed", err)
			blogs = []domain.BlogPost{}
			return
		}
		blogs = items
	}()

	wg.Wait()

	return &HomeFeedResponse{
		News:       nonNilProducts(news),
		Categories: nonNilCategories(categories),
		Near:       nonNilProducts(near),
		Blogs:      nonNilBlogs(blogs),
	}
}

// topLevelCategories returns up to limit root categories (ParentID == nil).
func topLevelCategories(all []domain.Category, limit int) []domain.Category {
	result := make([]domain.Category, 0, limit)
	for _, c := range all {
		if c.ParentID == nil {
			result = append(result, c)
			if len(result) == limit {
				break
			}
		}
	}
	return result
}

func nonNilProducts(s []domain.Product) []domain.Product {
	if s == nil {
		return []domain.Product{}
	}
	return s
}

func nonNilCategories(s []domain.Category) []domain.Category {
	if s == nil {
		return []domain.Category{}
	}
	return s
}

func nonNilBlogs(s []domain.BlogPost) []domain.BlogPost {
	if s == nil {
		return []domain.BlogPost{}
	}
	return s
}
