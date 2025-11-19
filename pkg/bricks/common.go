// Package bricks provides utilities for querying AI agents with tool support. For
// example, we have a BedrockAgent implementation that leverages action groups and the
// Bedrock Agent Runtime to satisfy queries.
//
// In addition, we have an MCP binder to expose functions as MCP tools via the mcp-go
// library.
package bricks

import (
	"context"
	"fmt"
)

var ErrInvalidArg = fmt.Errorf("invalid argument")

// "References" link back to what sources were used while answering a query. Bedrock calls
// them "citations" in its responses when it references knowledgebases or other resources.
type Reference struct {
	// One of [knowledgebase] (only one reference type implemented currently.
	Type string `json:"type"`
	// The reference data, e.g., this could contain a URL, Bedrock knowledgebase ID, etc.
	// Freeform key-value store.
	Data map[string]string
}

// Result of an agent query. Includes the text response and any references used.
type QueryResult struct {
	Response string
	Refs     []Reference
}

type Agent interface {
	// Query the agent with the input text. The session ID should be a unique string that
	// can be issued again later to continue the same session.
	///
	// TODO: That is how Bedrock works. We might want to rethink that if supporting other
	// vendors. For example, including SessionID in the QueryResult to be used in future
	// calls. Or a SessionHandle interface if we need to be more flexible.
	Query(ctx context.Context, inputText string, sessionID string) (QueryResult, error)
}
