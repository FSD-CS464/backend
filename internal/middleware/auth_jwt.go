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
		var tokenString string
		
		// Try Authorization header first (for API clients)
		h := c.GetHeader("Authorization")
		if strings.HasPrefix(h, "Bearer ") {
			tokenString = strings.TrimPrefix(h, "Bearer ")
		} else {
			// Fallback to cookie (for iframe/CORS scenarios) - Try access_token first then jwt_token as fallback
			tokenString, _ = c.Cookie("access_token")
			if tokenString == "" {
				tokenString, _ = c.Cookie("jwt_token")
			}
		}
		
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		
		claims, err := m.Signer.Parse(tokenString)
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
