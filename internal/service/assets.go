package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Functions for querying and working with assets.

type QueryAssetsResult struct {
	Columns   []string
	Rows      [][]string
	Truncated bool
}

// Returned when the query failed to execute from invalid input.
var ErrQueryFailed = errors.New("query failed")

// Returns a short hash for the given organization ID.
func getOrgHash(orgID uuid.UUID) string {
	// Simple hash function: take the last 12 characters of the UUID and treat that as the
	// hash. This is the same thing Cosmos does when partitioning asset tables.
	return orgID.String()[len(orgID.String())-12:]
}

// Convert an arbitrary SQL result to a string representation. Certain field types may not
// be supported properly.
func formatPGRow(rows pgx.Rows) ([]string, error) {
	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("failed to get row values: %w", err)
	}

	var rowStrings []string
	fields := rows.FieldDescriptions()
	for i, v := range values {
		var valueString string

		if v == nil {
			rowStrings = append(rowStrings, "")
			continue
		}

		// Check the type of the value to format the fields.
		switch fields[i].DataTypeOID {
		case pgtype.JSONBOID:
			// JSON field, format as JSON string
			jsonBytes, _ := json.Marshal(v)
			valueString = string(jsonBytes)
		case pgtype.UUIDOID:
			// UUID field, format as hex UUID.
			valueString = uuid.UUID(v.([16]byte)).String()
		case pgtype.TextArrayOID, pgtype.VarcharArrayOID:
			// Text Array, e.g., TEXT[], format as a comma separated string
			arrayValues := []string{}
			for _, s := range v.([]any) {
				if s != nil {
					arrayValues = append(arrayValues, fmt.Sprint(s))
				}
			}
			valueString = strings.Join(arrayValues, ",")
		default:
			// Otherwise, do our best and just rely on the default string conversion
			// This is not appropriate for certain types. For example, if we did this to
			// a UUID, we would get a byte array string instead of the hex format.
			valueString = fmt.Sprint(v)
		}
		rowStrings = append(rowStrings, valueString)
	}

	return rowStrings, nil
}

func (svc *MainService) QueryAssets(ctx context.Context, orgID uuid.UUID, query string) (QueryAssetsResult, error) {

	// In a real situation, we'd use connection pooling. Here, for simplicity, we are just
	// opening and closing a new connection for each request.
	connUrl := svc.getDBUrl()
	conn, err := pgx.Connect(ctx, connUrl)
	if err != nil {
		return QueryAssetsResult{}, err
	}
	defer conn.Close(ctx)

	var result QueryAssetsResult

	role_suffix := getOrgHash(orgID)

	err = pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		// Before executing the query, set the role to the customer-specific role that is
		// set up. This might not exist if the org id has not been initialized yet.
		//
		// Even though we change the role, care needs to be taken because the SESSION role
		// is still the superuser. For example, if the user was allowed to run multiple
		// queries, they could change back to the superuser first.
		//
		// Careful for builtin Postgres functions as well. There might be functions that
		// can expose unwanted info if the user is running in a superuser session. There
		// aren't any that I'm aware of, but it's something to keep in mind.
		_, err := tx.Exec(ctx, `
			SET LOCAL ROLE customer_query_role_`+role_suffix+`;
		`)
		if err != nil {
			return fmt.Errorf("failed to set org_id; %w", err)
		}

		// We're doing a simple SQL replacement here. Ideally we would want to parse the
		// query and replace tokens properly, but that has a high overhead of complexity
		// with a large build dependency (Postgres).
		//
		// For now, we're just instructing the model to query from tbl_assets, and then
		// using that as a replacement token to query their real table.
		query = strings.ReplaceAll(query, "tbl_assets", "assets_org_"+role_suffix)

		// If query parsing/manipulation is implemented later, we should also enforce
		// LIMIT 16. (truncation limit plus 1)

		rows, err := tx.Query(ctx, query)
		if err != nil {
			// If the query fails, there is likely a syntax error from the model. We want
			// the model to correct itself. When we return errQueryFailed, the response to
			// the model will not be treated as a system error, but rather as a message
			// that tells the model that they made a mistake.
			//
			// If the query fails from a connection error, that can also be returned to
			// the model, informing it that the query could not be executed, and the model
			// will act accordingly.
			//
			// This is a basic example. When returning raw errors, care needs to be taken
			// to ensure that nothing sensitive is leaked to the user.
			return fmt.Errorf("%w; %w", ErrQueryFailed, err)
		}
		defer rows.Close()
		for rows.Next() {
			values, err := formatPGRow(rows)
			if err != nil {
				return err
			}

			// Truncate the results at 15 rows. Ideally, we would want to enforce a LIMIT
			// on the query itself, but that is complex to do properly, needing to parse
			// Postgres queries. This is a less performant solution that should work for
			// most situations.
			//
			// We could also tell the model to apply LIMIT 16 where possible (plus one to
			// detect truncation at 15), but that is additional unwanted context, so we
			// don't recommend doing that. The proper way to do it is parsing and
			// modifying the query.
			result.Rows = append(result.Rows, values)
			if len(result.Rows) >= 15 {
				result.Truncated = true
				rows.Close()
				break
			}
		}
		err = rows.Err()
		if err != nil {
			return fmt.Errorf("%w; %w", ErrQueryFailed, err)
		}

		// The response to the model is a raw CSV format with a header row. The model
		// understands this fairly well, but it may be worth testing other formats, such
		// as JSON, XML, etc. Those incur higher context overhead, but can make it easier
		// to distinguish values in the response.
		fields := rows.FieldDescriptions()
		for _, f := range fields {
			result.Columns = append(result.Columns, f.Name)
		}

		return nil
	})

	if err != nil {
		return QueryAssetsResult{}, err
	}

	return result, nil
}
