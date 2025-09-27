package route

import (
	// "g3-g65-bsp/delivery/controller"
	// "g3-g65-bsp/infrastructure/cache"
	// "g3-g65-bsp/infrastructure/auth"
	// "g3-g65-bsp/infrastructure/middleware"
	// "time"

	// "github.com/didip/tollbooth/v7/limiter"
	// "github.com/didip/tollbooth_gin"

	"github.com/gin-gonic/gin"
)


// HealthRouter registers a health check endpoint
func HealthRouter(r *gin.Engine) {
    r.GET("/health", func(ctx *gin.Context) {
        ctx.JSON(200, gin.H{"status": "ok"})
    })
}


// NewRouter initializes the Gin engine and registers all routes
func NewRouter() *gin.Engine {
	r := gin.Default()
    HealthRouter(r) // Register health check endpoint
	return r
}
