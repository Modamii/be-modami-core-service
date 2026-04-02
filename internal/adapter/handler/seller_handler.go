package handler

import (
	"github.com/gin-gonic/gin"

	"be-modami-core-service/internal/service"
	"be-modami-core-service/pkg/storage/database/mongodb/pagination"
)

type SellerHandler struct {
	svc *service.SellerService
}

func NewSellerHandler(svc *service.SellerService) *SellerHandler {
	return &SellerHandler{svc: svc}
}

// GetProfile godoc
// @Summary Public seller profile
// @Tags Sellers
// @Produce json
// @Param id path string true "Seller user ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /sellers/{id} [get]
func (h *SellerHandler) GetProfile(c *gin.Context) {
	sellerID := c.Param("id")
	profile, err := h.svc.GetProfile(c.Request.Context(), sellerID)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, profile)
}

// GetProducts godoc
// @Summary Seller's listed products
// @Tags Sellers
// @Produce json
// @Param id path string true "Seller user ID"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /sellers/{id}/products [get]
func (h *SellerHandler) GetProducts(c *gin.Context) {
	sellerID := c.Param("id")
	cp := pagination.ParseCursor(c.Request)
	products, nextCursor, err := h.svc.GetProducts(c.Request.Context(), sellerID, cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, products, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// GetReviews godoc
// @Summary Seller reviews
// @Tags Sellers
// @Produce json
// @Param id path string true "Seller user ID"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /sellers/{id}/reviews [get]
func (h *SellerHandler) GetReviews(c *gin.Context) {
	sellerID := c.Param("id")
	cp := pagination.ParseCursor(c.Request)
	reviews, nextCursor, err := h.svc.GetReviews(c.Request.Context(), sellerID, cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, reviews, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// GetStats godoc
// @Summary Seller public stats
// @Tags Sellers
// @Produce json
// @Param id path string true "Seller user ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /sellers/{id}/stats [get]
func (h *SellerHandler) GetStats(c *gin.Context) {
	sellerID := c.Param("id")
	stats, err := h.svc.GetPublicStats(c.Request.Context(), sellerID)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, stats)
}
