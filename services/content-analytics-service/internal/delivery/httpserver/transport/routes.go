package transport

import (
	"net/http"
	"strconv"

	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/delivery/httpserver/transport/security"
	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/domain"
	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/infrastructure/config"
	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/usecase"
	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, logger *zap.Logger,
	content ContentPort, feedback FeedbackPort, legal LegalEntityPort, analytics AnalyticsPort) {

	// Health
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/readyz", func(c *gin.Context) { c.String(http.StatusOK, "ready") })

	v1 := r.Group("/v1")

	// Content (your area)
	v1.POST("/contents", security.GatewayAuth(), createContent(content))
	v1.PUT("/contents/:id", security.GatewayAuth(), updateContent(content))
	v1.GET("/contents/:id", getContentByID(content))
	v1.GET("/contents", searchContent(content))

	// Feedback (your area)
	v1.POST("/feedback", security.GatewayAuth(), submitFeedback(feedback))

	// Legal entity (teammate area)
	v1.POST("/legal-entities", security.GatewayAuth(), createLegalEntity(legal))
	v1.PUT("/legal-entities/:id", security.GatewayAuth(), updateLegalEntity(legal))
	v1.GET("/legal-entities/:id", getLegalEntityByID(legal))
	v1.GET("/legal-entities", searchLegalEntities(legal))

	// Analytics (teammate area)
	v1.GET("/analytics/query-trends", security.GatewayAuth(), security.RequireRoles("enterprise_user"), queryTrends(analytics))
}

// ---------- Content handlers ----------
func createContent(uc ContentPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req domain.Content
		if err := response.DecodeJSON(c.Request.Body, &req); err != nil {
			response.WriteBadRequest(c.Writer, domain.ErrInvalidInput, "Invalid JSON", map[string]string{"body": "malformed"})
			return
		}
		id, err := uc.Create(c.Request.Context(), req)
		if err != nil {
			if usecase.IsInvalidInput(err) {
				response.WriteBadRequest(c.Writer, domain.ErrInvalidInput, err.Error(), nil)
				return
			}
			response.WriteError(c.Writer, err)
			return
		}
		response.WriteCreated(c.Writer, map[string]string{"id": id})
	}
}

func updateContent(uc ContentPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req domain.Content
		if err := response.DecodeJSON(c.Request.Body, &req); err != nil {
			response.WriteBadRequest(c.Writer, domain.ErrInvalidInput, "Invalid JSON", map[string]string{"body": "malformed"})
			return
		}
		if err := uc.Update(c.Request.Context(), id, req); err != nil {
			if usecase.IsInvalidInput(err) {
				response.WriteBadRequest(c.Writer, domain.ErrInvalidInput, err.Error(), nil)
				return
			}
			response.WriteError(c.Writer, err)
			return
		}
		response.WriteOK(c.Writer, map[string]string{"status": "updated"})
	}
}

func getContentByID(uc ContentPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		res, err := uc.GetByID(c.Request.Context(), id)
		if err != nil {
			response.WriteError(c.Writer, err)
			return
		}
		response.WriteOK(c.Writer, res)
	}
}

func searchContent(uc ContentPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		lang := c.Query("language")
		limit, _ := strconv.Atoi(c.Query("limit"))
		offset, _ := strconv.Atoi(c.Query("offset"))
		if limit <= 0 {
			limit = 20
		}
		items, total, err := uc.Search(c.Request.Context(), q, lang, limit, offset)
		if err != nil {
			response.WriteError(c.Writer, err)
			return
		}
		response.WriteOK(c.Writer, map[string]any{"items": items, "total": total})
	}
}

// ---------- Feedback handlers ----------
type feedbackReq struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity,omitempty"`
}

func submitFeedback(uc FeedbackPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := security.UserID(c)
		var req feedbackReq
		if err := response.DecodeJSON(c.Request.Body, &req); err != nil {
			response.WriteBadRequest(c.Writer, domain.ErrInvalidInput, "Invalid JSON", map[string]string{"body": "malformed"})
			return
		}
		id, err := uc.Submit(c.Request.Context(), domain.Feedback{
			UserID:      userID,
			Type:        req.Type,
			Description: req.Description,
			Severity:    req.Severity,
		})
		if err != nil {
			if usecase.IsInvalidInput(err) {
				response.WriteBadRequest(c.Writer, domain.ErrInvalidInput, err.Error(), nil)
				return
			}
			response.WriteError(c.Writer, err)
			return
		}
		response.WriteCreated(c.Writer, map[string]string{"id": id})
	}
}

// ---------- Legal entity handlers (teammate area) ----------
func createLegalEntity(uc LegalEntityPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req domain.LegalEntity
		if err := response.DecodeJSON(c.Request.Body, &req); err != nil {
			response.WriteBadRequest(c.Writer, domain.ErrInvalidInput, "Invalid JSON", map[string]string{"body": "malformed"})
			return
		}
		id, err := uc.Create(c.Request.Context(), req)
		if err != nil {
			if usecase.IsInvalidInput(err) {
				response.WriteBadRequest(c.Writer, domain.ErrInvalidInput, err.Error(), nil)
				return
			}
			response.WriteError(c.Writer, err)
			return
		}
		response.WriteCreated(c.Writer, map[string]string{"id": id})
	}
}

func updateLegalEntity(uc LegalEntityPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req domain.LegalEntity
		if err := response.DecodeJSON(c.Request.Body, &req); err != nil {
			response.WriteBadRequest(c.Writer, domain.ErrInvalidInput, "Invalid JSON", map[string]string{"body": "malformed"})
			return
		}
		if err := uc.Update(c.Request.Context(), id, req); err != nil {
			if usecase.IsInvalidInput(err) {
				response.WriteBadRequest(c.Writer, domain.ErrInvalidInput, err.Error(), nil)
				return
			}
			response.WriteError(c.Writer, err)
			return
		}
		response.WriteOK(c.Writer, map[string]string{"status": "updated"})
	}
}

func getLegalEntityByID(uc LegalEntityPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		res, err := uc.GetByID(c.Request.Context(), id)
		if err != nil {
			response.WriteError(c.Writer, err)
			return
		}
		response.WriteOK(c.Writer, res)
	}
}

func searchLegalEntities(uc LegalEntityPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		country := c.Query("country")
		limit, _ := strconv.Atoi(c.Query("limit"))
		offset, _ := strconv.Atoi(c.Query("offset"))
		if limit <= 0 {
			limit = 20
		}
		items, total, err := uc.Search(c.Request.Context(), q, country, limit, offset)
		if err != nil {
			response.WriteError(c.Writer, err)
			return
		}
		response.WriteOK(c.Writer, map[string]any{"items": items, "total": total})
	}
}

// ---------- Analytics handlers (teammate area) ----------
func queryTrends(uc AnalyticsPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		tw := c.Query("time_window")
		limit, _ := strconv.Atoi(c.Query("limit"))
		if limit <= 0 || limit > 1000 {
			limit = 50
		}
		data, err := uc.GetQueryTrends(c.Request.Context(), tw, limit)
		if err != nil {
			response.WriteError(c.Writer, err)
			return
		}
		response.WriteOK(c.Writer, data)
	}
}