package middleware

import (
	"github.com/gin-gonic/gin"
	apperror "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
	"gitlab.com/lifegoeson-libs/pkg-gokit/response"
)

var errForbidden = apperror.New(apperror.CodeForbidden, "access denied")

func AdminOnly(c *gin.Context) {
	if Role(c) != "admin" {
		response.Err(c.Writer, errForbidden)
		c.Abort()
		return
	}
	c.Next()
}
