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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
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
	// and Cosmos does some weird things like including the hash symbols (url encoded) but
	// removing the first space.
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

	// For every request, we are rebuilding our function set and instantiating a new
	// agent. Why?
	//
	// In the future, we might want to have a dynamic function set, created each time an
	// ask request comes in. This way, you can select what tools are most relevant and
	// omit the rest to save context space.
	//
	// A good approach to selecting relevant tools is to vectorize the user input and
	// compare against the function descriptions.
	//
	// In addition, we may want to set the agent system instruction or other configuration
	// dynamically based on the request and user context. e.g., adding system instructions
	// for complex tool selections, or defining user context information such as the
	// organization name.
	fs := s.GetFunctions()

	agent := bricks.NewBedrockAgent(bricks.BedrockAgentConfig{
		// Claude 3.7 is deprecated, but using an old/cheaper model to see how robust it
		// is.
		Model:       "us.anthropic.claude-3-7-sonnet-20250219-v1:0",
		Instruction: agentInstruction,
		AgentName:   "Fox",
		Functions:   fs,

		// Link to our knowledgebase. To see how this knowledgebase is built, see the
		// knowledgebase folder. We also have a video on it.
		Knowledgebases: []types.KnowledgeBase{
			{
				Description:     aws.String("Contains documentation on the Cosmos platform, helpful for resolving user queries about platform functionality"),
				KnowledgeBaseId: aws.String("CETGU0P5D7"),
			},
		},
	})

	// We pass along user information via the request context which is visible when
	// invoking tools.
	response, err := agent.Query(s.WrapContextForQuery(ctx, orgID, authorization), query, sessionID)
	if err != nil {
		return AskResult{}, err
	}

	// Translate refs to urls. When we ingest the knowledgebase, we are creating two
	// custom metadata fields: "header" and "folder".
	//
	// "header" is the section header delimiting where the content came from.
	// "folder" is the knowledgebase folder name.
	//
	// Transform these two into this format: Title - <URL>
	//
	// The URL format is <buaseurl>/<org_id>/documentation/<folder>#<anchor>
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
	fs := bricks.NewFunctionSet("bishopfox")
	fs.AddFunction("query_assets", queryAssetsDesc,
		QueryAssetsRequest{}, s.QueryAssetsFunction)
	fs.AddFunction("get_assets_overview_link", getAssetsOverviewLinkDesc,
		GetAssetsOverviewLinkRequest{}, s.GetAssetsOverviewLinkFunction)
	fs.AddFunction("get_latest_emerging_threats", getLatestEmergingThreatsDesc,
		GetLatestEmergingThreatsRequest{}, s.GetLatestEmergingThreatsFunction)
	return fs
}
