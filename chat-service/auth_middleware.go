package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)
func AuthMiddleware(jwtHandler JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Set("userID", "")
			c.Set("plan_id", "free")
			c.Set("userRole", "user")
			c.Next()
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtHandler.ValidateAccessToken(token)
		if err != nil {
			c.Set("userID", "")
			c.Set("plan_id", "free")
			c.Set("userRole", "user")
			c.Next()
		}
		if claims == nil {
    		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
    		return
		}


		// Set claims in context
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("plan_id", claims.Plan)
		c.Set("age", claims.Age)
		c.Set("gender", claims.Gender)

		c.Next()
	}
}

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

func ProPlanMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		plan, exists := c.Get("plan_id")
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{"error": "plan not found"})
			return
		}

		if plan != "pro" {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden: insufficient permissions"})
			return
		}

		c.Next()
	}
}

func EnterprisePlanMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		plan, exists := c.Get("plan_id")
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{"error": "plan not found"})
			return
		}

		if plan != "enterprise" {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden: insufficient permissions"})
			return
		}

		c.Next()
	}
}
