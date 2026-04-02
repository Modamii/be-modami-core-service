package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"be-modami-core-service/internal/domain"
	"be-modami-core-service/internal/dto"
	"be-modami-core-service/internal/service"
	"be-modami-core-service/pkg/validator"

	"gitlab.com/lifegoeson-libs/pkg-gokit/response"
	"go.mongodb.org/mongo-driver/v2/bson"
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
		response.BadRequest(c.Writer, "tham số 'q' là bắt buộc")
		return
	}
	tags, err := h.svc.SuggestHashtags(c.Request.Context(), q, 10)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, tags)
}

// AdminCreateCategory godoc
// @Summary Admin: create category
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.CreateCategoryRequest true "Category"
// @Success 201 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /categories [post]
func (h *MasterdataHandler) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
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
			response.BadRequest(c.Writer, "parent_id không hợp lệ")
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
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param body body dto.UpdateCategoryRequest true "Fields"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /categories/{id} [put]
func (h *MasterdataHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	cat, err := h.svc.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	var req dto.UpdateCategoryRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}

	req.ApplyTo(cat)

	if err := h.svc.UpdateCategory(c.Request.Context(), cat); err != nil {
		handleError(c, err)
		return
	}
	ok(c, cat)
}

// AdminToggleCategory godoc
// @Summary Admin: toggle category active
// @Tags Categories
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /categories/{id}/toggle [put]
func (h *MasterdataHandler) ToggleCategory(c *gin.Context) {
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
// @Tags Categories
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 204 "No Content"
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /categories/{id} [delete]
func (h *MasterdataHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	noContent(c)
}

// AdminReorderCategories godoc
// @Summary Admin: reorder categories
// @Tags Categories
// @Accept json
// @Security BearerAuth
// @Param body body []domain.CategoryOrder true "Ordered ids"
// @Success 204 "No Content"
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /categories/reorder [put]
func (h *MasterdataHandler) ReorderCategories(c *gin.Context) {
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
