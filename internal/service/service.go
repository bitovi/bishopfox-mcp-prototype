package service

// This file contains the implementation for a prototype "service", similar to what Bishop
// Fox has in their environment.

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/bitovi/bishopfox-mcp-prototype/pkg/bricks"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// The key that holds QueryContext.
type QueryContextKey struct{}

// The QueryContext traverses the request lifetime via the context variable, carrying
// extra information relevant to Cosmos such as the organization_id and authorization
// token.
//
// The authorization token is especially useful if you want to call other APIs on the
// user's behalf. For example, an asset question could make a call to the asset token
// service.
//
// Keep in mind though that this approach might not be forward compatible if the
// infrastructure changes to have tokens that aren't forwarded directly to the service, or
// otherwise cannot be "replayed" like this. In those cases, you may want to consider
// traditional m2m tokens or other communication approaches (or data duplication) to get
// the needed data.
type QueryContext struct {
	OrgID         uuid.UUID
	Authorization string
}

// A reference points to a source of information used in generating a response.
//
// Currently we have Bedrock Citations stored in here as type "knowledgebase". Our service
// formats those as a title and URL.
type Reference struct {
	Type string `json:"type"`
	Ref  string `json:"ref"`
}

// Result of the Ask function. Contains the response, a session ID for making further
// requests (which can be set initially by the caller), and any references used.
type AskResult struct {
	Response  string   `json:"response"`
	Refs      []string `json:"refs"`
	SessionID string   `json:"session_id"`
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

//go:embed prompt/get_assets_overview_link_desc.txt
var getAssetsOverviewLinkDesc string

//go:embed prompt/get_latest_emerging_threats_desc.txt
var getLatestEmergingThreatsDesc string

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
	fs.AddGroup("bf_api", "Bishop Fox API")
	fs.AddFunction("bf_api", "query_assets", queryAssetsDesc,
		QueryAssetsRequest{}, svc.QueryAssetsFunction)
	fs.AddFunction("bf_api", "get_assets_overview_link", getAssetsOverviewLinkDesc,
		GetAssetsOverviewLinkRequest{}, svc.GetAssetsOverviewLinkFunction)
	fs.AddFunction("bf_api", "get_latest_emerging_threats", getLatestEmergingThreatsDesc,
		GetLatestEmergingThreatsRequest{}, svc.GetLatestEmergingThreatsFunction)

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

func formatRefTitle(header string) string {
	title := strings.TrimLeft(header, "#")
	title = strings.TrimSpace(title)

	return title
}

// This needs to match the UI formatting code. It might not currently.
func formatRefAnchor(header string) string {
	// Lowercase the header
	header = strings.ToLower(header)
	header = regexp.MustCompile(`^(#+)\s*`).ReplaceAllString(header, `$1`)
	header = strings.ReplaceAll(header, " ", "-")

	// URL encode the header to make it safe for use in URLs
	return url.QueryEscape(header)
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

	// Translate refs to urls
	var refURLs []string
	baseUrl := "https://ui.api.non.usea2.bf9.io"
	for _, ref := range response.Refs {
		// For now, only handling knowledge base refs
		if ref.Type == "knowledgebase" {
			header := ref.Data["header"]
			folder := ref.Data["folder"]
			if header == "" || folder == "" {
				continue
			}
			url := fmt.Sprintf("%s - %s/%s/documentation/%s#%s",
				formatRefTitle(header),
				baseUrl,
				orgID.String(),
				folder,
				formatRefAnchor(header))
			refURLs = append(refURLs, url)
		}
	}

	return AskResult{
		Response:  response.Response,
		Refs:      refURLs,
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
