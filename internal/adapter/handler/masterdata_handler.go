package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/modami/core-service/internal/domain"
	"github.com/modami/core-service/internal/service"
	"github.com/modami/core-service/pkg/validator"
	"go.mongodb.org/mongo-driver/v2/bson"
	"gitlab.com/lifegoeson-libs/pkg-gokit/response"
)

type MasterdataHandler struct {
	svc *service.MasterdataService
}

func NewMasterdataHandler(svc *service.MasterdataService) *MasterdataHandler {
	return &MasterdataHandler{svc: svc}
}

// ListCategories godoc
// @Summary List categories
// @Tags Categories
// @Produce json
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /categories [get]
func (h *MasterdataHandler) ListCategories(c *gin.Context) {
	cats, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, cats)
}

// GetCategory godoc
// @Summary Get category by slug
// @Tags Categories
// @Produce json
// @Param slug path string true "Category slug"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /categories/{slug} [get]
func (h *MasterdataHandler) GetCategory(c *gin.Context) {
	slug := c.Param("slug")
	cat, err := h.svc.GetCategoryBySlug(c.Request.Context(), slug)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, cat)
}

// GetCategoryChildren godoc
// @Summary List child categories
// @Tags Categories
// @Produce json
// @Param slug path string true "Parent category slug"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /categories/{slug}/children [get]
func (h *MasterdataHandler) GetCategoryChildren(c *gin.Context) {
	slug := c.Param("slug")
	children, err := h.svc.GetCategoryChildren(c.Request.Context(), slug)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, children)
}

// ListPackages godoc
// @Summary List subscription packages
// @Tags Packages
// @Produce json
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /packages [get]
func (h *MasterdataHandler) ListPackages(c *gin.Context) {
	pkgs, err := h.svc.ListPackages(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, pkgs)
}

// GetPackage godoc
// @Summary Get package by code
// @Tags Packages
// @Produce json
// @Param code path string true "Package code"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /packages/{code} [get]
func (h *MasterdataHandler) GetPackage(c *gin.Context) {
	code := c.Param("code")
	p, err := h.svc.GetPackageByCode(c.Request.Context(), code)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}

// TrendingHashtags godoc
// @Summary Trending hashtags
// @Tags Hashtags
// @Produce json
// @Param limit query int false "Max tags"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /hashtags/trending [get]
func (h *MasterdataHandler) TrendingHashtags(c *gin.Context) {
	limit := 20
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}
	tags, err := h.svc.TrendingHashtags(c.Request.Context(), limit)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, tags)
}

// SuggestHashtags godoc
// @Summary Suggest hashtags (autocomplete)
// @Tags Hashtags
// @Produce json
// @Param q query string true "Prefix"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /hashtags/suggest [get]
func (h *MasterdataHandler) SuggestHashtags(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.BadRequest(c.Writer, "query parameter 'q' is required")
		return
	}
	tags, err := h.svc.SuggestHashtags(c.Request.Context(), q, 10)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, tags)
}

// Admin category handlers
type CreateCategoryRequest struct {
	Name      string  `json:"name" validate:"required"`
	NameVI    string  `json:"name_vi" validate:"required"`
	Slug      string  `json:"slug" validate:"required"`
	Icon      string  `json:"icon"`
	ParentID  *string `json:"parent_id"`
	SortOrder int     `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name      *string `json:"name"`
	NameVI    *string `json:"name_vi"`
	Slug      *string `json:"slug"`
	Icon      *string `json:"icon"`
	SortOrder *int    `json:"sort_order"`
}

// AdminCreateCategory godoc
// @Summary Admin: create category
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateCategoryRequest true "Category"
// @Success 201 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/categories [post]
func (h *MasterdataHandler) AdminCreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}
	cat := &domain.Category{
		Name:      req.Name,
		NameVI:    req.NameVI,
		Slug:      req.Slug,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
	}
	if req.ParentID != nil {
		oid, err := bson.ObjectIDFromHex(*req.ParentID)
		if err != nil {
			response.BadRequest(c.Writer, "invalid parent_id")
			return
		}
		cat.ParentID = &oid
	}
	if err := h.svc.CreateCategory(c.Request.Context(), cat); err != nil {
		handleError(c, err)
		return
	}
	created(c, cat)
}

// AdminUpdateCategory godoc
// @Summary Admin: update category
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param body body UpdateCategoryRequest true "Fields"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/categories/{id} [put]
func (h *MasterdataHandler) AdminUpdateCategory(c *gin.Context) {
	id := c.Param("id")
	cat, err := h.svc.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	var req UpdateCategoryRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}

	if req.Name != nil {
		cat.Name = *req.Name
	}
	if req.NameVI != nil {
		cat.NameVI = *req.NameVI
	}
	if req.Slug != nil {
		cat.Slug = *req.Slug
	}
	if req.Icon != nil {
		cat.Icon = *req.Icon
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	}

	if err := h.svc.UpdateCategory(c.Request.Context(), cat); err != nil {
		handleError(c, err)
		return
	}
	ok(c, cat)
}

// AdminToggleCategory godoc
// @Summary Admin: toggle category active
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/categories/{id}/toggle [put]
func (h *MasterdataHandler) AdminToggleCategory(c *gin.Context) {
	id := c.Param("id")
	cat, err := h.svc.ToggleCategory(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, cat)
}

// AdminDeleteCategory godoc
// @Summary Admin: delete category
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 204 "No Content"
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/categories/{id} [delete]
func (h *MasterdataHandler) AdminDeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	noContent(c)
}

// AdminReorderCategories godoc
// @Summary Admin: reorder categories
// @Tags Admin
// @Accept json
// @Security BearerAuth
// @Param body body []domain.CategoryOrder true "Ordered ids"
// @Success 204 "No Content"
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/categories/reorder [put]
func (h *MasterdataHandler) AdminReorderCategories(c *gin.Context) {
	var orders []domain.CategoryOrder
	if errs := validator.DecodeAndValidateGin(c, &orders); errs != nil {
		validationResponse(c, errs)
		return
	}
	if err := h.svc.ReorderCategories(c.Request.Context(), orders); err != nil {
		handleError(c, err)
		return
	}
	noContent(c)
}

// Admin package handlers
type CreatePackageRequest struct {
	Code            string `json:"code" validate:"required"`
	Name            string `json:"name" validate:"required"`
	Tier            int    `json:"tier"`
	PriceMonthly    int64  `json:"price_monthly"`
	PriceYearly     int64  `json:"price_yearly"`
	CreditsPerMonth int    `json:"credits_per_month"`
	SearchBoost     bool   `json:"search_boost"`
	SearchPriority  bool   `json:"search_priority"`
	BadgeName       string `json:"badge_name"`
	PrioritySupport bool   `json:"priority_support"`
	FeaturedSlots   int    `json:"featured_slots"`
	SortOrder       int    `json:"sort_order"`
}

// AdminCreatePackage godoc
// @Summary Admin: create package
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreatePackageRequest true "Package"
// @Success 201 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/packages [post]
func (h *MasterdataHandler) AdminCreatePackage(c *gin.Context) {
	var req CreatePackageRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}
	p := &domain.Package{
		Code:            req.Code,
		Name:            req.Name,
		Tier:            req.Tier,
		PriceMonthly:    req.PriceMonthly,
		PriceYearly:     req.PriceYearly,
		CreditsPerMonth: req.CreditsPerMonth,
		SearchBoost:     req.SearchBoost,
		SearchPriority:  req.SearchPriority,
		BadgeName:       req.BadgeName,
		PrioritySupport: req.PrioritySupport,
		FeaturedSlots:   req.FeaturedSlots,
		SortOrder:       req.SortOrder,
		IsActive:        true,
	}
	if err := h.svc.CreatePackage(c.Request.Context(), p); err != nil {
		handleError(c, err)
		return
	}
	created(c, p)
}

// AdminUpdatePackage godoc
// @Summary Admin: update package
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Package ID"
// @Param body body CreatePackageRequest true "Package fields"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/packages/{id} [put]
func (h *MasterdataHandler) AdminUpdatePackage(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.GetPackageByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	var req CreatePackageRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}

	p.Name = req.Name
	p.Tier = req.Tier
	p.PriceMonthly = req.PriceMonthly
	p.PriceYearly = req.PriceYearly
	p.CreditsPerMonth = req.CreditsPerMonth
	p.SearchBoost = req.SearchBoost
	p.SearchPriority = req.SearchPriority
	p.BadgeName = req.BadgeName
	p.PrioritySupport = req.PrioritySupport
	p.FeaturedSlots = req.FeaturedSlots
	p.SortOrder = req.SortOrder

	if err := h.svc.UpdatePackage(c.Request.Context(), p); err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}

// AdminTogglePackage godoc
// @Summary Admin: toggle package active
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "Package ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/packages/{id}/toggle [put]
func (h *MasterdataHandler) AdminTogglePackage(c *gin.Context) {
	id := c.Param("id")
	p, err := h.svc.TogglePackage(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, p)
}
