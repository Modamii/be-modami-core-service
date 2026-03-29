package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/modami/core-service/internal/domain"
	"github.com/modami/core-service/internal/dto"
	"github.com/modami/core-service/internal/port"
	apperror "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
)

type OrderService struct {
	repo    port.OrderRepository
	product port.OrderProductReader
}

func NewOrderService(repo port.OrderRepository, product port.OrderProductReader) *OrderService {
	return &OrderService{repo: repo, product: product}
}

func (s *OrderService) CreateOrder(ctx context.Context, buyerID string, req dto.CreateOrderRequest) (*domain.Order, error) {
	bid, err := bson.ObjectIDFromHex(buyerID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid buyer_id")
	}

	pd, err := s.product.GetByIDForOrder(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	if pd.SellerID == bid {
		return nil, apperror.New(apperror.CodeBadRequest,"cannot purchase your own product")
	}

	orderCode := fmt.Sprintf("MD%d", time.Now().UnixMilli())

	platformFee := pd.Price * 5 / 100
	totalPrice := pd.Price + platformFee

	o := &domain.Order{
		OrderCode: orderCode,
		BuyerID:   bid,
		SellerID:  pd.SellerID,
		ProductID: pd.ID,
		Snapshot: domain.OrderSnapshot{
			Title:     pd.Title,
			ImageURL:  pd.ImageURL,
			Brand:     pd.Brand,
			Condition: pd.Condition,
			Size:      pd.Size,
			Category:  pd.Category,
		},
		ItemPrice:   pd.Price,
		ShippingFee: 0,
		PlatformFee: platformFee,
		TotalPrice:  totalPrice,
		Shipping: domain.ShippingInfo{
			Name:     req.Shipping.Name,
			Phone:    req.Shipping.Phone,
			Address:  req.Shipping.Address,
			Province: req.Shipping.Province,
			District: req.Shipping.District,
			Ward:     req.Shipping.Ward,
		},
		Status: domain.StatusCreated,
	}

	if err := s.repo.Create(ctx, o); err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to create order")
	}

	_ = s.repo.CreateEvent(ctx, &domain.OrderEvent{
		OrderID:    o.ID,
		FromStatus: "",
		ToStatus:   string(domain.StatusCreated),
		ActorID:    bid,
		ActorType:  "buyer",
		Note:       "order created",
	})

	return o, nil
}

func (s *OrderService) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid order id")
	}
	o, err := s.repo.GetByID(ctx, oid)
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal,"failed to get order")
	}
	if o == nil {
		return nil, apperror.New(apperror.CodeNotFound,"order not found")
	}
	return o, nil
}

func (s *OrderService) Confirm(ctx context.Context, id string, sellerID string) (*domain.Order, error) {
	o, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	sid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid seller_id")
	}
	if o.SellerID != sid {
		return nil, apperror.New(apperror.CodeForbidden,"only the seller can confirm this order")
	}
	if !domain.ValidOrderTransition(o.Status, domain.StatusConfirmed) {
		return nil, apperror.New(apperror.CodeBadRequest,"order cannot be confirmed in current status")
	}

	fromStatus := o.Status
	o.Status = domain.StatusConfirmed
	now := time.Now()
	o.ConfirmedAt = &now

	if err := s.repo.Update(ctx, o); err != nil {
		if err == domain.ErrOrderVersionConflict {
			return nil, apperror.New(apperror.CodeConflict,"order was modified by another request, please retry")
		}
		return nil, apperror.New(apperror.CodeInternal,"failed to confirm order")
	}

	_ = s.repo.CreateEvent(ctx, &domain.OrderEvent{
		OrderID:    o.ID,
		FromStatus: string(fromStatus),
		ToStatus:   string(domain.StatusConfirmed),
		ActorID:    sid,
		ActorType:  "seller",
		Note:       "seller confirmed order",
	})

	return o, nil
}

func (s *OrderService) Ship(ctx context.Context, id string, sellerID string, req dto.ShipOrderRequest) (*domain.Order, error) {
	o, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	sid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid seller_id")
	}
	if o.SellerID != sid {
		return nil, apperror.New(apperror.CodeForbidden,"only the seller can ship this order")
	}
	if !domain.ValidOrderTransition(o.Status, domain.StatusShipped) {
		return nil, apperror.New(apperror.CodeBadRequest,"order cannot be shipped in current status")
	}

	fromStatus := o.Status
	o.Status = domain.StatusShipped
	o.TrackingCode = req.TrackingCode
	o.ShippingProvider = req.ShippingProvider
	now := time.Now()
	o.ShippedAt = &now

	if err := s.repo.Update(ctx, o); err != nil {
		if err == domain.ErrOrderVersionConflict {
			return nil, apperror.New(apperror.CodeConflict,"order was modified by another request, please retry")
		}
		return nil, apperror.New(apperror.CodeInternal,"failed to ship order")
	}

	_ = s.repo.CreateEvent(ctx, &domain.OrderEvent{
		OrderID:    o.ID,
		FromStatus: string(fromStatus),
		ToStatus:   string(domain.StatusShipped),
		ActorID:    sid,
		ActorType:  "seller",
		Note:       "seller shipped order",
	})

	return o, nil
}

func (s *OrderService) Receive(ctx context.Context, id string, buyerID string) (*domain.Order, error) {
	o, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	bid, err := bson.ObjectIDFromHex(buyerID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid buyer_id")
	}
	if o.BuyerID != bid {
		return nil, apperror.New(apperror.CodeForbidden,"only the buyer can confirm delivery")
	}
	if !domain.ValidOrderTransition(o.Status, domain.StatusDelivered) {
		return nil, apperror.New(apperror.CodeBadRequest,"order cannot be marked as delivered in current status")
	}

	fromStatus := o.Status
	o.Status = domain.StatusDelivered
	now := time.Now()
	o.DeliveredAt = &now

	if err := s.repo.Update(ctx, o); err != nil {
		if err == domain.ErrOrderVersionConflict {
			return nil, apperror.New(apperror.CodeConflict,"order was modified by another request, please retry")
		}
		return nil, apperror.New(apperror.CodeInternal,"failed to receive order")
	}

	_ = s.repo.CreateEvent(ctx, &domain.OrderEvent{
		OrderID:    o.ID,
		FromStatus: string(fromStatus),
		ToStatus:   string(domain.StatusDelivered),
		ActorID:    bid,
		ActorType:  "buyer",
		Note:       "buyer confirmed delivery",
	})

	return o, nil
}

func (s *OrderService) Cancel(ctx context.Context, id string, userID string, req dto.CancelOrderRequest) (*domain.Order, error) {
	o, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid user_id")
	}

	var actorType string
	switch {
	case o.BuyerID == uid:
		if o.Status != domain.StatusCreated && o.Status != domain.StatusConfirmed {
			return nil, apperror.New(apperror.CodeBadRequest,"buyer can only cancel orders in created or confirmed status")
		}
		actorType = "buyer"
	case o.SellerID == uid:
		if o.Status != domain.StatusCreated {
			return nil, apperror.New(apperror.CodeBadRequest,"seller can only cancel orders in created status")
		}
		actorType = "seller"
	default:
		return nil, apperror.New(apperror.CodeForbidden,"you are not a participant of this order")
	}

	if !domain.ValidOrderTransition(o.Status, domain.StatusCancelled) {
		return nil, apperror.New(apperror.CodeBadRequest,"order cannot be cancelled in current status")
	}

	fromStatus := o.Status
	o.Status = domain.StatusCancelled
	o.CancelReason = req.Reason
	o.CancelledBy = actorType
	now := time.Now()
	o.CancelledAt = &now

	if err := s.repo.Update(ctx, o); err != nil {
		if err == domain.ErrOrderVersionConflict {
			return nil, apperror.New(apperror.CodeConflict,"order was modified by another request, please retry")
		}
		return nil, apperror.New(apperror.CodeInternal,"failed to cancel order")
	}

	_ = s.repo.CreateEvent(ctx, &domain.OrderEvent{
		OrderID:    o.ID,
		FromStatus: string(fromStatus),
		ToStatus:   string(domain.StatusCancelled),
		ActorID:    uid,
		ActorType:  actorType,
		Note:       req.Reason,
	})

	return o, nil
}

func (s *OrderService) ForceCancel(ctx context.Context, id string, adminID string, reason string) (*domain.Order, error) {
	o, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	aid, err := bson.ObjectIDFromHex(adminID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid admin_id")
	}

	if o.Status == domain.StatusCancelled || o.Status == domain.StatusCompleted {
		return nil, apperror.New(apperror.CodeBadRequest,"order is already in a terminal status")
	}

	fromStatus := o.Status
	o.Status = domain.StatusCancelled
	o.CancelReason = reason
	o.CancelledBy = "admin"
	now := time.Now()
	o.CancelledAt = &now

	if err := s.repo.Update(ctx, o); err != nil {
		if err == domain.ErrOrderVersionConflict {
			return nil, apperror.New(apperror.CodeConflict,"order was modified by another request, please retry")
		}
		return nil, apperror.New(apperror.CodeInternal,"failed to force cancel order")
	}

	_ = s.repo.CreateEvent(ctx, &domain.OrderEvent{
		OrderID:    o.ID,
		FromStatus: string(fromStatus),
		ToStatus:   string(domain.StatusCancelled),
		ActorID:    aid,
		ActorType:  "admin",
		Note:       reason,
	})

	return o, nil
}

func (s *OrderService) ForceComplete(ctx context.Context, id string, adminID string) (*domain.Order, error) {
	o, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	aid, err := bson.ObjectIDFromHex(adminID)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid admin_id")
	}

	if o.Status == domain.StatusCancelled || o.Status == domain.StatusCompleted {
		return nil, apperror.New(apperror.CodeBadRequest,"order is already in a terminal status")
	}

	fromStatus := o.Status
	o.Status = domain.StatusCompleted

	if err := s.repo.Update(ctx, o); err != nil {
		if err == domain.ErrOrderVersionConflict {
			return nil, apperror.New(apperror.CodeConflict,"order was modified by another request, please retry")
		}
		return nil, apperror.New(apperror.CodeInternal,"failed to force complete order")
	}

	_ = s.repo.CreateEvent(ctx, &domain.OrderEvent{
		OrderID:    o.ID,
		FromStatus: string(fromStatus),
		ToStatus:   string(domain.StatusCompleted),
		ActorID:    aid,
		ActorType:  "admin",
		Note:       "admin force completed order",
	})

	return o, nil
}

func (s *OrderService) ListByBuyer(ctx context.Context, buyerID string, cursor string, limit int) ([]domain.Order, string, error) {
	bid, err := bson.ObjectIDFromHex(buyerID)
	if err != nil {
		return nil, "", apperror.New(apperror.CodeBadRequest,"invalid buyer_id")
	}
	return s.repo.ListByBuyer(ctx, bid, cursor, limit)
}

func (s *OrderService) ListBySeller(ctx context.Context, sellerID string, cursor string, limit int) ([]domain.Order, string, error) {
	sid, err := bson.ObjectIDFromHex(sellerID)
	if err != nil {
		return nil, "", apperror.New(apperror.CodeBadRequest,"invalid seller_id")
	}
	return s.repo.ListBySeller(ctx, sid, cursor, limit)
}

func (s *OrderService) ListAll(ctx context.Context, status string, cursor string, limit int) ([]domain.Order, string, error) {
	return s.repo.ListAll(ctx, status, cursor, limit)
}

func (s *OrderService) ListEvents(ctx context.Context, id string) ([]domain.OrderEvent, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, apperror.New(apperror.CodeBadRequest,"invalid order id")
	}
	return s.repo.ListEvents(ctx, oid)
}
