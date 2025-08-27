package httpserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/delivery/httpserver/transport"
	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/infrastructure/config"
	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/pkg/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	http *http.Server
	log  *zap.Logger
}

func NewServer(cfg *config.Config, logger *zap.Logger,
	content transport.ContentPort,
	feedback transport.FeedbackPort,
	legal transport.LegalEntityPort,
	analytics transport.AnalyticsPort,
) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(log.GinZap(logger))

	transport.RegisterRoutes(engine, cfg, logger, content, feedback, legal, analytics)

	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           engine,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return &Server{http: s, log: logger}
}

func (s *Server) Start() error {
	s.log.Info("HTTP server starting", zap.String("addr", s.http.Addr))
	return s.http.ListenAndServe()
}