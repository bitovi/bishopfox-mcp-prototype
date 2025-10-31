package bricks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
)

// A function group is a group of functions with a description. The name and description
// might not be used in the model context.
type FunctionGroup struct {
	Name        string
	Description string
	Functions   map[string]Function
}

// A function is an invokable action that will be shared in the model context schema.
type Function struct {
	Name        string
	Description string
	Params      any
	Handler     FunctionHandler
}

// Signature for callable functions via the model context.
type FunctionHandler func(FunctionContext) (any, error)

// A set of functions organized into groups.
type FunctionSet struct {
	Groups map[string]FunctionGroup
}

// Create a new function set with an empty group mapping.
func NewFunctionSet() *FunctionSet {
	return &FunctionSet{
		Groups: make(map[string]FunctionGroup),
	}
}

// Convert the given function information into a Bedrock function definition.
func createBedrockFunctionDefinition(name string, description string, params any) types.FunctionDefinition {
	var def types.FunctionDefinition
	def.Name = &name
	def.Description = &description
	def.Parameters = make(map[string]types.ParameterDetail)
	t := reflect.TypeOf(params)
	if t.Kind() != reflect.Struct {
		panic("T must be a struct")
	}

	for i := range t.NumField() {
		field := t.Field(i)
		fieldType := field.Type.Kind()
		var paramType types.ParameterType
		if fieldType == reflect.String {
			paramType = types.ParameterTypeString
		} else if fieldType == reflect.Int || fieldType == reflect.Int32 || fieldType == reflect.Int64 {
			paramType = types.ParameterTypeInteger
		} else if fieldType == reflect.Bool {
			paramType = types.ParameterTypeBoolean
		} else if fieldType == reflect.Float64 {
			paramType = types.ParameterTypeNumber
		} else if fieldType == reflect.Slice {
			if field.Type.Elem().Kind() == reflect.String {
				paramType = types.ParameterTypeArray
			} else {
				panic("unsupported field type for " + field.Name + ": only []string is supported")
			}
		} else {
			panic("unsupported field type: " + fieldType.String())
		}

		paramDetail := types.ParameterDetail{
			Description: aws.String(field.Tag.Get("desc")),
			Required:    aws.Bool(field.Tag.Get("required") == "true"),
			Type:        paramType,
		}

		name := field.Tag.Get("json")

		def.Parameters[name] = paramDetail
	}

	return def
}

// Convert the function set into a list of Bedrock Agent Action Groups for configuration.
func (fs *FunctionSet) GetActionGroups() []types.AgentActionGroup {
	var actionGroups []types.AgentActionGroup
	for _, group := range fs.Groups {
		var functions []types.FunctionDefinition
		for _, function := range group.Functions {
			def := createBedrockFunctionDefinition(function.Name, function.Description, function.Params)
			functions = append(functions, def)
		}

		actionGroup := types.AgentActionGroup{
			ActionGroupName: aws.String(group.Name),
			Description:     aws.String(group.Description),
			FunctionSchema: &types.FunctionSchemaMemberFunctions{
				Value: functions,
			},
			ActionGroupExecutor: &types.ActionGroupExecutorMemberCustomControl{
				Value: types.CustomControlMethodReturnControl,
			},
		}
		actionGroups = append(actionGroups, actionGroup)
	}
	return actionGroups
}

// Add a new function group. The name and description might not be used in the model
// context, depending on the implementation.
func (fs *FunctionSet) AddGroup(name string, description string) {
	group := FunctionGroup{
		Name:        name,
		Description: description,
		Functions:   make(map[string]Function),
	}
	fs.Groups[name] = group
}

// Add a function with input schema T.
func (fs *FunctionSet) AddFunction(group string, name string, description string, params any, handler FunctionHandler) {

	fn := Function{
		Name:        name,
		Description: description,
		Params:      params,
		Handler:     handler,
	}

	fs.Groups[group].Functions[name] = fn
}

// Error if the function or group doesn't exist.
var ErrNoFunction = errors.New("no function found")

// Invoke a function by group and name with the given JSON marshalled input.
func (fs *FunctionSet) Invoke(ctx context.Context, group string, function string, input []byte) (any, error) {
	fg, ok := fs.Groups[group]
	if !ok {
		return nil, fmt.Errorf("%w; group %s is not defined", ErrNoFunction, group)
	}

	fn, ok := fg.Functions[function]
	if !ok {
		return nil, fmt.Errorf("%w; function %s.%s is not defined", ErrNoFunction, group, function)
	}

	// Call the handler
	fct := FunctionContextFromJSON(ctx, input)
	return fn.Handler(fct)
}

// FunctionContext carries user data and call parameters for a function invocation. Maybe
// we should change this to a context value though, rather than making a new type.
type FunctionContext struct {
	context.Context
	Input []byte
}

// Wrap the given JSON input in a FunctionContext for passing to handlers.
func FunctionContextFromJSON(ctx context.Context, input []byte) FunctionContext {
	return FunctionContext{
		Input:   input,
		Context: ctx,
	}
}

// Bind the input to the given structure. Panics if unmarshalling fails.
func (c *FunctionContext) MustBind(out any) {
	err := json.Unmarshal(c.Input, out)
	if err != nil {
		panic("failed to unmarshal request json: " + err.Error())
	}
}
