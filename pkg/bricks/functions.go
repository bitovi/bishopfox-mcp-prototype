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

// A function is an invokable action that is exposed in the model context schema.
type Function struct {
	// Name of the function, visible to the model.
	Name string
	// Description of the function, visible to the model.
	Description string
	// Parameter structure. This should be an empty instance or a nil pointer to a struct
	// that will be reflected over for the input schema.
	Params any
	// The handler that implements the function.
	Handler FunctionHandler
}

// Signature for callable functions via the model context.
type FunctionHandler func(FunctionContext) (any, error)

// A set of functions organized into groups.
type FunctionSet struct {
	// Name of the function set. Should not use spaces or special characters other than
	// underscores.
	Name      string
	Functions map[string]Function
}

// Create a new function set with the given name.
func NewFunctionSet(name string) *FunctionSet {
	return &FunctionSet{
		Name:      name,
		Functions: make(map[string]Function),
	}
}

// Convert the given function information into a Bedrock function definition.
func createBedrockFunctionDefinition(fn Function) types.FunctionDefinition {
	var def types.FunctionDefinition
	def.Name = &fn.Name
	def.Description = &fn.Description
	def.Parameters = make(map[string]types.ParameterDetail)
	t := reflect.TypeOf(fn.Params)
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

// Convert the function set into a Bedrock Agent Action Group configuration. In the future
// this could be expanded to detect what functions are actually relevant to the user's
// query, and only use that subset of functions to optimize performance.
//
// Bedrock Agents are limited to a max of 11 actions. Too many actions can easily confuse
// the model on what its purpose is.
func (fs *FunctionSet) GetActionGroup() types.AgentActionGroup {
	var functions []types.FunctionDefinition
	for _, function := range fs.Functions {
		def := createBedrockFunctionDefinition(function)
		functions = append(functions, def)
	}

	actionGroup := types.AgentActionGroup{
		ActionGroupName: aws.String(fs.Name),
		FunctionSchema: &types.FunctionSchemaMemberFunctions{
			Value: functions,
		},
		ActionGroupExecutor: &types.ActionGroupExecutorMemberCustomControl{
			Value: types.CustomControlMethodReturnControl,
		},
	}

	return actionGroup
}

// Add a function. The name and description describe the function to the model. The params
// should be an empty struct instance e.g., MyParamsStruct{}, used for reflection only.
func (fs *FunctionSet) AddFunction(name string, description string, params any, handler FunctionHandler) {

	fn := Function{
		Name:        name,
		Description: description,
		Params:      params,
		Handler:     handler,
	}

	fs.Functions[name] = fn
}

// Error if the function or group doesn't exist.
var ErrNoFunction = errors.New("no function found")

// Invoke a function by group and name with the given JSON marshalled input.
func (fs *FunctionSet) Invoke(ctx context.Context, function string, input []byte) (any, error) {
	fn, ok := fs.Functions[function]
	if !ok {
		return nil, fmt.Errorf("%w; function %s is not defined", ErrNoFunction, function)
	}

	// Call the handler. We pass the input into a FunctionContext which the functions can
	// use to bind parameters.
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
