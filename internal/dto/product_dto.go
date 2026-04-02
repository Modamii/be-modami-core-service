package dto

import "be-modami-core-service/internal/domain"

type CreateProductRequest struct {
	Title       string       `json:"title" validate:"required,min=5,max=200"`
	Description string       `json:"description" validate:"required,min=20,max=5000"`
	Price       int64        `json:"price" validate:"required,min=1000"`
	CategoryID  string       `json:"category_id" validate:"required"`
	Condition   string       `json:"condition" validate:"required,oneof=new like_new good fair"`
	Size        string       `json:"size" validate:"required"`
	Brand       string       `json:"brand"`
	Color       string       `json:"color"`
	Material    string       `json:"material"`
	Images      []ImageInput `json:"images" validate:"required,min=1,max=6,dive"`
	Hashtags    []string     `json:"hashtags" validate:"max=10"`
	CreditCost  int          `json:"credit_cost"`
}

type ImageInput struct {
	URL      string `json:"url" validate:"required,url"`
	Position int    `json:"position"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type UpdateProductRequest struct {
	Title       *string      `json:"title" validate:"omitempty,min=5,max=200"`
	Description *string      `json:"description" validate:"omitempty,min=20,max=5000"`
	Price       *int64       `json:"price" validate:"omitempty,min=1000"`
	CategoryID  *string      `json:"category_id"`
	Condition   *string      `json:"condition" validate:"omitempty,oneof=new like_new good fair"`
	Size        *string      `json:"size"`
	Brand       *string      `json:"brand"`
	Color       *string      `json:"color"`
	Material    *string      `json:"material"`
	Images      []ImageInput `json:"images" validate:"omitempty,min=1,max=6,dive"`
	Hashtags    []string     `json:"hashtags" validate:"omitempty,max=10"`
	CreditCost  *int         `json:"credit_cost"`
}

func (r *UpdateProductRequest) ApplyTo(p *domain.Product) {
	if r.Title != nil {
		p.Title = *r.Title
	}
	if r.Description != nil {
		p.Description = *r.Description
	}
	if r.Price != nil {
		p.Price = *r.Price
	}
	if r.Condition != nil {
		p.Condition = *r.Condition
	}
	if r.Size != nil {
		p.Size = *r.Size
	}
	if r.Brand != nil {
		p.Brand = *r.Brand
	}
	if r.Color != nil {
		p.Color = *r.Color
	}
	if r.Material != nil {
		p.Material = *r.Material
	}
	if r.Images != nil {
		p.Images = toProductImages(r.Images)
	}
	if r.Hashtags != nil {
		p.Hashtags = r.Hashtags
	}
	if r.CreditCost != nil {
		p.CreditCost = *r.CreditCost
	}
}

type SubmitRequest struct {
	Note string `json:"note"`
}

type ResubmitRequest struct {
	Title       *string      `json:"title" validate:"omitempty,min=5,max=200"`
	Description *string      `json:"description" validate:"omitempty,min=20,max=5000"`
	Price       *int64       `json:"price" validate:"omitempty,min=1000"`
	CategoryID  *string      `json:"category_id"`
	Condition   *string      `json:"condition" validate:"omitempty,oneof=new like_new good fair"`
	Size        *string      `json:"size"`
	Brand       *string      `json:"brand"`
	Color       *string      `json:"color"`
	Material    *string      `json:"material"`
	Images      []ImageInput `json:"images" validate:"omitempty,min=1,max=6,dive"`
	Hashtags    []string     `json:"hashtags" validate:"omitempty,max=10"`
	Note        string       `json:"note"`
}

func (r *ResubmitRequest) ApplyTo(p *domain.Product) {
	if r.Title != nil {
		p.Title = *r.Title
	}
	if r.Description != nil {
		p.Description = *r.Description
	}
	if r.Price != nil {
		p.Price = *r.Price
	}
	if r.Condition != nil {
		p.Condition = *r.Condition
	}
	if r.Size != nil {
		p.Size = *r.Size
	}
	if r.Brand != nil {
		p.Brand = *r.Brand
	}
	if r.Color != nil {
		p.Color = *r.Color
	}
	if r.Material != nil {
		p.Material = *r.Material
	}
	if r.Images != nil {
		p.Images = toProductImages(r.Images)
	}
	if r.Hashtags != nil {
		p.Hashtags = r.Hashtags
	}
}

func toProductImages(inputs []ImageInput) []domain.ProductImage {
	images := make([]domain.ProductImage, len(inputs))
	for i, img := range inputs {
		images[i] = domain.ProductImage{
			URL:      img.URL,
			Position: img.Position,
			Width:    img.Width,
			Height:   img.Height,
		}
	}
	return images
}
