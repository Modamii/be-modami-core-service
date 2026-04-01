package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	gokit "gitlab.com/lifegoeson-libs/pkg-gokit/apperror"
	"gitlab.com/lifegoeson-libs/pkg-gokit/response"
)

const (
	ctxKeyUserID      = "user_id"
	ctxKeyRole        = "user_role"
	ctxKeyPermissions = "user_permissions"
)

var (
	errUnauthorized = gokit.New(gokit.CodeUnauthorized, "chưa xác thực")
	errForbidden    = gokit.New(gokit.CodeForbidden, "không có quyền truy cập")
)

// Auth validates Keycloak-issued JWTs via JWKS.
type Auth struct {
	jwksURL string
	cache   *jwksCache
}

// NewAuth creates an Auth middleware. If jwksURL is empty, tokens are parsed
// without signature verification (dev/test mode).
func NewAuth(jwksURL string) *Auth {
	a := &Auth{
		jwksURL: jwksURL,
		cache:   &jwksCache{keys: make(map[string]*rsa.PublicKey)},
	}
	if jwksURL != "" {
		_ = a.cache.refresh(jwksURL) // best-effort on startup
	}
	return a
}

// Required returns a gin middleware that enforces a valid Bearer token.
func (a *Auth) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}
		parts := strings.SplitN(raw, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}
		tokenStr := parts[1]

		claims, err := a.parse(tokenStr)
		if err != nil {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}

		sub, _ := claims["sub"].(string)
		if sub == "" {
			response.Err(c.Writer, errUnauthorized)
			c.Abort()
			return
		}
		c.Set(ctxKeyUserID, sub)

		// Extract first role from realm_access.roles
		if ra, ok := claims["realm_access"].(map[string]interface{}); ok {
			if rolesRaw, ok := ra["roles"].([]interface{}); ok && len(rolesRaw) > 0 {
				for _, r := range rolesRaw {
					if rs, ok := r.(string); ok && rs != "" {
						c.Set(ctxKeyRole, strings.ToLower(rs))
						break
					}
				}
			}
		}

		// Extract resource_access permissions for the client (azp claim)
		if clientID, ok := claims["azp"].(string); ok && clientID != "" {
			if ra, ok := claims["resource_access"].(map[string]interface{}); ok {
				if client, ok := ra[clientID].(map[string]interface{}); ok {
					if rolesRaw, ok := client["roles"].([]interface{}); ok {
						perms := make([]string, 0, len(rolesRaw))
						for _, r := range rolesRaw {
							if rs, ok := r.(string); ok && rs != "" {
								perms = append(perms, rs)
							}
						}
						c.Set(ctxKeyPermissions, perms)
					}
				}
			}
		}

		c.Next()
	}
}

// UserID returns the authenticated user's ID (JWT sub claim) from context.
func UserID(c *gin.Context) string {
	val, _ := c.Get(ctxKeyUserID)
	id, _ := val.(string)
	return id
}

// Role returns the first realm role of the authenticated user (lowercased).
func Role(c *gin.Context) string {
	val, _ := c.Get(ctxKeyRole)
	role, _ := val.(string)
	return role
}

// Permissions returns the resource_access permissions for the authenticated user.
func Permissions(c *gin.Context) []string {
	val, _ := c.Get(ctxKeyPermissions)
	perms, _ := val.([]string)
	return perms
}

// HasPermission reports whether the authenticated user holds the given permission.
func HasPermission(c *gin.Context, permission string) bool {
	for _, p := range Permissions(c) {
		if p == permission {
			return true
		}
	}
	return false
}

// RequirePermission returns a middleware that allows the request only when
// the authenticated user holds the specified resource_access permission.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasPermission(c, permission) {
			response.Err(c.Writer, errForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// parse validates and extracts claims from a JWT token string.
func (a *Auth) parse(tokenStr string) (jwt.MapClaims, error) {
	if a.jwksURL == "" {
		// Dev mode: no signature verification
		token, _, err := jwt.NewParser().ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			return nil, err
		}
		return token.Claims.(jwt.MapClaims), nil
	}

	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		key, err := a.cache.get(kid, a.jwksURL)
		if err != nil {
			return nil, err
		}
		return key, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token không hợp lệ")
	}
	return token.Claims.(jwt.MapClaims), nil
}

// ---------------------------------------------------------------------------
// JWKS cache
// ---------------------------------------------------------------------------

type jwksCache struct {
	mu          sync.RWMutex
	keys        map[string]*rsa.PublicKey
	lastRefresh time.Time
}

func (c *jwksCache) get(kid string, url string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	c.mu.RUnlock()
	if ok {
		return key, nil
	}
	// Unknown kid — refresh and retry once
	if err := c.refresh(url); err != nil {
		return nil, err
	}
	c.mu.RLock()
	key, ok = c.keys[kid]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown key id: %s", kid)
	}
	return key, nil
}

func (c *jwksCache) refresh(url string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := parseRSAPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		c.keys[k.Kid] = pub
	}
	c.lastRefresh = time.Now()
	return nil
}

func parseRSAPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)
	return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
}
