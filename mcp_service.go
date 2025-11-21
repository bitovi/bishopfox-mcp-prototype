package main

import (
	"context"
	"net/http"
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

			// TODO: We can validate authorization here.

			orgID, ok := ctx.Value(orgIDContextKey{}).(string)
			if !ok || orgID == "" {
				// TODO: If empty, pull from auth token.
				return mcp.NewToolResultError("missing organization_id param"), nil
			}
			orgUUID, err := uuid.Parse(orgID)
			if err != nil {
				return mcp.NewToolResultError("invalid organization_id param"), nil
			}
			// There is also the _meta field that can pass information like the
			// organization ID, but that is something less accessible to existing tools.
			// For example, there is no configuration for Claude Desktop for specifying
			// _meta fields.

			// TODO: Validate access to org.

			ctx = service.WrapContextForTool(ctx, orgUUID, auth, svc)

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

// This middleware is executed at the server level to capture the organization_id from the
// query param and add it to the current context. Naturally, this is only relevant to HTTP
// transports. For Stdio, we would want to use a commandline parameter or env variable.
func addHTTPContext(ctx context.Context, r *http.Request) context.Context {
	orgID := r.URL.Query().Get("organization_id")
	return context.WithValue(ctx, orgIDContextKey{}, orgID)
}

func newMCPServer(svc service.Service, fs *bricks.FunctionSet) *server.StreamableHTTPServer {
	// Create a new MCP server
	serverBase := server.NewMCPServer(
		"Cosmos MCP",
		"1.0.0",
		server.WithToolCapabilities(false),
		// Default recovery middleware may not be a good choice. We might want finer
		// control over what the model sees if a tool call fails. Unlike normal API
		// failures where the execution terminates, tool failures are returned to the
		// model and inference may continue, so it matters what shows up in the error
		// message.
		server.WithRecovery(),
		server.WithToolHandlerMiddleware(mcpRecovery),
		server.WithToolHandlerMiddleware(newAuthenticationMiddleware(svc)),
	)

	bricks.BindFunctionsToMCPServer(fs, serverBase)

	return server.NewStreamableHTTPServer(
		serverBase,
		server.WithStateLess(true),
		server.WithEndpointPath("/mcp"),
		server.WithDisableStreaming(true),
		server.WithHTTPContextFunc(addHTTPContext),
	)
}
