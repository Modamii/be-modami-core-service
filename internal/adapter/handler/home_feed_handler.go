package handler

import (
	"github.com/gin-gonic/gin"

	"be-modami-core-service/internal/service"
)

type HomeFeedHandler struct {
	svc *service.HomeFeedService
}

func NewHomeFeedHandler(svc *service.HomeFeedService) *HomeFeedHandler {
	return &HomeFeedHandler{svc: svc}
}

// GetHomeFeed godoc
// @Summary Home feed
// @Description Returns four sections for the home screen: newest products, top-level categories, featured products (near), and latest blog posts.
// @Tags Feed
// @Produce json
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /home-feeds [get]
func (h *HomeFeedHandler) GetHomeFeed(c *gin.Context) {
	feed := h.svc.GetHomeFeed(c.Request.Context())
	ok(c, feed)
}
