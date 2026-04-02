package dto

import "be-modami-core-service/internal/domain"

type AuthorDTO struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Bio   string `json:"bio"`
}

type CreateBlogPostRequest struct {
	Slug          string    `json:"slug"           validate:"required"`
	SeriesName    string    `json:"series_name"`
	SeriesNo      int       `json:"series_no"`
	SeriesQuarter string    `json:"series_quarter"`
	PostType      string    `json:"post_type"`
	Depth         string    `json:"depth"          validate:"omitempty,oneof=quick deep"`
	Title         string    `json:"title"          validate:"required"`
	Subtitle      string    `json:"subtitle"`
	Body          string    `json:"body"`
	CoverImage    string    `json:"cover_image"`
	CoverCaption  string    `json:"cover_caption"`
	ReadTimeMin   int       `json:"read_time_min"`
	WordCount     int       `json:"word_count"`
	Author        AuthorDTO `json:"author"`
	KeyPoints     []string  `json:"key_points"`
	References    []string  `json:"references"`
	Hashtags      []string  `json:"hashtags"`
	CTALink       string    `json:"cta_link"`
	IsFeatured    bool      `json:"is_featured"`
}

type UpdateBlogPostRequest struct {
	Slug          *string    `json:"slug"`
	SeriesName    *string    `json:"series_name"`
	SeriesNo      *int       `json:"series_no"`
	SeriesQuarter *string    `json:"series_quarter"`
	PostType      *string    `json:"post_type"`
	Depth         *string    `json:"depth"     validate:"omitempty,oneof=quick deep"`
	Title         *string    `json:"title"`
	Subtitle      *string    `json:"subtitle"`
	Body          *string    `json:"body"`
	CoverImage    *string    `json:"cover_image"`
	CoverCaption  *string    `json:"cover_caption"`
	ReadTimeMin   *int       `json:"read_time_min"`
	WordCount     *int       `json:"word_count"`
	Author        *AuthorDTO `json:"author"`
	KeyPoints     []string   `json:"key_points"`
	References    []string   `json:"references"`
	Hashtags      []string   `json:"hashtags"`
	CTALink       *string    `json:"cta_link"`
	IsFeatured    *bool      `json:"is_featured"`
}

// ApplyTo patches p with every non-nil field in the request.
func (r *UpdateBlogPostRequest) ApplyTo(p *domain.BlogPost) {
	if r.Slug != nil {
		p.Slug = *r.Slug
	}
	if r.SeriesName != nil {
		p.SeriesName = *r.SeriesName
	}
	if r.SeriesNo != nil {
		p.SeriesNo = *r.SeriesNo
	}
	if r.SeriesQuarter != nil {
		p.SeriesQuarter = *r.SeriesQuarter
	}
	if r.PostType != nil {
		p.PostType = *r.PostType
	}
	if r.Depth != nil {
		p.Depth = domain.PostDepth(*r.Depth)
	}
	if r.Title != nil {
		p.Title = *r.Title
	}
	if r.Subtitle != nil {
		p.Subtitle = *r.Subtitle
	}
	if r.Body != nil {
		p.Body = *r.Body
	}
	if r.CoverImage != nil {
		p.CoverImage = *r.CoverImage
	}
	if r.CoverCaption != nil {
		p.CoverCaption = *r.CoverCaption
	}
	if r.ReadTimeMin != nil {
		p.ReadTimeMin = *r.ReadTimeMin
	}
	if r.WordCount != nil {
		p.WordCount = *r.WordCount
	}
	if r.Author != nil {
		p.Author = domain.BlogAuthor{
			Name:  r.Author.Name,
			Title: r.Author.Title,
			Bio:   r.Author.Bio,
		}
	}
	if r.KeyPoints != nil {
		p.KeyPoints = r.KeyPoints
	}
	if r.References != nil {
		p.References = r.References
	}
	if r.Hashtags != nil {
		p.Hashtags = r.Hashtags
	}
	if r.CTALink != nil {
		p.CTALink = *r.CTALink
	}
	if r.IsFeatured != nil {
		p.IsFeatured = *r.IsFeatured
	}
}
