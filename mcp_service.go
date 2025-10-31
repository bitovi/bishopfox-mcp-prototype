package main

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/bitovi/bishopfox-mcp-prototype/internal/service"
	"github.com/bitovi/bishopfox-mcp-prototype/pkg/bricks"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
)

func newAuthenticationMiddleware(svc service.Service) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			auth := request.Header.Get("Authorization")
			if auth == "" {
				return mcp.NewToolResultError("missing Authorization header"), nil
			}

			// TODO: Validate authorization.

			orgid := request.Header.Get("X-BF-OrgID")
			// Note that we should also explore using the _meta field to pass additional context like organization_id.
			// Claude Desktop doesn't support it.

			// TODO: If empty, pull from auth token.
			// TODO: Validate access to org.

			ctx = svc.WrapContextForQuery(ctx, uuid.MustParse(orgid), auth)

			// Proceed to the next handler
			return next(ctx, request)
		}
	}
}

func mcpRecovery(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, rerr error) {
		defer func() {
			if r := recover(); r != nil {
				traceback := string(debug.Stack())
				log.Errorf("panic in MCP tool handler: %v\n%s", r, traceback)
				result = mcp.NewToolResultError("function invocation failed unexpectedly")
			}
		}()
		return next(ctx, request)
	}
}

func runMCPServer(svc service.Service) {
	// Create a new MCP server
	serverBase := server.NewMCPServer(
		"Calculator Demo",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
		server.WithToolHandlerMiddleware(mcpRecovery),
		server.WithToolHandlerMiddleware(newAuthenticationMiddleware(svc)),
	)

	fs := svc.GetFunctions()

	for groupName := range fs.Groups {
		bricks.BindFunctionsToMCPServer(fs, groupName, serverBase)
	}

	httpServer := server.NewStreamableHTTPServer(
		serverBase,
		server.WithEndpointPath("/mcp"),
	)

	mcpPort := os.Getenv("MCP_PORT")
	if mcpPort == "" {
		mcpPort = "8102"
	}

	binding := ":" + mcpPort
	log.Infoln("Starting MCP server on", binding)
	if err := httpServer.Start(binding); err != nil {
		log.Errorf("Streamable HTTP server error: %v", err)
	}
}
