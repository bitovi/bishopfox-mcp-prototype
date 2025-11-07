package main

import (
	"context"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/bitovi/bishopfox-mcp-prototype/internal/service"
	"github.com/bitovi/bishopfox-mcp-prototype/pkg/bricks"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
)

// func getHTTPRequestFromContext(ctx context.Context) *http.Request {
// 	if req, ok := ctx.Value(http.RequestContextKey).(*http.Request); ok {
// 		return req
// 	}
// 	return nil
// }

func newAuthenticationMiddleware(svc service.Service) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			auth := request.Header.Get("Authorization")
			if auth == "" {
				return mcp.NewToolResultError("missing Authorization header"), nil
			}

			// TODO: Validate authorization.

			orgID, ok := ctx.Value(orgIDContextKey{}).(string)
			if !ok || orgID == "" {
				return mcp.NewToolResultError("missing organization_id param"), nil
			}
			orgUUID, err := uuid.Parse(orgID)
			if err != nil {
				return mcp.NewToolResultError("invalid organization_id param"), nil
			}
			// Note that we should also explore using the _meta field to pass additional context like organization_id.
			// Claude Desktop doesn't support it.

			// TODO: If empty, pull from auth token.
			// TODO: Validate access to org.

			ctx = svc.WrapContextForQuery(ctx, orgUUID, auth)

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

type orgIDContextKey struct{}

func addHTTPContext(ctx context.Context, r *http.Request) context.Context {
	orgID := r.URL.Query().Get("organization_id")
	return context.WithValue(ctx, orgIDContextKey{}, orgID)
}

func runMCPServer(svc service.Service) {
	// Create a new MCP server
	serverBase := server.NewMCPServer(
		"Bishop Fox MCP Prototype",
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
		server.WithStateLess(true),
		server.WithEndpointPath("/mcp"),
		server.WithDisableStreaming(true),
		server.WithHTTPContextFunc(addHTTPContext),
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
