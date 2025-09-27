package middleware

import (
	infrastructure "lawgen/admin-service/Infrastructure"

	// "log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)


type JWT interface {
    ValidateAccessToken(token string) (*infrastructure.Claims, error)
}


type Claims struct {
	UserID string
	Role   string
	Plan   string
	Age    int
	Gender string
}

// --- Middlewares ---

func AuthMiddleware(jwtHandler JWT) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid auth header"})
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")

        claims, err := jwtHandler.ValidateAccessToken(token)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }

        // Set claims in context
        c.Set("user_id", claims.UserID)
        c.Set("role", claims.Role)
        c.Set("plan", claims.Plan)
        c.Set("age", claims.Age)
        c.Set("gender", claims.Gender)

        // fmt.Printf("[AUTH] User=%s Role=%s Plan=%s Age=%d Gender=%s",
        //     claims.UserID, claims.Role, claims.Plan, claims.Age, claims.Gender)

        c.Next()
    }
}


func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "role not found"})
			return
		}

		if role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
			return
		}

		c.Next()
	}
}

func PlanMiddleware(requiredPlan string) gin.HandlerFunc {
	return func(c *gin.Context) {
		plan, exists := c.Get("plan")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "plan not found"})
			return
		}

		if plan != requiredPlan {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient plan"})
			return
		}

		c.Next()
	}
}

func ProPlanMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    plan, exists := c.Get("plan")
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
    plan, exists := c.Get("plan")
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

