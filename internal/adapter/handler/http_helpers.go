package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"gitlab.com/lifegoeson-libs/pkg-gokit/mongodb/pagination"
	"be-modami-core-service/pkg/validator"

	"github.com/gin-gonic/gin"
	gokit "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
	"gitlab.com/lifegoeson-libs/pkg-gokit/response"
)

func handleError(c *gin.Context, err error) {
	if ae := gokit.AsAppError(err); ae != nil {
		response.Err(c.Writer, ae)
		return
	}
	response.InternalError(c.Writer, "lỗi hệ thống nội bộ")
}

func validationResponse(c *gin.Context, errs []validator.FieldError) {
	fields := make([]response.FieldError, 0, len(errs))
	for _, e := range errs {
		fields = append(fields, response.FieldError{
			Field:   e.Field,
			Message: e.Message,
		})
	}
	response.ValidationErr(c.Writer, fields)
}

func ok(c *gin.Context, data any) {
	response.OK(c.Writer, data)
}

func created(c *gin.Context, data any) {
	response.Created(c.Writer, data)
}

func noContent(c *gin.Context) {
	response.NoContent(c.Writer)
}

// okWithCursor matches the previous cursor list shape: success envelope + meta.next_cursor / has_more.
func okWithCursor(c *gin.Context, data any, meta pagination.CursorMeta) {
	type metaBlock struct {
		Timestamp  int64  `json:"timestamp"`
		NextCursor string `json:"next_cursor,omitempty"`
		HasMore    bool   `json:"has_more"`
	}
	out := struct {
		Success bool      `json:"success"`
		Data    any       `json:"data"`
		Meta    metaBlock `json:"meta"`
	}{
		Success: true,
		Data:    data,
		Meta: metaBlock{
			Timestamp:  time.Now().Unix(),
			NextCursor: meta.NextCursor,
			HasMore:    meta.HasMore,
		},
	}
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(c.Writer).Encode(out)
}
