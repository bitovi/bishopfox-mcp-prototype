package bricks

import (
	"context"
	"fmt"
)

var ErrInvalidArg = fmt.Errorf("invalid argument")

type Agent interface {
	Query(ctx context.Context, inputText string, sessionID string) (string, error)
}
