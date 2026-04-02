package service

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"be-modami-core-service/internal/domain"
	"be-modami-core-service/internal/dto"
	"be-modami-core-service/internal/port"

	apperror "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
)

// CommunityFeed is the composite response for the community feed endpoint.
type CommunityFeed struct {
	Featured *domain.BlogPost  `json:"featured"`
	Posts    []domain.BlogPost `json:"posts"`
	Next     string            `json:"next_cursor,omitempty"`
	HasMore  bool              `json:"has_more"`
}

// BlogService provides business logic for the Community & Blog feature.
type BlogService struct {
	repo port.BlogRepository
}

// NewBlogService creates a new BlogService backed by the given repository.
func NewBlogService(repo port.BlogRepository) *BlogService {
	return &BlogService{repo: repo}
}

// GetCommunityFeed returns the featured post together with the most recent
// published posts for the community feed screen.
func (s *BlogService) GetCommunityFeed(ctx context.Context, cursor string, limit int) (*CommunityFeed, error) {
	featured, err := s.repo.GetFeatured(ctx)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "lấy bài viết nổi bật thất bại")
	}

	posts, next, err := s.repo.List(ctx, "", cursor, limit)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "lấy danh sách bài viết thất bại")
	}

	return &CommunityFeed{
		Featured: featured,
		Posts:    posts,
		Next:     next,
		HasMore:  next != "",
	}, nil
}

// GetPost returns a single published blog post by slug.
func (s *BlogService) GetPost(ctx context.Context, slug string) (*domain.BlogPost, error) {
	p, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "lấy bài viết thất bại")
	}
	if p == nil {
		return nil, apperror.New(apperror.CodeNotFound, "không tìm thấy bài viết")
	}
	return p, nil
}

// ListPosts returns a cursor-paginated list of published posts, optionally
// filtered by postType.
func (s *BlogService) ListPosts(ctx context.Context, postType string, cursor string, limit int) ([]domain.BlogPost, string, error) {
	posts, next, err := s.repo.List(ctx, postType, cursor, limit)
	if err != nil {
		return nil, "", apperror.New(apperror.CodeInternal, "lấy danh sách bài viết thất bại")
	}
	return posts, next, nil
}

// ListTrendReports returns the paginated monthly trend report listing.
func (s *BlogService) ListTrendReports(ctx context.Context, cursor string, limit int) ([]domain.BlogPost, string, error) {
	posts, next, err := s.repo.ListTrendReports(ctx, cursor, limit)
	if err != nil {
		return nil, "", apperror.New(apperror.CodeInternal, "lấy danh sách báo cáo xu hướng thất bại")
	}
	return posts, next, nil
}

// ListByHashtag returns posts tagged with the given hashtag.
func (s *BlogService) ListByHashtag(ctx context.Context, tag string, cursor string, limit int) ([]domain.BlogPost, string, error) {
	posts, next, err := s.repo.ListByHashtag(ctx, tag, cursor, limit)
	if err != nil {
		return nil, "", apperror.New(apperror.CodeInternal, "lấy bài viết theo hashtag thất bại")
	}
	return posts, next, nil
}

// -- Admin operations --------------------------------------------------------

// CreatePost creates a new blog post from the admin request payload.
func (s *BlogService) CreatePost(ctx context.Context, req dto.CreateBlogPostRequest) (*domain.BlogPost, error) {
	p := &domain.BlogPost{
		Slug:          req.Slug,
		SeriesName:    req.SeriesName,
		SeriesNo:      req.SeriesNo,
		SeriesQuarter: req.SeriesQuarter,
		PostType:      req.PostType,
		Depth:         domain.PostDepth(req.Depth),
		Title:         req.Title,
		Subtitle:      req.Subtitle,
		Body:          req.Body,
		CoverImage:    req.CoverImage,
		CoverCaption:  req.CoverCaption,
		ReadTimeMin:   req.ReadTimeMin,
		WordCount:     req.WordCount,
		Author: domain.BlogAuthor{
			Name:  req.Author.Name,
			Title: req.Author.Title,
			Bio:   req.Author.Bio,
		},
		KeyPoints:  req.KeyPoints,
		References: req.References,
		Hashtags:   req.Hashtags,
		CTALink:    req.CTALink,
		IsFeatured: req.IsFeatured,
		Status:     domain.PostStatusDraft,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal, "tạo bài viết thất bại")
	}
	return p, nil
}

// UpdatePost applies partial updates from the admin request to the stored post.
func (s *BlogService) UpdatePost(ctx context.Context, id string, req dto.UpdateBlogPostRequest) (*domain.BlogPost, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest, "ID bài viết không hợp lệ")
	}

	p, err := s.repo.GetByID(ctx, oid)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "lấy bài viết thất bại")
	}
	if p == nil {
		return nil, apperror.New(apperror.CodeNotFound, "không tìm thấy bài viết")
	}

	if req.Slug != nil {
		p.Slug = *req.Slug
	}
	if req.SeriesName != nil {
		p.SeriesName = *req.SeriesName
	}
	if req.SeriesNo != nil {
		p.SeriesNo = *req.SeriesNo
	}
	if req.SeriesQuarter != nil {
		p.SeriesQuarter = *req.SeriesQuarter
	}
	if req.PostType != nil {
		p.PostType = *req.PostType
	}
	if req.Depth != nil {
		p.Depth = domain.PostDepth(*req.Depth)
	}
	if req.Title != nil {
		p.Title = *req.Title
	}
	if req.Subtitle != nil {
		p.Subtitle = *req.Subtitle
	}
	if req.Body != nil {
		p.Body = *req.Body
	}
	if req.CoverImage != nil {
		p.CoverImage = *req.CoverImage
	}
	if req.CoverCaption != nil {
		p.CoverCaption = *req.CoverCaption
	}
	if req.ReadTimeMin != nil {
		p.ReadTimeMin = *req.ReadTimeMin
	}
	if req.WordCount != nil {
		p.WordCount = *req.WordCount
	}
	if req.Author != nil {
		p.Author = domain.BlogAuthor{
			Name:  req.Author.Name,
			Title: req.Author.Title,
			Bio:   req.Author.Bio,
		}
	}
	if req.KeyPoints != nil {
		p.KeyPoints = req.KeyPoints
	}
	if req.References != nil {
		p.References = req.References
	}
	if req.Hashtags != nil {
		p.Hashtags = req.Hashtags
	}
	if req.CTALink != nil {
		p.CTALink = *req.CTALink
	}
	if req.IsFeatured != nil {
		p.IsFeatured = *req.IsFeatured
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal, "cập nhật bài viết thất bại")
	}
	return p, nil
}

// DeletePost hard-deletes a blog post by ID.
func (s *BlogService) DeletePost(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return apperror.New(apperror.CodeBadRequest, "ID bài viết không hợp lệ")
	}
	p, err := s.repo.GetByID(ctx, oid)
	if err != nil {
		return apperror.New(apperror.CodeInternal, "lấy bài viết thất bại")
	}
	if p == nil {
		return apperror.New(apperror.CodeNotFound, "không tìm thấy bài viết")
	}
	if err := s.repo.Delete(ctx, oid); err != nil {
		return apperror.New(apperror.CodeInternal, "xóa bài viết thất bại")
	}
	return nil
}

// PublishPost transitions a draft post to published status and records the publish time.
func (s *BlogService) PublishPost(ctx context.Context, id string) (*domain.BlogPost, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest, "ID bài viết không hợp lệ")
	}

	p, err := s.repo.GetByID(ctx, oid)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, "lấy bài viết thất bại")
	}
	if p == nil {
		return nil, apperror.New(apperror.CodeNotFound, "không tìm thấy bài viết")
	}
	if p.Status == domain.PostStatusPublished {
		return nil, apperror.New(apperror.CodeBadRequest, "bài viết đã được xuất bản")
	}

	now := time.Now()
	p.Status = domain.PostStatusPublished
	p.PublishedAt = &now

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.New(apperror.CodeInternal, "xuất bản bài viết thất bại")
	}
	return p, nil
}
