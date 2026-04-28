package handler

import (
	"github.com/gin-gonic/gin"

	"be-modami-core-service/internal/dto"
	"be-modami-core-service/internal/service"
	"gitlab.com/lifegoeson-libs/pkg-gokit/mongodb/pagination"
	"be-modami-core-service/pkg/validator"
)

// BlogHandler exposes public and admin HTTP endpoints for the Community & Blog feature.
type BlogHandler struct {
	svc *service.BlogService
}

// NewBlogHandler creates a new BlogHandler.
func NewBlogHandler(svc *service.BlogService) *BlogHandler {
	return &BlogHandler{svc: svc}
}

// CommunityFeed godoc
// @Summary Community feed — featured post + recent posts
// @Description Returns the hero featured post and a cursor-paginated list of recent published posts for the community screen.
// @Tags Community
// @Produce json
// @Param cursor query string false "Pagination cursor"
// @Param limit  query int    false "Page size (default 20, max 100)"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /community [get]
func (h *BlogHandler) CommunityFeed(c *gin.Context) {
	cp := pagination.ParseCursor(c.Request)
	feed, err := h.svc.GetCommunityFeed(c.Request.Context(), cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, feed)
}

// ListPosts godoc
// @Summary Paginated blog post list
// @Description Returns published blog posts, optionally filtered by post_type.
// @Tags Blog
// @Produce json
// @Param post_type query string false "Post type filter"
// @Param cursor    query string false "Pagination cursor"
// @Param limit     query int    false "Page size (default 20, max 100)"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /blog/posts [get]
func (h *BlogHandler) ListPosts(c *gin.Context) {
	postType := c.Query("post_type")
	cp := pagination.ParseCursor(c.Request)
	posts, nextCursor, err := h.svc.ListPosts(c.Request.Context(), postType, cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, posts, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// GetPost godoc
// @Summary Blog post detail by slug
// @Description Returns the full content of a single published blog post.
// @Tags Blog
// @Produce json
// @Param slug path string true "URL slug"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /blog/posts/{slug} [get]
func (h *BlogHandler) GetPost(c *gin.Context) {
	slug := c.Param("slug")
	p, err := h.svc.GetPost(c.Request.Context(), slug)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}

// ListTrendReports godoc
// @Summary Monthly trend report listing
// @Description Returns cursor-paginated trend report posts (post_type = "trend_report").
// @Tags Blog
// @Produce json
// @Param cursor query string false "Pagination cursor"
// @Param limit  query int    false "Page size (default 20, max 100)"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /blog/reports [get]
func (h *BlogHandler) ListTrendReports(c *gin.Context) {
	cp := pagination.ParseCursor(c.Request)
	posts, nextCursor, err := h.svc.ListTrendReports(c.Request.Context(), cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, posts, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// HashtagPosts godoc
// @Summary Posts by hashtag
// @Description Returns published blog posts that carry the given hashtag.
// @Tags Blog
// @Produce json
// @Param tag    path  string true  "Hashtag (without #)"
// @Param cursor query string false "Pagination cursor"
// @Param limit  query int    false "Page size (default 20, max 100)"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /blog/hashtags/{tag} [get]
func (h *BlogHandler) HashtagPosts(c *gin.Context) {
	tag := c.Param("tag")
	cp := pagination.ParseCursor(c.Request)
	posts, nextCursor, err := h.svc.ListByHashtag(c.Request.Context(), tag, cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, posts, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// -- Admin handlers ----------------------------------------------------------

// AdminCreatePost godoc
// @Summary Admin — create blog post
// @Tags Blog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.CreateBlogPostRequest true "Blog post payload"
// @Success 201 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /blog/posts [post]
func (h *BlogHandler) CreatePost(c *gin.Context) {
	var req dto.CreateBlogPostRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}
	p, err := h.svc.CreatePost(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}
	created(c, p)
}

// AdminUpdatePost godoc
// @Summary Admin — update blog post
// @Tags Blog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path string                     true "Post ID"
// @Param body body dto.UpdateBlogPostRequest  true "Fields to update"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /blog/posts/{id} [put]
func (h *BlogHandler) UpdatePost(c *gin.Context) {
	var req dto.UpdateBlogPostRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}
	id := c.Param("id")
	p, err := h.svc.UpdatePost(c.Request.Context(), id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}

// AdminDeletePost godoc
// @Summary Admin — delete blog post
// @Tags Blog
// @Security BearerAuth
// @Param id path string true "Post ID"
// @Success 204 "No Content"
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /blog/posts/{id} [delete]
func (h *BlogHandler) DeletePost(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeletePost(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	noContent(c)
}

// AdminPublishPost godoc
// @Summary Admin — publish blog post
// @Tags Blog
// @Security BearerAuth
// @Param id path string true "Post ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /blog/posts/{id}/publish [post]
func (h *BlogHandler) PublishPost(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.PublishPost(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}
