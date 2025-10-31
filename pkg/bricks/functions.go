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

type FunctionGroup struct {
	Name        string
	Description string
	Functions   map[string]Function
}

type Function struct {
	Name        string
	Description string
	Params      any
	Handler     FunctionHandler
}

type FunctionHandler func(FunctionContext) (any, error)

type FunctionSet struct {
	Groups map[string]FunctionGroup
}

func NewFunctionSet() *FunctionSet {
	return &FunctionSet{
		Groups: make(map[string]FunctionGroup),
	}
}

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

var ErrNoFunction = errors.New("no function found")

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

// FunctionContext carries user data and call parameters for a function invocation.
type FunctionContext struct {
	context.Context
	Input []byte
}

// Wrap the given JSON text in a FunctionContext.
func FunctionContextFromJSON(ctx context.Context, input []byte) FunctionContext {
	return FunctionContext{
		Input:   input,
		Context: ctx,
	}
}

func (c *FunctionContext) MustBind(out any) {
	err := json.Unmarshal(c.Input, out)
	if err != nil {
		panic("failed to unmarshal request json: " + err.Error())
	}
}
