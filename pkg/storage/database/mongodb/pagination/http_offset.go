package pagination

import (
	"math"
	"net/http"
	"strconv"
)

type OffsetParams struct {
	Page  int
	Limit int
}

type OffsetMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

func ParseOffset(r *http.Request) OffsetParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > MaxLimit {
		limit = DefaultLimit
	}
	return OffsetParams{Page: page, Limit: limit}
}

func (p OffsetParams) Skip() int64 {
	return int64((p.Page - 1) * p.Limit)
}

func NewOffsetMeta(total int64, params OffsetParams) OffsetMeta {
	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))
	return OffsetMeta{
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}
}
