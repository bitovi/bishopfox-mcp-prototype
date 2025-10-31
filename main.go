package main

import (
	"bishopfox-mcp-prototype/internal/service"
	"errors"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

var ErrInvalidArg = fmt.Errorf("invalid argument")

func main() {
	log.SetLevel(log.DebugLevel)
	log.Infoln("Starting test service")

	os.Setenv("AWS_PROFILE", "bfbedrock")

	svc, err := service.CreateMainService()
	if err != nil {
		log.Errorf("Failed to create service: %v", err)
		return
	}
	router := setupRouter(svc)
	go runMCPServer(svc)

	srv := &http.Server{
		Addr:    ":8100",
		Handler: router,
	}

	log.Infoln("API server starting")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Errorf("API server failed: %v", err)
		return
	}
	log.Infoln("API server stopped.")
}
