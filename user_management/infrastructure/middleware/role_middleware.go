package middleware

import "github.com/gin-gonic/gin"

func RoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{"error": "role not found"})
			return
		}

		if role != "admin" {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden: insufficient permissions"})
			return
		}

		c.Next()
	}
}