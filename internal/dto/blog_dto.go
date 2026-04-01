package dto

// AuthorDTO is the request payload for embedded author data.
type AuthorDTO struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Bio   string `json:"bio"`
}

// CreateBlogPostRequest is the payload for creating a new blog post.
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

// UpdateBlogPostRequest is the payload for updating an existing blog post.
// All fields are optional — only non-nil pointer values are applied.
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
