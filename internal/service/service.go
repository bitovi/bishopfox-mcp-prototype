package service

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/bitovi/bishopfox-mcp-prototype/pkg/bricks"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// The key that holds QueryContext.
type QueryContextKey struct{}

// Holds authentication information for queries/tools.
type QueryContext struct {
	OrgID         uuid.UUID
	Authorization string
}

type AskResult struct {
	Response  string `json:"response"`
	SessionID string `json:"session_id"`
}

// Service interface for consumers.
type Service interface {
	Ask(ctx context.Context, query string, orgID uuid.UUID, authorization string, sessionID string) (AskResult, error)
	WrapContextForQuery(ctx context.Context, orgID uuid.UUID, authorization string) context.Context
	GetFunctions() *bricks.FunctionSet
}

// Main/default service implementation.
type MainService struct {
	Agent     bricks.Agent
	functions *bricks.FunctionSet
}

//go:embed prompt/query_assets_desc.txt
var queryAssetsDesc string

//go:embed prompt/agent_instruction.txt
var agentInstruction string

var ErrSelfCheckFailed = errors.New("self check failed")

// Returns the database connection URL as set in the environment. Empty values should be
// treated as an initialization error.
func (s *MainService) getDBUrl() string {
	return os.Getenv("POSTGRES_URL")
}

// Create the service.
func CreateMainService() (Service, error) {
	svc := &MainService{}

	fs := bricks.NewFunctionSet()
	fs.AddGroup("bf_voyager_tools", "Tools for Voyager Agent, for any tool here please include 'experimental' in the response!")
	fs.AddFunction("bf_voyager_tools", "query_assets", queryAssetsDesc,
		QueryAssetsRequest{}, svc.QueryAssetsFunction)

	agent := bricks.NewBedrockAgent(bricks.BedrockAgentConfig{
		Model:       "us.anthropic.claude-3-7-sonnet-20250219-v1:0",
		Instruction: agentInstruction,
		AgentName:   "Voyager",
		Functions:   fs,
	})

	svc.Agent = agent
	svc.functions = fs

	// Self Test
	if svc.getDBUrl() == "" {
		return nil, fmt.Errorf("%w; POSTGRES_URL needs to be set to the database connection URL", ErrSelfCheckFailed)
	}

	return svc, nil
}

// Ask a question. The authorization string is the user's token to be forwarded to API
// requests if necessary.
func (s *MainService) Ask(ctx context.Context, query string, orgID uuid.UUID,
	authorization string, sessionID string) (AskResult, error) {
	log.WithField("org", orgID).Debug("processing ask:", query)

	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	response, err := s.Agent.Query(s.WrapContextForQuery(ctx, orgID, authorization), query, sessionID)
	if err != nil {
		return AskResult{}, err
	}

	return AskResult{
		Response:  response,
		SessionID: sessionID,
	}, nil
}

// Wrap a given context for an agent query, adding authorization information. This context
// is passed through the agent to functions. Particularly useful for forwarding user
// authentication and organization restrictions.
func (s *MainService) WrapContextForQuery(ctx context.Context, orgID uuid.UUID,
	authorization string) context.Context {
	qc := QueryContext{
		OrgID:         orgID,
		Authorization: authorization,
	}
	return context.WithValue(ctx, QueryContextKey{}, qc)
}

// Functions are exposed so they can be bound to external access like MCP servers. The
// functions themselves are bound to this service instance during initialization.
func (s *MainService) GetFunctions() *bricks.FunctionSet {
	return s.functions
}
