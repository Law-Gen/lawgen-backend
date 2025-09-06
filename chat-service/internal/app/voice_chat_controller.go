package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/config"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

// VoiceChatController handles voice chat requests
// Add to router as needed
// VoiceChatHandlerWithConfig returns a handler using the provided config
func VoiceChatHandlerWithConfig(chatService *usecase.ChatService, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. Receive audio file from frontend
		file, header, err := ctx.Request.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing audio file"})
			return
		}
		defer file.Close()

		// Validate file type and size
		contentType := header.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "audio/") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type: " + contentType})
			return
		}
		buf := new(bytes.Buffer)
		size, err := io.Copy(buf, file)
		if err != nil || size == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Empty or unreadable audio file"})
			return
		}

		// 2. Send audio to STT API (use correct language endpoint)
		lang := ctx.DefaultPostForm("language", "en")
		sttAPIURL := fmt.Sprintf("%s%s?mode=file", cfg.STTApiBase, lang)
		var sttBuf bytes.Buffer
		writer := multipart.NewWriter(&sttBuf)
		fileContentType := header.Header.Get("Content-Type")
		if fileContentType == "" {
			fileContentType = "audio/mpeg"
		}
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, header.Filename))
		h.Set("Content-Type", fileContentType)
		part, err := writer.CreatePart(h)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create form file"})
			return
		}
		io.Copy(part, bytes.NewReader(buf.Bytes()))
		writer.Close()

		sttReq, err := http.NewRequest("POST", sttAPIURL, &sttBuf)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create STT request"})
			return
		}
		sttReq.Header.Set("Content-Type", writer.FormDataContentType())
		sttResp, err := http.DefaultClient.Do(sttReq)
		if err != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": "STT service unreachable: " + err.Error()})
			return
		}
		defer sttResp.Body.Close()
		if sttResp.StatusCode != 200 {
			body, _ := io.ReadAll(sttResp.Body)
			ctx.JSON(http.StatusBadGateway, gin.H{"error": "STT service error", "details": string(body)})
			return
		}
		var sttResult struct {
			Text string `json:"text"`
		}
		json.NewDecoder(sttResp.Body).Decode(&sttResult)
		if sttResult.Text == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "No text transcribed"})
			return
		}

		// 3. If language is not English, translate to English
		queryText := sttResult.Text
		if lang != "en" {
			transReqBody, _ := json.Marshal(map[string]string{
				"text":        sttResult.Text,
				"target_lang": "en",
			})
			transResp, err := http.Post(cfg.TranslateApiUrl, "application/json", bytes.NewReader(transReqBody))
			if err != nil || transResp.StatusCode != 200 {
				ctx.JSON(http.StatusBadGateway, gin.H{"error": "Translation service error"})
				return
			}
			defer transResp.Body.Close()
			var transResult struct {
				TranslatedText string `json:"translated_text"`
			}
			json.NewDecoder(transResp.Body).Decode(&transResult)
			if transResult.TranslatedText != "" {
				queryText = transResult.TranslatedText
			}
		}

		// 4. Call chat service API
		chatReq := usecase.QueryRequest{
			SessionID: ctx.DefaultPostForm("sessionId", ""),
			UserID:    ctx.DefaultPostForm("userId", ""),
			PlanID:    ctx.DefaultPostForm("planId", "visitor"),
			Message:   queryText,
			Language:  "en",
		}
		responseStream, err := chatService.ProcessQuery(ctx, chatReq)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Chat service error"})
			return
		}

		// 5. Collect full answer from stream
		var answerBuilder strings.Builder
		for chunk := range responseStream {
			if chunk.Error != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": chunk.Error.Error()})
				return
			}
			answerBuilder.WriteString(chunk.Text)
		}
		finalAnswer := answerBuilder.String()

		// 6. If original language is not English, translate back
		if lang != "en" {
			transReqBody, _ := json.Marshal(map[string]string{
				"text":        finalAnswer,
				"target_lang": lang,
			})
			transResp, err := http.Post(cfg.TranslateApiUrl, "application/json", bytes.NewReader(transReqBody))
			if err == nil && transResp.StatusCode == 200 {
				defer transResp.Body.Close()
				var transResult struct {
					TranslatedText string `json:"translated_text"`
				}
				json.NewDecoder(transResp.Body).Decode(&transResult)
				if transResult.TranslatedText != "" {
					finalAnswer = transResult.TranslatedText
				}
			}
		}

		// 7. Convert answer to voice
		ttsReqBody, _ := json.Marshal(map[string]string{
			"text":     finalAnswer,
			"language": lang,
		})
		ttsResp, err := http.Post(cfg.TTSApiUrl, "application/json", bytes.NewReader(ttsReqBody))
		if err != nil || ttsResp.StatusCode != 200 {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": "TTS service error"})
			return
		}
		defer ttsResp.Body.Close()
		ctx.Header("Content-Type", "audio/mpeg")
		io.Copy(ctx.Writer, ttsResp.Body)
	}
}
