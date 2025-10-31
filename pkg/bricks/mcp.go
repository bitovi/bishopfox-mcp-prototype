package bricks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
)

func BindFunctionsToMCPServer(fs *FunctionSet, group string, s *server.MCPServer) error {

	fg, ok := fs.Groups[group]
	if !ok {
		return fmt.Errorf("group not defined: %s", group)
	}

	for _, fn := range fg.Functions {

		var toolOpts []mcp.ToolOption
		t := reflect.TypeOf(fn.Params)
		for i := range t.NumField() {
			field := t.Field(i)
			fieldType := field.Type.Kind()
			name := field.Tag.Get("json")
			if name == "" {
				continue
			}
			required := field.Tag.Get("required") == "true"
			desc := field.Tag.Get("desc")
			var props []mcp.PropertyOption

			if desc != "" {
				props = append(props, mcp.Description(desc))
			}
			if required {
				props = append(props, mcp.Required())
			}

			if fieldType == reflect.String {
				toolOpts = append(toolOpts, mcp.WithString(name, props...))
			} else if fieldType == reflect.Int || fieldType == reflect.Int32 || fieldType == reflect.Int64 {
				toolOpts = append(toolOpts, mcp.WithNumber(name, props...))
			} else if fieldType == reflect.Bool {
				toolOpts = append(toolOpts, mcp.WithBoolean(name, props...))
			} else if fieldType == reflect.Float64 {
				toolOpts = append(toolOpts, mcp.WithNumber(name, props...))
			} else {
				return fmt.Errorf("unsupported field type: %s", fieldType.String())
			}
		}

		toolOpts = append(toolOpts, mcp.WithDescription(fn.Description))
		tool := mcp.NewTool(fn.Name, toolOpts...)

		s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Build the input struct
			args := request.GetArguments()
			jsonBytes, err := json.Marshal(args)
			log.Debugln("MCP bound input:", string(jsonBytes))
			if err != nil {
				log.Errorln("function invocation error (marshaling input)", err)
				return mcp.NewToolResultError("function invocation failed; invalid input"), nil
			}

			result, err := fs.Invoke(ctx, group, fn.Name, jsonBytes)
			if err != nil {
				if errors.Is(err, ErrInvalidArg) {
					return mcp.NewToolResultError(fmt.Sprintf("function invocation failed; %v", err)), nil
				} else {
					log.Errorln("function invocation error", err)
					return mcp.NewToolResultError("function invocation failed unexpectedly"), nil
				}
			}

			switch result := result.(type) {
			case string:
				return mcp.NewToolResultText(result), nil
			default:
				jsonBytes, err := json.Marshal(result)
				if err != nil {
					log.Errorln("function invocation error (marshaling result)", err)
					return mcp.NewToolResultError("function invocation failed!"), nil
				}
				return mcp.NewToolResultText(string(jsonBytes)), nil
			}
		})
	}

	return nil
}
