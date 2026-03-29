package dto

type CreateOrderRequest struct {
	ProductID string              `json:"product_id" validate:"required"`
	Shipping  CreateShippingInput `json:"shipping" validate:"required"`
}

type CreateShippingInput struct {
	Name     string `json:"name" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	Address  string `json:"address" validate:"required"`
	Province string `json:"province" validate:"required"`
	District string `json:"district" validate:"required"`
	Ward     string `json:"ward" validate:"required"`
}

type ShipOrderRequest struct {
	TrackingCode     string `json:"tracking_code" validate:"required"`
	ShippingProvider string `json:"shipping_provider" validate:"required"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason" validate:"required"`
}
