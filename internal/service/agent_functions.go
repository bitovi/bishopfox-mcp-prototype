package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

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

		if v == nil {
			rowStrings = append(rowStrings, "")
			continue
		}

		switch fields[i].DataTypeOID {
		case pgtype.JSONBOID:
			jsonBytes, _ := json.Marshal(v)
			valueString = string(jsonBytes)
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

func getOrgHash(orgID uuid.UUID) string {
	// Simple hash function: take the last 12 characters of the UUID
	return orgID.String()[len(orgID.String())-12:]
}

var errQueryFailed = errors.New("query failed")

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

	role_suffix := getOrgHash(qc.OrgID)

	err = pgx.BeginFunc(c, conn, func(tx pgx.Tx) error {
		_, err := tx.Exec(c, `
			SET LOCAL ROLE customer_query_role_`+role_suffix+`;
		`)
		if err != nil {
			return fmt.Errorf("failed to set org_id; %w", err)
		}

		query := req.Query
		query = strings.ReplaceAll(query, "tbl_assets", "assets_org_"+role_suffix)

		rows, err := tx.Query(c, query)
		if err != nil {
			return fmt.Errorf("%w; %w", errQueryFailed, err)
		}
		defer rows.Close()

		var results []string
		resultsTruncated := false
		for rows.Next() {
			err := rows.Err()
			if err != nil {
				return fmt.Errorf("query failed (rows.Next): %w", err)
			}
			str, err := pgRowToString(rows)
			if err != nil {
				return err
			}
			results = append(results, str)
			if len(results) >= 15 {
				resultsTruncated = true
				break
			}
		}
		err = rows.Err()
		if err != nil {
			return fmt.Errorf("query failed (rows err after): %w", err)
		}

		fields := rows.FieldDescriptions()
		var header []string
		for _, f := range fields {
			header = append(header, f.Name)
		}
		output += strings.Join(header, "|") + "\n"
		if len(results) == 0 {
			output += "-- No results --"
			return nil
		}
		output += strings.Join(results, "\n")
		if resultsTruncated {
			output += "\n-- Result set truncated; more than 15 rows returned --"
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, errQueryFailed) {
			log.Debugln("query failed:", err)
			return "The query failed to execute with this error:\n" + err.Error(), nil
		}
		return nil, fmt.Errorf("db query failed; %w", err)
	}

	log.Debugln("query output:\n", output)
	return output, nil
}

// Input schema for querying assets with SQL. Output is a text string.
type DescribeAssetRequest struct {
	AssetID string `json:"asset_id" desc:"ID of the asset to describe" required:"true"`
}

/*
func (svc *MainService) DescribeAssetsFunction(c bricks.FunctionContext) (any, error) {
	var req DescribeAssetRequest
	c.MustBind(&req)

	qc, ok := c.Value(QueryContextKey{}).(QueryContext)
	if !ok {
		return nil, ErrMissingContext
	}
	log.Debugf("--- Received describe asset request ---\n%s\n---", req.AssetID)

	connUrl := svc.getDBUrl()
	conn, err := pgx.Connect(c, connUrl)
	if err != nil {
		return nil, err
	}
	defer conn.Close(c)

	var assetType, parentID, parentType, details, link string
	var tags []string
	err = conn.QueryRow(c, `
		SELECT type, parent_id, parent_type, details, tags, link
		FROM assets_org_`+getOrgHash(qc.OrgID)+`
		WHERE id = $1
	`, req.AssetID).Scan(&assetType, &parentID, &parentType, &details, &tags, &link)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "No asset found with the ID " + req.AssetID, nil
		}
		return nil, fmt.Errorf("db query failed; %w", err)
	}

}
*/

type GetAssetsOverviewLinkRequest struct {
	AssetType string `json:"asset_type" desc:"Type of asset to view, one of [domain, subdomain, ip, service, network, webapp]" required:"true"`
	Filters   string `json:"filters" desc:"Optional filters to apply, formatted as URL params, e.g. key1=value1&key2=value2" required:"false"`
	Search    string `json:"search" desc:"Optional search term to filter assets" required:"false"`
}

func (svc *MainService) GetAssetsOverviewLinkFunction(c bricks.FunctionContext) (any, error) {
	var req GetAssetsOverviewLinkRequest
	c.MustBind(&req)

	qc, ok := c.Value(QueryContextKey{}).(QueryContext)
	if !ok {
		return nil, ErrMissingContext
	}
	log.Debugf("--- Received get assets overview link request ---\n%+v\n---", req)

	baseUrl := fmt.Sprintf("https://ui.api.non.usea2.bf9.io/%s/assets/%s", qc.OrgID.String(), req.AssetType)

	// Note about these translations. See the instructions for this function, we are
	// avoiding using technical URL params in favor of ubiquitous language. For example,
	// instead of taking raw dates for expiration time (LLMs are not good with any kind of
	// math), we take the relative term in days and then programatically translate it
	// here.
	timeFormat := "2006-01-02T15:04:05.000Z"
	filters, err := url.ParseQuery(req.Filters)
	if err != nil {
		return "(Error) Invalid filters format. It needs to be formatted as a valid URL query string.", nil
	}
	params := url.Values{}
	for key, values := range filters {
		key = strings.ToLower(key)
		switch key {
		case "expiry": // translate to expiry range
			days, err := strconv.Atoi(values[0])
			if err != nil {
				// invalid input
				continue
			}
			timeNow := time.Now().UTC()
			timeNow = timeNow.Truncate(time.Hour * 24)
			if days == 0 {
				params.Set("expiry", "before,"+timeNow.Add(-time.Millisecond).Format(timeFormat))
			} else {
				expiryDate := timeNow.AddDate(0, 0, days)
				params.Set("expiry", "between,"+expiryDate.Format(timeFormat)+","+timeNow.Format(timeFormat))
			}
		case "tld":
			domains := strings.Split(values[0], " ")
			for i := range domains {
				domains[i] = "." + domains[i]
			}
			params.Set("domainExtension", strings.Join(domains, ","))
		case "tags":
			tags := strings.Split(values[0], " ")
			params.Set("tags", strings.Join(tags, ","))
		default:
			for _, value := range values {
				params.Add(key, value)
			}
		}
	}

	if req.Search != "" {
		params.Set("search", req.Search)
	}

	fullURL, _ := url.Parse(baseUrl)
	fullURL.RawQuery = params.Encode()
	return fullURL.String(), nil
}
