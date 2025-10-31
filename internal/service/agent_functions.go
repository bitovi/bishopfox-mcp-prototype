package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bitovi/bishopfox-mcp-prototype/pkg/bricks"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	log "github.com/sirupsen/logrus"
)

var ErrMissingContext = errors.New("missing required context value")

// This is an example request schema.
//   - json: Name of the field exposed in the MCP.
//   - desc: Description shown in the MCP tool listing. This is typically a shorter
//     description as you can leave the main instructions in the upper
//     function-level description.
//   - required: Whether the field is required. Set to 'true' to reflect in the MCP schema
//     that the field is required.
type GetWeatherRequest struct {
	City string `json:"city" desc:"The name of the city to get weather for" required:"true"`
}

// This is an example response schema. The MCP supports output schemas. Bedrock Action
// Groups do not. Currently, we are not implementing MCP output schemas (see mcp.go). We
// believe that output schemas are not entirely useful for LLMs.
type GetWeatherResponse struct {
	Temperature string `json:"temperature"`
	Condition   string `json:"condition"`
}

// Example function implementation. When an LLM makes a tool request, it is routed to
// these handlers. The flow for Bedrock invoking a function hosted by the client
// application is referred to as RETURN_CONTROL.
func (svc *MainService) GetWeatherFunction(c bricks.FunctionContext) (any, error) {
	var req GetWeatherRequest
	c.MustBind(&req)

	qc, ok := c.Value(QueryContextKey{}).(QueryContext)
	if !ok {
		return nil, ErrMissingContext
	}

	log.Debugln("authorization:", qc.Authorization)

	// Simulate a weather API call
	response := GetWeatherResponse{
		Temperature: "72°F",
		Condition:   "Sunny",
	}
	return response, nil
}

// Input schema for querying assets with SQL. Output is a text string.
type QueryAssetsRequest struct {
	Query string `json:"query" desc:"The SQL query to execute" required:"true"`
}

// Convert an arbitrary SQL result to a string representation. Fields are separated by
// '|'. Certain field types may not be supported properly.
func pgRowToString(rows pgx.Rows) (string, error) {
	values, err := rows.Values()
	if err != nil {
		return "", fmt.Errorf("failed to get row values: %w", err)
	}

	var rowStrings []string
	fields := rows.FieldDescriptions()
	for i, v := range values {
		var valueString string

		switch fields[i].DataTypeOID {
		case pgtype.UUIDOID:
			valueString = uuid.UUID(v.([16]byte)).String()
		case pgtype.TextArrayOID, pgtype.VarcharArrayOID:
			// This column is a []string array
			arrayValues := []string{}
			for _, s := range v.([]any) {
				if s != nil {
					arrayValues = append(arrayValues, fmt.Sprint(s))
				}
			}
			valueString = strings.Join(arrayValues, ",")
		default:
			valueString = fmt.Sprint(v)
		}
		rowStrings = append(rowStrings, valueString)
	}

	return strings.Join(rowStrings, "|"), nil
}

// Handler for query_assets. Queries the asset database with a given SQL query and returns
// the result. Enforces row security so that the query can only access the rows for the
// organization in the context.
func (svc *MainService) QueryAssetsFunction(c bricks.FunctionContext) (any, error) {
	var req QueryAssetsRequest
	c.MustBind(&req)

	qc, ok := c.Value(QueryContextKey{}).(QueryContext)
	if !ok {
		return nil, ErrMissingContext
	}
	log.Debugln("authorization:", qc.Authorization)
	log.Debugf("--- Received query request ---\n%s\n---", req.Query)

	connUrl := svc.getDBUrl()
	conn, err := pgx.Connect(c, connUrl)
	if err != nil {
		return nil, err
	}
	defer conn.Close(c)

	var output string

	err = pgx.BeginFunc(c, conn, func(tx pgx.Tx) error {
		_, err := tx.Exec(c, `
			
			SET LOCAL app.org_id = '`+qc.OrgID.String()+`';
			SET LOCAL ROLE customer_query_role;
		`)
		if err != nil {
			return fmt.Errorf("failed to set org_id; %w", err)
		}

		rows, err := tx.Query(c, req.Query)
		if err != nil {
			return fmt.Errorf("query execution failed; %w", err)
		}
		defer rows.Close()

		var results []string
		for rows.Next() {
			err := rows.Err()
			if err != nil {
				return fmt.Errorf("query execution failed; %w", err)
			}
			str, err := pgRowToString(rows)
			if err != nil {
				return err
			}
			results = append(results, str)
		}

		fields := rows.FieldDescriptions()
		var header []string
		for _, f := range fields {
			header = append(header, f.Name)
		}
		output += strings.Join(header, "|") + "\n"
		output += strings.Join(results, "\n")

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("db query failed; %w", err)
	}

	return output, nil
}
