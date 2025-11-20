package main

import (
	"fmt"

	"github.com/bitovi/bishopfox-mcp-prototype/internal/service"
	"github.com/mark3labs/mcp-go/server"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func setupRouter(svc service.Service, mcpServer *server.StreamableHTTPServer) *gin.Engine {
	r := gin.Default()

	r.POST("/ask", AskHandler(svc))

	// A separate server created from mcp-go handles the /mcp endpoint. Forward requests
	// from that endpoint to there.
	//
	// The MCP server only uses POST endpoints. We use Any because the MCP server also
	// handles 405 for other method types (which is part of the MCP spec).
	r.Any("/mcp", func(c *gin.Context) {
		mcpServer.ServeHTTP(c.Writer, c.Request)
	})

	return r
}

// The /ask function invokes the internal LLM host in the service with the given query.
//
// So there are two ways to interact with the tools that the service provides. One is
// through this method, which is handling the LLM internally. The other is through the MCP
// interface, which is ONLY exposing the tools and doesn't contain any system prompt,
// model choice, etc.
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

		// TODO: organization_id validation. We want to verify that the Org ID matches
		// what the auth token allows. We also want to validate the auth token before it
		// gets in here (part of the standard routing middleware for BF services).
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
