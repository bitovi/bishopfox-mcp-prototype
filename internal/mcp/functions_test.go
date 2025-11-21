// Example test approach.
package mcp_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bitovi/bishopfox-mcp-prototype/internal/mcp"
	"github.com/bitovi/bishopfox-mcp-prototype/internal/service"
	"github.com/bitovi/bishopfox-mcp-prototype/pkg/bricks"
	"github.com/google/uuid"
)

type MockService struct{}

func (m *MockService) QueryAssets(ctx context.Context, orgID uuid.UUID, query string) (service.QueryAssetsResult, error) {
	return service.QueryAssetsResult{
		Columns: []string{"query", "orgid"},
		Rows: [][]string{
			{query, orgID.String()},
		},
		Truncated: false,
	}, nil
}

func (m *MockService) Ask(ctx context.Context, query string, orgID uuid.UUID, authorization string, sessionID string) (service.AskResult, error) {
	return service.AskResult{}, nil
}

func (m *MockService) SetFunctions(fs *bricks.FunctionSet) {}

func toJSON(data any) []byte {
	b, _ := json.Marshal(data)
	return b
}

func TestQueryAssetsFunction(t *testing.T) {
	svc := &MockService{}

	orgID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	auth := "test-auth"
	query := "testing-query"

	fs := mcp.GetFunctions(svc)
	result, err := fs.Invoke(
		service.WrapContextForTool(context.Background(), orgID, auth, svc),
		"query_assets",
		toJSON(mcp.QueryAssetsRequest{
			Query: query,
		}),
	)
	if err != nil {
		t.Fatalf("Function invocation failed: %v", err)
	}

	// [SPEC] The query_assets tool returns a string response in a CSV format. First
	// we have a header with column names, then rows of data. Columns are separated by |.
	// Includes a trailing newline at the end.
	strResult, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string result, got %T", result)
	}

	expected := "query|orgid\ntesting-query|11111111-1111-1111-1111-111111111111"
	if strResult != expected {
		t.Errorf("Expected result %q, got %q", expected, strResult)
	}

	// Todo...
	// [SPEC] If there are no results, the text "-- No results --" is included.
	// [SPEC] If there are more than 15 results, the results are truncated and the text
	// "-- Result set truncated; more than 15 rows returned --"
	// is included at the end.
	//
	// These help to guide the model.
}
