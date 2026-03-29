package service

import (
	"context"

	"github.com/modami/core-service/internal/port"
	apperror "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
)

type orderProductReader struct {
	products *ProductService
}

// NewOrderProductReader adapts the product catalog for order creation snapshots.
func NewOrderProductReader(products *ProductService) port.OrderProductReader {
	return &orderProductReader{products: products}
}

func (r *orderProductReader) GetByIDForOrder(ctx context.Context, id string) (port.OrderProductSnapshot, error) {
	p, err := r.products.GetByID(ctx, id)
	if err != nil {
		return port.OrderProductSnapshot{}, err
	}
	if p == nil {
		return port.OrderProductSnapshot{}, apperror.New(apperror.CodeNotFound,"product not found")
	}
	imgURL := ""
	if len(p.Images) > 0 {
		imgURL = p.Images[0].URL
	}
	return port.OrderProductSnapshot{
		ID:        p.ID,
		SellerID:  p.SellerID,
		Title:     p.Title,
		ImageURL:  imgURL,
		Brand:     p.Brand,
		Condition: p.Condition,
		Size:      p.Size,
		Category:  p.CategoryID.Hex(),
		Price:     p.Price,
	}, nil
}
