package dto

import "be-modami-core-service/internal/domain"

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

func (r *UpdateCategoryRequest) ApplyTo(c *domain.Category) {
	if r.Name != nil {
		c.Name = *r.Name
	}
	if r.NameVI != nil {
		c.NameVI = *r.NameVI
	}
	if r.Slug != nil {
		c.Slug = *r.Slug
	}
	if r.Icon != nil {
		c.Icon = *r.Icon
	}
	if r.SortOrder != nil {
		c.SortOrder = *r.SortOrder
	}
}
