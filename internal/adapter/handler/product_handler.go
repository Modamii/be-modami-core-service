package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/modami/core-service/internal/adapter/handler/middleware"
	"github.com/modami/core-service/internal/domain"
	"github.com/modami/core-service/internal/dto"
	"github.com/modami/core-service/internal/service"
	"github.com/modami/core-service/pkg/storage/database/mongodb/pagination"
	"github.com/modami/core-service/pkg/validator"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// Create godoc
// @Summary Create product
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.CreateProductRequest true "Product payload"
// @Success 201 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}

	userID := middleware.UserID(c)
	p, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	created(c, p)
}

// GetByID godoc
// @Summary Get product by ID
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/{id} [get]
func (h *ProductHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}

// GetBySlug godoc
// @Summary Get product by slug
// @Tags Products
// @Produce json
// @Param slug path string true "URL slug"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/slug/{slug} [get]
func (h *ProductHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	p, err := h.svc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}

// Update godoc
// @Summary Update product
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param body body dto.UpdateProductRequest true "Fields to update"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	var req dto.UpdateProductRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}

	id := c.Param("id")
	userID := middleware.UserID(c)
	p, err := h.svc.Update(c.Request.Context(), id, userID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}

// Delete godoc
// @Summary Delete product
// @Tags Products
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 204 "No Content"
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.UserID(c)
	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		handleError(c, err)
		return
	}
	noContent(c)
}

// Submit godoc
// @Summary Submit product for review
// @Tags Products
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/{id}/submit [post]
func (h *ProductHandler) Submit(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.UserID(c)
	p, err := h.svc.Submit(c.Request.Context(), id, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}

// Resubmit godoc
// @Summary Resubmit rejected product
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param body body dto.ResubmitRequest true "Updates"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/{id}/resubmit [post]
func (h *ProductHandler) Resubmit(c *gin.Context) {
	var req dto.ResubmitRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}

	id := c.Param("id")
	userID := middleware.UserID(c)
	p, err := h.svc.Resubmit(c.Request.Context(), id, userID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}

// Archive godoc
// @Summary Archive product
// @Tags Products
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/{id}/archive [post]
func (h *ProductHandler) Archive(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.UserID(c)
	p, err := h.svc.Archive(c.Request.Context(), id, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}

// Unarchive godoc
// @Summary Unarchive product
// @Tags Products
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/{id}/unarchive [post]
func (h *ProductHandler) Unarchive(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.UserID(c)
	p, err := h.svc.Unarchive(c.Request.Context(), id, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}

// GetModeration godoc
// @Summary List moderation history for product
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/{id}/moderation [get]
func (h *ProductHandler) GetModeration(c *gin.Context) {
	id := c.Param("id")
	mods, err := h.svc.GetModeration(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, mods)
}

// MyProducts godoc
// @Summary List my products (seller)
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/me [get]
func (h *ProductHandler) MyProducts(c *gin.Context) {
	userID := middleware.UserID(c)
	status := c.Query("status")
	cp := pagination.ParseCursor(c.Request)
	products, nextCursor, err := h.svc.MyProducts(c.Request.Context(), userID, status, cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, products, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// SellerProducts godoc
// @Summary List products by seller
// @Tags Products
// @Produce json
// @Param seller_id path string true "Seller user ID"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /sellers/{seller_id}/products [get]
func (h *ProductHandler) SellerProducts(c *gin.Context) {
	sellerID := c.Param("id")
	cp := pagination.ParseCursor(c.Request)
	products, nextCursor, err := h.svc.SellerProducts(c.Request.Context(), sellerID, cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, products, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// Feed godoc
// @Summary Product feed (cursor)
// @Tags Products
// @Produce json
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/feed [get]
func (h *ProductHandler) Feed(c *gin.Context) {
	cp := pagination.ParseCursor(c.Request)
	products, nextCursor, err := h.svc.Feed(c.Request.Context(), cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, products, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// Featured godoc
// @Summary Featured products
// @Tags Products
// @Produce json
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/featured [get]
func (h *ProductHandler) Featured(c *gin.Context) {
	cp := pagination.ParseCursor(c.Request)
	products, nextCursor, err := h.svc.Featured(c.Request.Context(), cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, products, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// SelectProducts godoc
// @Summary Curated select products
// @Tags Products
// @Produce json
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/select [get]
func (h *ProductHandler) SelectProducts(c *gin.Context) {
	cp := pagination.ParseCursor(c.Request)
	products, nextCursor, err := h.svc.Select(c.Request.Context(), cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, products, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// Similar godoc
// @Summary Similar products
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Param limit query int false "Max items (default 10, max 50)"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/{id}/similar [get]
func (h *ProductHandler) Similar(c *gin.Context) {
	id := c.Param("id")
	limitStr := c.Query("limit")
	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
		limit = l
	}
	products, err := h.svc.Similar(c.Request.Context(), id, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, products)
}

// TrackView godoc
// @Summary Track product view
// @Tags Products
// @Param id path string true "Product ID"
// @Success 204 "No Content"
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/{id}/view [post]
func (h *ProductHandler) TrackView(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.TrackView(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	noContent(c)
}

// Search godoc
// @Summary Search products (catalog)
// @Tags Search
// @Produce json
// @Param q query string false "Search query"
// @Param category_id query string false "Category filter"
// @Param condition query string false "Condition"
// @Param brand query string false "Brand"
// @Param min_price query int false "Min price"
// @Param max_price query int false "Max price"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /products/search [get]
func (h *ProductHandler) Search(c *gin.Context) {
	q := c.Query("q")
	cp := pagination.ParseCursor(c.Request)
	params := domain.ProductListParams{
		CategoryID: c.Query("category_id"),
		Condition:  c.Query("condition"),
		Brand:      c.Query("brand"),
	}
	if v, err := strconv.ParseInt(c.Query("min_price"), 10, 64); err == nil {
		params.MinPrice = v
	}
	if v, err := strconv.ParseInt(c.Query("max_price"), 10, 64); err == nil {
		params.MaxPrice = v
	}

	products, nextCursor, err := h.svc.Search(c.Request.Context(), q, params, cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, products, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// HashtagProducts godoc
// @Summary Products by hashtag
// @Tags Products
// @Produce json
// @Param tag path string true "Hashtag (without #)"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /hashtags/{tag}/products [get]
func (h *ProductHandler) HashtagProducts(c *gin.Context) {
	tag := c.Param("tag")
	cp := pagination.ParseCursor(c.Request)
	products, nextCursor, err := h.svc.ListByHashtag(c.Request.Context(), tag, cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, products, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}
