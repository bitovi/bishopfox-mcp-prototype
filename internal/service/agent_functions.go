package service

import (
	"bishopfox-mcp-prototype/pkg/bricks"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	log "github.com/sirupsen/logrus"
)

var ErrMissingContext = errors.New("missing required context value")

type GetWeatherRequest struct {
	City string `json:"city" desc:"The name of the city to get weather for" required:"true"`
}

type GetWeatherResponse struct {
	Temperature string `json:"temperature"`
	Condition   string `json:"condition"`
}

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

type QueryAssetsRequest struct {
	Query string `json:"query" desc:"The SQL query to execute" required:"true"`
}

type SearchAssetsResponse struct {
	Results []string `json:"results"`
}

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
