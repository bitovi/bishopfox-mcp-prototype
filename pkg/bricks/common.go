package bricks

import (
	"context"
	"fmt"
)

var ErrInvalidArg = fmt.Errorf("invalid argument")

type Reference struct {
	Type string `json:"type"`
	Data map[string]string
}

type QueryResult struct {
	Response string
	Refs     []Reference
}

type Agent interface {
	Query(ctx context.Context, inputText string, sessionID string) (QueryResult, error)
}
