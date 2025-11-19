package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/bitovi/bishopfox-mcp-prototype/internal/service"

	log "github.com/sirupsen/logrus"
)

var ErrInvalidArg = fmt.Errorf("invalid argument")

func main() {
	log.SetLevel(log.DebugLevel)
	log.Infoln("Starting test service")
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	svc, err := service.CreateMainService()
	if err != nil {
		log.Errorf("Failed to create service: %v", err)
		return
	}
	router := setupRouter(svc)
	go newMCPServer(svc)

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8100"
	}

	srv := &http.Server{
		Addr:    ":" + apiPort,
		Handler: router,
	}

	log.Infoln("API server starting on", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Errorf("API server failed: %v", err)
		return
	}
	log.Infoln("API server stopped.")
}
