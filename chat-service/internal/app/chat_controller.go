package app

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/domain"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/usecase"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/util"
)

type ChatController struct {
	chatService *usecase.ChatService
}

func NewChatController(cs *usecase.ChatService) *ChatController {
	return &ChatController{chatService: cs}
}

type QueryRequest struct {
	SessionID string `json:"sessionId"` // Can be empty for the first message
	Query     string `json:"query"`
	Language  string `json:"language"`
}

func (c *ChatController) postQuery(ctx *gin.Context) {
	// 1. Getting Inputs (request, userID, planID)
	userID, _ := ctx.Get("userID") // Will be empty string for guests if not set by middleware
	planID, _ := ctx.Get("planID") // Will be empty string or "visitor" for guests

	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}
	planIDStr := ""
	if planID != nil {
		planIDStr = planID.(string)
	}
	if planIDStr == "" { // Default to visitor if no plan is provided by middleware
		planIDStr = string(domain.TierGuest)
	}

	var reqBody QueryRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		util.SendSSEError(ctx.Writer, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// 2. Set up SSE headers
	w := ctx.Writer
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Adjust for CORS policy
	flusher, ok := w.(http.Flusher)
	if !ok {
		util.SendSSEError(w, "Streaming unsupported")
		return
	}

	// Ensure the client context is properly cancelled if the connection closes
	requestContext := ctx.Request.Context()

	// Create service request
	svcReq := usecase.QueryRequest{
		SessionID: reqBody.SessionID,
		UserID:    userIDStr,
		PlanID:    planIDStr,
		Message:   reqBody.Query,
		Language:  reqBody.Language,
	}

	// Call the service to process the query and get a stream
	responseStream, err := c.chatService.ProcessQuery(requestContext, svcReq)
	if err != nil {
		util.SendSSEError(w, "Service processing error: "+err.Error())
		flusher.Flush()
		return
	}

	// Stream the response chunks back to the client
	for chunk := range responseStream {
		select {
		case <-requestContext.Done(): // Check if request context cancelled
			log.Printf("Request context cancelled during SSE stream for session %s: %v", reqBody.SessionID, requestContext.Err())
			return
		default:
			if chunk.Error != nil {
				util.SendSSEError(w, chunk.Error.Error())
				flusher.Flush()
				return
			}

			if chunk.SessionID != "" {
				// This is the first chunk, containing the new session ID for client to use
				// Set cookie for guests, or just send the ID.
				userParams := domain.GetUserParamsFromPlanID(planIDStr)
				if !userParams.SaveHistory { // It's a guest
					setSessionIDCookie(w, chunk.SessionID)
				}
				util.SendSSEEvent(w, "session_id", map[string]string{"id": chunk.SessionID})
				flusher.Flush()
			} else if chunk.IsComplete {
				// Send the final chunk with suggested questions and completion only (no sources)
				data := map[string]interface{}{
					"is_complete": true,
				}
				if len(chunk.SuggestedQuestions) > 0 {
					data["suggested_questions"] = chunk.SuggestedQuestions
				}
				if chunk.Text != "" {
					data["text"] = chunk.Text
				}
				util.SendSSEEvent(w, "complete", data)
				flusher.Flush()
				return
			} else if chunk.Text != "" {
				// Send text chunks, include sources if present
				msg := map[string]interface{}{"text": chunk.Text}
				if len(chunk.Sources) > 0 {
					msg["sources"] = chunk.Sources
				}
				util.SendSSEEvent(w, "message", msg)
				flusher.Flush()
			}
		}
	}
}

func (c *ChatController) listSessions(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists || userID == "" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Query history is not available for visitors"})
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	sessions, total, err := c.chatService.ListSessions(ctx, userID.(string), page, limit) // Use chatService
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sessions"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"sessions": sessions, "total": total, "page": page, "limit": limit})
}

func (c *ChatController) getMessages(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists || userID == "" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Query history is not available for visitors"})
		return
	}

	sessionID := ctx.Param("sessionId")
	session, err := c.chatService.GetSession(ctx, sessionID) // Use chatService
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}
	// Ensure the user owns this session
	if session.UserID != userID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this session"})
		return
	}

	// Limit 0 implies no limit for history display for account holders
	messages, err := c.chatService.ListMessages(ctx, sessionID, 0)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}
	ctx.JSON(http.StatusOK, messages)
}

// setSessionIDCookie sets a session_id cookie for guests.
func setSessionIDCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour), // Example: 7 days expiry
		HttpOnly: true,
		Secure:   true, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}
