package dto

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
