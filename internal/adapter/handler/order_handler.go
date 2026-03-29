package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/modami/core-service/internal/adapter/handler/middleware"
	"github.com/modami/core-service/internal/dto"
	"github.com/modami/core-service/internal/service"
	"github.com/modami/core-service/pkg/storage/database/mongodb/pagination"
	"github.com/modami/core-service/pkg/validator"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// CreateOrder godoc
// @Summary Create order (purchase)
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.CreateOrderRequest true "Order payload"
// @Success 201 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}

	userID := middleware.UserID(c)
	o, err := h.svc.CreateOrder(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	created(c, o)
}

// GetByID godoc
// @Summary Get order by ID
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /orders/{id} [get]
func (h *OrderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	o, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, o)
}

// ListByBuyer godoc
// @Summary List orders I purchased
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /orders/my-purchases [get]
func (h *OrderHandler) ListByBuyer(c *gin.Context) {
	userID := middleware.UserID(c)
	cp := pagination.ParseCursor(c.Request)
	orders, nextCursor, err := h.svc.ListByBuyer(c.Request.Context(), userID, cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, orders, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// ListBySeller godoc
// @Summary List orders I sold
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /orders/my-sales [get]
func (h *OrderHandler) ListBySeller(c *gin.Context) {
	userID := middleware.UserID(c)
	cp := pagination.ParseCursor(c.Request)
	orders, nextCursor, err := h.svc.ListBySeller(c.Request.Context(), userID, cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, orders, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// Confirm godoc
// @Summary Seller confirms order
// @Tags Orders
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 409 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /orders/{id}/confirm [put]
func (h *OrderHandler) Confirm(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.UserID(c)
	o, err := h.svc.Confirm(c.Request.Context(), id, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, o)
}

// Ship godoc
// @Summary Seller ships order
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param body body dto.ShipOrderRequest true "Tracking"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /orders/{id}/ship [put]
func (h *OrderHandler) Ship(c *gin.Context) {
	var req dto.ShipOrderRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}

	id := c.Param("id")
	userID := middleware.UserID(c)
	o, err := h.svc.Ship(c.Request.Context(), id, userID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, o)
}

// Receive godoc
// @Summary Buyer confirms receipt
// @Tags Orders
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /orders/{id}/receive [put]
func (h *OrderHandler) Receive(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.UserID(c)
	o, err := h.svc.Receive(c.Request.Context(), id, userID)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, o)
}

// Cancel godoc
// @Summary Cancel order
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param body body dto.CancelOrderRequest true "Reason"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /orders/{id}/cancel [put]
func (h *OrderHandler) Cancel(c *gin.Context) {
	var req dto.CancelOrderRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}

	id := c.Param("id")
	userID := middleware.UserID(c)
	o, err := h.svc.Cancel(c.Request.Context(), id, userID, req)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, o)
}

// ListEvents godoc
// @Summary Order timeline events
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /orders/{id}/events [get]
func (h *OrderHandler) ListEvents(c *gin.Context) {
	id := c.Param("id")
	events, err := h.svc.ListEvents(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, events)
}

// AdminListAll godoc
// @Summary Admin: list all orders
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/orders [get]
func (h *OrderHandler) AdminListAll(c *gin.Context) {
	status := c.Query("status")
	cp := pagination.ParseCursor(c.Request)
	orders, nextCursor, err := h.svc.ListAll(c.Request.Context(), status, cp.Cursor, cp.Limit)
	if err != nil {
		handleError(c, err)
		return
	}
	okWithCursor(c, orders, pagination.CursorMeta{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	})
}

// AdminGetByID godoc
// @Summary Admin: get order
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 404 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/orders/{id} [get]
func (h *OrderHandler) AdminGetByID(c *gin.Context) {
	id := c.Param("id")
	o, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, o)
}

// ForceCancel godoc
// @Summary Admin: force cancel order
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param body body dto.CancelOrderRequest true "Reason"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/orders/{id}/force-cancel [put]
func (h *OrderHandler) ForceCancel(c *gin.Context) {
	var req dto.CancelOrderRequest
	if errs := validator.DecodeAndValidateGin(c, &req); errs != nil {
		validationResponse(c, errs)
		return
	}

	id := c.Param("id")
	adminID := middleware.UserID(c)
	o, err := h.svc.ForceCancel(c.Request.Context(), id, adminID, req.Reason)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, o)
}

// ForceComplete godoc
// @Summary Admin: force complete order
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 401 {object} StandardErrorEnvelope
// @Failure 403 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /admin/orders/{id}/force-complete [put]
func (h *OrderHandler) ForceComplete(c *gin.Context) {
	id := c.Param("id")
	adminID := middleware.UserID(c)
	o, err := h.svc.ForceComplete(c.Request.Context(), id, adminID)
	if err != nil {
		handleError(c, err)
		return
	}
	ok(c, o)
}
