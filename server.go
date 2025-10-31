package main

import (
	"fmt"

	"github.com/bitovi/bishopfox-mcp-prototype/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
			Query string `json:"query"`
			OrgID string `form:"organization_id"`
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

		response, err := svc.Ask(c.Request.Context(), req.Query, orgID, auth)
		if err != nil {
			fmt.Println(err)
			c.JSON(500, gin.H{"error": "Failed to process request; the issue has been logged"})
			return
		}

		c.JSON(200, gin.H{"data": response})
	}
}
