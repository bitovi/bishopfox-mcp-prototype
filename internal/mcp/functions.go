// Maybe a poor name for this package. This package provides tools and prompts for model
// context. It isn't necessarily "MCP" specific.
package mcp

import (
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bitovi/bishopfox-mcp-prototype/internal/service"
	"github.com/bitovi/bishopfox-mcp-prototype/pkg/bricks"

	log "github.com/sirupsen/logrus"
)

type Service = service.Service

// This file defines functions that can be invoked by AI models. The request and response
// schema is defined using structs and reflection.
//
// These functions are somewhat equivalent to API endpoints. When the service is the host,
// it is calling back to itself. When MCP is used, the MCP server is exposing these
// functions as tools, which is very similar to API endpoints.
//
// Server Host:
// Server --(Invokes Bedrock)-> Model --(Return Control))-> Tool Function --(Returns Result)-> Model ...
//
// External Host:
// Client --(Invokes MCP Tool)-> Tool Function --(Returns Result)-> Client ...

var ErrMissingContext = errors.New("missing required context value")

// This is an example request schema.
//   - json: Name of the field exposed in the MCP.
//   - desc: Description shown in the MCP tool listing. This is typically a shorter
//     description as you can leave the main instructions in the upper
//     function-level description.
//   - required: Whether the field is required. Set to 'true' to reflect in the MCP schema
//     that the field is required.
type GetWeatherRequest struct {
	// Concise descriptions like these convey the usage effectively to LLMs. Examples are
	// always great, just like in real life. If something is simple enough to be described
	// in a minimal example like this, do so.
	City string `json:"city" desc:"e.g. Chicago" required:"true"`
}

// This is an example response schema. The MCP supports output schemas. Bedrock Action
// Groups do not. Currently, we are not implementing MCP output schemas (see mcp.go). We
// believe that output schemas are not entirely useful for LLMs.
//
// When using non-string results with Bedrock, our lib is just marshaling to JSON and
// using that as the text output. We do the same thing with our MCP.
type GetWeatherResponse struct {
	Temperature string `json:"temperature"`
	Condition   string `json:"condition"`
}

// Example function implementation. When an LLM makes a tool request, it is routed to
// these handlers. The flow for Bedrock invoking a function hosted by the client
// application is referred to as RETURN_CONTROL.
func GetWeatherFunction(c bricks.FunctionContext) (any, error) {
	var req GetWeatherRequest
	c.MustBind(&req)

	// The query context allows us to pass data from the service or MCP server into the
	// function execution. This is where we can forward things like authentication
	// information (e.g., what orgs the user has access to) and the organization_id.
	qc := service.MustGetQueryContext(c)

	log.Debugln("Authorization:", qc.Authorization)

	// Simulate a weather API call. Our package doesn't support structured outputs, but
	// this is what a structured output might look like in the handler.
	return GetWeatherResponse{
		Temperature: "72°F",
		Condition:   "Sunny",
	}, nil
}

// This is our input for query_assets, it accepts a raw SQL query.
//
// An AI fundamental to keep in mind is that this query can be literally anything. It
// cannot be restricted at the prompt level, so we need to execute it securely with no
// assumptions.
//
// The output is text, so there is no output struct defined (and we don't support
// structured outputs anyway).
type QueryAssetsRequest struct {
	Query string `json:"query" desc:"The SQL query to execute" required:"true"`
}

// Handler for query_assets. Queries the asset database with a given SQL query and returns
// the result. Enforces security so that the calling organization can only access its own
// data.
func QueryAssetsFunction(c bricks.FunctionContext) (any, error) {
	var req QueryAssetsRequest
	c.MustBind(&req)
	qc := service.MustGetQueryContext(c)
	svc := qc.Service

	// Test print of auth token and incoming query.
	log.Debugln("Authorization:", qc.Authorization)
	log.Debugf("--- Received query request ---\n%s\n---", req.Query)

	result, err := svc.QueryAssets(c, qc.OrgID, req.Query)
	if errors.Is(err, service.ErrQueryFailed) {
		// Treat query failed as a message to the model; don't break the execution. We
		// want the model to see the error in case it can correct itself (it can
		// correct syntax errors).
		log.Debugln("Query failed:", err)
		return "The query failed to execute with this error:\n" + err.Error(), nil
	} else if err != nil {
		return nil, fmt.Errorf("db query failed; %w", err)
	}

	// Format the results when the AI is querying the asset database. We've seen
	// decent results with this "CSV" type of output. While we could benefit from
	// formatting certain fields in certain ways, we can't depend on any field names,
	// since the model might declare aliases.
	//
	// For best results, the fields themselves should contain values that are
	// meaningful without reading the CSV header. Take for example this output:
	//
	//  service|example.com|80
	//
	// It is likely that this means that "the service example.com is running on port
	// 80", so the model wouldn't need to see "service|domain|port" for context.
	// That's not to say the header should be omitted. It just helps to have the
	// context as clear as possible.
	//
	// There is room for more experimentation of what is the best format to use when
	// returning results for the model.
	outputStrings := []string{}
	outputStrings = append(outputStrings, strings.Join(result.Columns, "|"))
	for _, row := range result.Rows {
		outputStrings = append(outputStrings, strings.Join(row, "|"))
	}
	if result.Truncated {
		outputStrings = append(outputStrings, "-- Result set truncated; more than 15 rows returned --")
	}
	if len(result.Rows) == 0 {
		outputStrings = append(outputStrings, "-- No results --")
	}

	return strings.Join(outputStrings, "\n"), nil
}

// The describe_asset function is intended to query an asset and related assets
// programatically and format as a natural language response for the model. We feel this
// might be useful, but haven't implemented a prototype yet.
type DescribeAssetRequest struct {
	AssetID string `json:"asset_id" desc:"ID of the asset to describe" required:"true"`
}

/********
func (svc *Service) DescribeAssetFunction(c bricks.FunctionContext) (any, error) {
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
********/

// get_assets_overview_link is a function that generates links to the web UI for viewing
// assets. Some rationale behind the schema:
//
//   - The name itself tries to describe the purpose of the function clearly. The asset
//     pages are like "overview" pages.
//   - The asset_type parameter's "one of" language makes it clear what valid options
//     there are.
//   - The filter and search are more complex inputs described further in the tool
//     description. The parameter description can exist in either place, the tool
//     description or the parameter description. See get_assets_overview_link_desc.txt.
//     Since we're using Go reflection, it's aesthetically better to have longer parameter
//     details in a separate place instead of stuffed into a struct tag.
type GetAssetsOverviewLinkRequest struct {
	AssetType string `json:"asset_type" desc:"Type of asset to view, one of [domain, subdomain, ip, service, network, webapp]" required:"true"`
	Filters   string `json:"filters" desc:"Optional filters to apply, formatted as URL params, e.g. key1=value1&key2=value2" required:"false"`
	Search    string `json:"search" desc:"Optional search term to filter assets" required:"false"`
}

// Interestingly, this function doesn't really require the service. All of the work here
// is input translation, so there is nothing to delegate to the service layer.
func GetAssetsOverviewLinkFunction(c bricks.FunctionContext) (any, error) {
	var req GetAssetsOverviewLinkRequest
	c.MustBind(&req)
	qc := service.MustGetQueryContext(c)

	log.Debugf("--- Received get assets overview link request ---\n%+v\n---", req)

	// We can pull the organization_id out of the query context to be used in the URL.
	baseUrl := fmt.Sprintf("https://ui.api.non.usea2.bf9.io/%s/assets/%s",
		qc.OrgID.String(), req.AssetType)

	// Note about these translations. See the instructions for this function. We are
	// avoiding using technical URL params in favor of ubiquitous language.
	//
	// For example, LLMs are not good at any kind of math, especially date math, so
	// instead of taking raw dates for the "expiry" filter, we take the relative time in
	// days and then programatically translate it to the real query param.
	timeFormat := "2006-01-02T15:04:05.000Z"
	filters, err := url.ParseQuery(req.Filters)
	if err != nil {
		return "(Error) Invalid filters format. It needs to be formatted as a valid URL query string.", nil
	}
	params := url.Values{}
	for key, values := range filters {
		key = strings.ToLower(key)
		switch key {
		case "expiry":
			// The input is LLM friendly: amount of days
			// The output here is the range between now and the expiry date as a between filter.
			// Or in the case of "0" days, we just set it to "before now".
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
			// The language "domainExtension" seems foreign to me. We call this filter
			// param "tld" instead and translate it here.
			domains := strings.Split(values[0], " ")
			for i := range domains {
				domains[i] = "." + domains[i]
			}
			params.Set("domainExtension", strings.Join(domains, ","))
		case "tags":
			tags := strings.Split(values[0], " ")
			params.Set("tags", strings.Join(tags, ","))
		default:
			// Any other filters are passed as-is, but typically this should not happen
			// unless the user explicitly tells the model to use the specific
			// filter key.
			for _, value := range values {
				params.Add(key, value)
			}
		}
	}

	// In Cosmos, the search seems to be a single raw term only (spaces are included in
	// the term). The tool instructions elaborate on that for the model so it doesn't try
	// to use multiple keywords.
	if req.Search != "" {
		params.Set("search", req.Search)
	}

	fullURL, _ := url.Parse(baseUrl)
	fullURL.RawQuery = params.Encode()
	return fullURL.String(), nil
}

// Example parameterless function. get_latest_emerging_threats is a mock function that
// returns a list of emerging_threat data for testing against.
type GetLatestEmergingThreatsRequest struct{}

func GetLatestEmergingThreatsFunction(c bricks.FunctionContext) (any, error) {
	// For demonstration purposes, returning a static list of threats.
	//
	// Rationale behind this function
	// - The tool title takes care of most of the usage context.
	// - The tool description contains a little context of what an "emerging threat" is.
	//   - The model has otherwise no idea what that is.
	// - Uses a natural language header to declare what is being returned.
	// - Claude is known to prefer XML tags around blocks of data or instructions, so we
	//   have <threat> blocks.
	// - Each result contains only the most relevant fields for user queries and LLM
	//   context, excluding many fields that are less useful, for example, created_by,
	//   percent_orgs_affected, description, remediation, etc.
	//   - The purpose of a listing function should be focused on high level details only,
	//     much like what you see in the Emerging Threats overview table.

	// Some fields might contain excessive data for the model. In those cases, we would
	// want to truncate the field. In the example below, we mark one such truncated field
	// with natural language "(truncated list...)".

	return `Here is information about the 5 latest emerging threats:
<threat>
cpe: n/a
cve: CVE-2025-59118
cvss_score: 0
cwe: n/a
identified_at: 2025-11-12T22:07:31.62Z
investigation_status: in-progress
products_affected: Apache OFBiz
id: et-00171
technology: Apache Software Foundation
tier: 2
title: Apache OFBiz: Critical Remote Command Execution via Unrestricted File Upload
</threat>
<threat>
cpe: n/a
cve: CVE-2025-12480
cvss_score: 0
cwe: n/a
identified_at: 2025-11-12T00:00:00Z
investigation_status: complete
products_affected: TrioFox
id: et-00170
technology: TrioFox
tier: 3
title: Glaidnet Triofox Improper Access Control
</threat>
<threat>
cpe: cpe:2.3:a:wpexperts:post_smtp:-:*:*:*:*:wordpress:*:*
cve: CVE-2025-11833
cvss_score: 9.8
cwe: CWE-862
identified_at: 2025-11-04T00:00:00Z
investigation_status: in-progress
products_affected: Post SMTP Plugin
id: et-00169
technology: WordPress
tier: 2
title: WordPress Post SMTP plugin Account Takeover
</threat>
<threat>
cpe: n/a
cve: CVE-2025-59287
cvss_score: 0
cwe: n/a
identified_at: 2025-10-16T00:00:00Z
investigation_status: complete
products_affected: Windows Server 2012, Windows Server 2012 R2, Windows Server 2012 R2 (Server Core installation) (truncated list...)
id: et-00168
technology: Microsoft
tier: 3
title: Windows Server Update Service (WSUS) Remote Code Execution Vulnerability
</threat>
<threat>
cpe: n/a
cve: CVE-2025-53521, CVE-2025-58153, CVE-2025-59478, CVE-2025-54479, CVE-2025-53860, CVE-2025-58424
cvss_score: 8
cwe: n/a
identified_at: 2025-10-15T09:00:00Z
investigation_status: complete
products_affected: BIG-IP
id: et-00167
technology: F5
tier: 3
title: F5 Review and Response
</threat>
`, nil
}

//go:embed prompt/query_assets_desc.txt
var queryAssetsDesc string

//go:embed prompt/query_assets_extended_desc.txt
var queryAssetsExtendedDesc string

//go:embed prompt/get_assets_overview_link_desc.txt
var getAssetsOverviewLinkDesc string

//go:embed prompt/get_latest_emerging_threats_desc.txt
var getLatestEmergingThreatsDesc string

// This is just like routing in an API.
func GetFunctions(svc Service) *bricks.FunctionSet {
	fs := bricks.NewFunctionSet("bishopfox")
	fs.AddFunction("query_assets", queryAssetsDesc, queryAssetsExtendedDesc,
		QueryAssetsRequest{}, QueryAssetsFunction)
	fs.AddFunction("get_assets_overview_link", getAssetsOverviewLinkDesc, "",
		GetAssetsOverviewLinkRequest{}, GetAssetsOverviewLinkFunction)
	fs.AddFunction("get_latest_emerging_threats", getLatestEmergingThreatsDesc, "",
		GetLatestEmergingThreatsRequest{}, GetLatestEmergingThreatsFunction)
	return fs
}
