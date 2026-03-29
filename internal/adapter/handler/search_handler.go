package handler

import "github.com/gin-gonic/gin"

// SearchHandler exposes search/suggest/trending endpoints (composition of catalog + masterdata).
type SearchHandler struct {
	product    *ProductHandler
	masterdata *MasterdataHandler
}

func NewSearchHandler(product *ProductHandler, masterdata *MasterdataHandler) *SearchHandler {
	return &SearchHandler{product: product, masterdata: masterdata}
}

// Search godoc
// @Summary Search products (alias of GET /products/search)
// @Tags Search
// @Produce json
// @Param q query string false "Search query"
// @Param category_id query string false "Category filter"
// @Param condition query string false "Condition"
// @Param brand query string false "Brand"
// @Param min_price query int false "Min price"
// @Param max_price query int false "Max price"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	h.product.Search(c)
}

// Suggest godoc
// @Summary Hashtag suggestions (alias of GET /hashtags/suggest)
// @Tags Search
// @Produce json
// @Param q query string true "Prefix"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 400 {object} StandardErrorEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /search/suggest [get]
func (h *SearchHandler) Suggest(c *gin.Context) {
	h.masterdata.SuggestHashtags(c)
}

// Trending godoc
// @Summary Trending hashtags (alias of GET /hashtags/trending)
// @Tags Search
// @Produce json
// @Param limit query int false "Max tags"
// @Success 200 {object} StandardSuccessEnvelope
// @Failure 500 {object} StandardErrorEnvelope
// @Router /search/trending [get]
func (h *SearchHandler) Trending(c *gin.Context) {
	h.masterdata.TrendingHashtags(c)
}

// HashtagProducts delegates to product handler (see Products tag).
func (h *SearchHandler) HashtagProducts(c *gin.Context) {
	h.product.HashtagProducts(c)
}
