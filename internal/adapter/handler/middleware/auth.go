package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	apperror "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
	"gitlab.com/lifegoeson-libs/pkg-gokit/response"
)

var errUnauthorized = apperror.New(apperror.CodeUnauthorized, "authentication required")

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

type Auth struct {
	secret []byte
}

func NewAuth(secret string) *Auth {
	return &Auth{secret: []byte(secret)}
}

func (a *Auth) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr == header {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return a.secret, nil
		})
		if err != nil || !token.Valid {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}

		userID, _ := claims["user_id"].(string)
		role, _ := claims["role"].(string)
		if userID == "" {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}
		if role == "" {
			role = "user"
		}

		c.Set(string(UserIDKey), userID)
		c.Set(string(RoleKey), role)
		c.Next()
	}
}

// Optional extracts user info if token is present but does not require it.
func (a *Auth) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.Next()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr == header {
			c.Next()
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return a.secret, nil
		})
		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Next()
			return
		}

		userID, _ := claims["user_id"].(string)
		role, _ := claims["role"].(string)
		if userID != "" {
			c.Set(string(UserIDKey), userID)
			c.Set(string(RoleKey), role)
		}
		c.Next()
	}
}

// UserID returns the authenticated user id from the Gin context.
func UserID(c *gin.Context) string {
	v, exists := c.Get(string(UserIDKey))
	if !exists {
		return ""
	}
	s, _ := v.(string)
	return s
}

// Role returns the role from the Gin context.
func Role(c *gin.Context) string {
	v, exists := c.Get(string(RoleKey))
	if !exists {
		return ""
	}
	s, _ := v.(string)
	return s
}
