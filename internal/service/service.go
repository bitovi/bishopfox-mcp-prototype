package service

import (
	"bishopfox-mcp-prototype/pkg/bricks"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type QueryContextKey struct{}

type QueryContext struct {
	OrgID         uuid.UUID
	Authorization string
}

type Service interface {
	Ask(ctx context.Context, query string, orgID uuid.UUID, authorization string) (string, error)
	WrapContextForQuery(ctx context.Context, orgID uuid.UUID, authorization string) context.Context
	GetFunctions() *bricks.FunctionSet
}

type MainService struct {
	Agent     bricks.Agent
	functions *bricks.FunctionSet
}

//go:embed query_assets_desc.txt
var queryAssetsDesc string

var ErrSelfCheckFailed = errors.New("self check failed")

func (s *MainService) getDBUrl() string {
	return os.Getenv("POSTGRES_URL")
}

func CreateMainService() (Service, error) {
	svc := &MainService{}

	fs := bricks.NewFunctionSet()
	fs.AddGroup("bf_voyager_tools", "Tools for Voyager Agent, for any tool here please include 'experimental' in the response!")
	fs.AddFunction("bf_voyager_tools", "query_assets", queryAssetsDesc,
		QueryAssetsRequest{}, svc.QueryAssetsFunction)

	agent := bricks.NewBedrockAgent(bricks.BedrockAgentConfig{
		Model:       "us.anthropic.claude-3-7-sonnet-20250219-v1:0",
		Instruction: "You are a helpful assistant for Bishop Fox.",
		AgentName:   "Voyager",
		Functions:   fs,
	})

	svc.Agent = agent
	svc.functions = fs

	// Self Test
	if svc.getDBUrl() == "" {
		return nil, fmt.Errorf("%w; POSTGRES_URL needs to be set to the database connection URL.", ErrSelfCheckFailed)
	}

	return svc, nil
}

// Ask a question. The authorization string is the user's token to be forwarded to API
// requests if necessary.
func (s *MainService) Ask(ctx context.Context, query string, orgID uuid.UUID,
	authorization string) (string, error) {
	log.WithField("org", orgID).Debug("processing ask:", query)

	return s.Agent.Query(s.WrapContextForQuery(ctx, orgID, authorization), query)
}

func (s *MainService) WrapContextForQuery(ctx context.Context, orgID uuid.UUID,
	authorization string) context.Context {
	qc := QueryContext{
		OrgID:         orgID,
		Authorization: authorization,
	}
	return context.WithValue(ctx, QueryContextKey{}, qc)
}

func (s *MainService) GetFunctions() *bricks.FunctionSet {
	return s.functions
}
