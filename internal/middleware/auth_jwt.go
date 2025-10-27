package middleware

import (
	"net/http"
	"strings"

	"fsd-backend/internal/auth"
	"github.com/gin-gonic/gin"
)

const CtxUserID = "uid"

type JWTMiddleware struct {
	Signer *auth.Signer
}

func NewJWT(s *auth.Signer) *JWTMiddleware {
	return &JWTMiddleware{Signer: s}
}

func (m *JWTMiddleware) Require() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		raw := strings.TrimPrefix(h, "Bearer ")
		claims, err := m.Signer.Parse(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Next()
	}
}

// Helper for controllers (avoid import loops)
func UserID(c *gin.Context) string {
	if v, ok := c.Get(CtxUserID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
