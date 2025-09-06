package app

import (
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/config"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/usecase"
)

type VoiceChatController struct {
	ChatService *usecase.ChatService
	Config      *config.Config
}

func NewVoiceChatController(chatService *usecase.ChatService, cfg *config.Config) *VoiceChatController {
	return &VoiceChatController{
		ChatService: chatService,
		Config:      cfg,
	}
}
