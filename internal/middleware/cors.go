package middleware

import "github.com/gin-gonic/gin"

func CORS(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin matches allowed origin
		if origin != "" {
			if allowedOrigin == "*" {
				// Allow all origins (but cannot use credentials)
				c.Header("Access-Control-Allow-Origin", "*")
			} else if origin == allowedOrigin {
				// Allow specific origin
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
			} else {
				// Reject if origin doesn't match
				c.AbortWithStatus(403)
				return
			}
		} else {
			// No origin header (same-origin request) - allow it
			if allowedOrigin == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", allowedOrigin)
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
