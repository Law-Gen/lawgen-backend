package security

import (
	"net/http"
	"strings"

	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/domain"
	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	HeaderUserID = "X-User-Id"
	HeaderRoles  = "X-User-Roles" // comma-separated

	CtxUserIDKey = "user_id"
	CtxRolesKey  = "roles"
)

func GatewayAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader(HeaderUserID)
		if userID == "" {
			response.WriteUnauthorized(c.Writer, domain.ErrUnauthorized, "Missing identity headers")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		rolesHdr := c.GetHeader(HeaderRoles)
		var roles []string
		if rolesHdr != "" {
			for _, r := range strings.Split(rolesHdr, ",") {
				if s := strings.TrimSpace(r); s != "" {
					roles = append(roles, s)
				}
			}
		}
		c.Set(CtxUserIDKey, userID)
		c.Set(CtxRolesKey, roles)
		c.Next()
	}
}

func RequireRoles(required ...string) gin.HandlerFunc {
	req := map[string]struct{}{}
	for _, r := range required {
		req[r] = struct{}{}
	}
	return func(c *gin.Context) {
		val, ok := c.Get(CtxRolesKey)
		if !ok {
			response.WriteForbidden(c.Writer, domain.ErrAccessDenied, "Forbidden")
			c.Abort()
			return
		}
		roles, _ := val.([]string)
		for _, r := range roles {
			if _, ok := req[r]; ok {
				c.Next()
				return
			}
		}
		response.WriteForbidden(c.Writer, domain.ErrAccessDenied, "Forbidden")
		c.Abort()
	}
}

func UserID(c *gin.Context) string {
	if v, ok := c.Get(CtxUserIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}