package main

import (
	"fmt"

	"github.com/bitovi/bishopfox-mcp-prototype/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func setupRouter(svc service.Service) *gin.Engine {
	r := gin.Default()
	r.POST("/ask", AskHandler(svc))
	return r
}

func AskHandler(svc service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		var req struct {
			Query     string `json:"query"`
			OrgID     string `form:"organization_id"`
			SessionID string `form:"session_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// TODO: organization_id validation
		orgID := uuid.MustParse(req.OrgID)
		if req.SessionID != "" {
			if _, err := uuid.Parse(req.SessionID); err != nil {
				c.JSON(400, gin.H{"error": "session_id must be empty or a valid UUID"})
				return
			}
		}

		response, err := svc.Ask(c.Request.Context(), req.Query, orgID, auth, req.SessionID)
		if err != nil {
			fmt.Println(err)
			c.JSON(500, gin.H{"error": "Failed to process request; the issue has been logged"})
			return
		}

		log.Debugln("/ask response:", response)

		c.JSON(200, gin.H{
			"session_id": response.SessionID,
			"data":       response.Response,
			"refs":       response.Refs,
		})
	}
}
