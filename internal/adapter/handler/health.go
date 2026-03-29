package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health returns liveness status (no auth).
// @Summary Health check
// @Description Returns service availability for probes
// @Tags System
// @Produce json
// @Success 200 {object} map[string]string "Example: {\"status\":\"ok\"}"
// @Router /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
