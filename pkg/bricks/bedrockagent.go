package bricks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
)

type BedrockInvokeInput = types.FunctionInvocationInput
type BedrockInvokeOutput = types.FunctionResult

// BedrockAgent is an agent implementation that uses AWS Bedrock Agent Runtime.
type BedrockAgent struct {
	Config BedrockAgentConfig
}

// Configuration input for a BedrockAgent, passed to NewBedrockAgent.
type BedrockAgentConfig struct {
	AgentName   string
	Model       string
	Instruction string
	Functions   *FunctionSet
}

// AWS Client
var bedrockAgentRuntimeClient *bedrockagentruntime.Client

// Returns the AWS client singleton.
func getBedrockAgentRuntime() *bedrockagentruntime.Client {
	if bedrockAgentRuntimeClient == nil {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Fatalf("Failed to load AWS config: %v", err)
		}

		client := bedrockagentruntime.NewFromConfig(cfg)
		bedrockAgentRuntimeClient = client
	}

	return bedrockAgentRuntimeClient
}

// Create a new agent that uses the Bedrock Agent Runtime.
func NewBedrockAgent(config BedrockAgentConfig) Agent {
	return &BedrockAgent{
		Config: config,
	}
}

// Take function parameter input from Bedrock RETURN_CONTROL and marshal it into a JSON
// string.
func marshalBedrockFunctionParams(params []types.FunctionParameter) ([]byte, error) {

	jsonMap := make(map[string]any)
	for _, param := range params {
		if param.Value == nil || param.Name == nil || param.Type == nil {
			return nil, fmt.Errorf("%w; fields must not be nil, name=%v type=%v value=%v",
				ErrInvalidArg, param.Name, param.Type, param.Value)
		}

		switch types.ParameterType(*param.Type) {
		case types.ParameterTypeString:
			jsonMap[*param.Name] = *param.Value
		case types.ParameterTypeInteger:
			jsonMap[*param.Name], _ = strconv.Atoi(*param.Value)
		case types.ParameterTypeBoolean:
			jsonMap[*param.Name] = *param.Value == "true"
		case types.ParameterTypeNumber:
			jsonMap[*param.Name], _ = strconv.ParseFloat(*param.Value, 64)
		case types.ParameterTypeArray:
			jsonMap[*param.Name] = *param.Value // TODO
		default:
			jsonMap[*param.Name] = *param.Value
		}
	}

	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal function params; %w", err)
	}
	return jsonBytes, nil
}

// Invoke one of our functions from a RETURN_CONTROL flow and return the output.
func (ba *BedrockAgent) invokeFunction(ctx context.Context, input BedrockInvokeInput) (BedrockInvokeOutput, error) {
	fs := ba.Config.Functions
	out := types.FunctionResult{
		ActionGroup:   input.ActionGroup,
		ResponseState: types.ResponseStateFailure,
		Function:      input.Function,
	}

	if input.ActionGroup == nil || input.Function == nil {
		return out, fmt.Errorf("%w; missing required params", ErrInvalidArg)
	}

	bodyBytes, err := marshalBedrockFunctionParams(input.Parameters)
	if err != nil {
		return out, fmt.Errorf("failed to marshal function params: %w", err)
	}

	result, err := fs.Invoke(ctx, *input.ActionGroup, *input.Function, bodyBytes)
	if err != nil {
		return out, fmt.Errorf("function %s.%s failed: %w",
			*input.ActionGroup, *input.Function, err)
	}

	switch v := result.(type) {
	case string:
		out.ResponseState = ""
		out.ResponseBody = map[string]types.ContentBody{
			"TEXT": {
				Body: aws.String(v),
			},
		}
	default:
		resultBytes, err := json.Marshal(v)
		if err != nil {
			return out, fmt.Errorf("failed to marshal function result: %w", err)
		}
		out.ResponseState = ""
		out.ResponseBody = map[string]types.ContentBody{
			"TEXT": {
				Body: aws.String(string(resultBytes)),
			},
		}
	}

	return out, nil
}

func (ba *BedrockAgent) makeBaseInput(sessionID string) bedrockagentruntime.InvokeInlineAgentInput {
	input := bedrockagentruntime.InvokeInlineAgentInput{
		SessionId:       aws.String(sessionID),
		FoundationModel: aws.String(ba.Config.Model),
		Instruction:     aws.String(ba.Config.Instruction),
		AgentName:       aws.String(ba.Config.AgentName),
	}
	if ba.Config.Functions != nil {
		input.ActionGroups = ba.Config.Functions.GetActionGroups()
	}
	return input
}

// Query the Bedrock Agent with the given prompt and return the response.
func (ba *BedrockAgent) Query(ctx context.Context, inputText string, sessionID string) (string, error) {

	client := getBedrockAgentRuntime()

	input := ba.makeBaseInput(sessionID)
	input.InputText = aws.String(inputText)

	// Invoke the agent with the input text.
	result, err := client.InvokeInlineAgent(ctx, &input)
	if err != nil {
		return "", fmt.Errorf("failed to invoke agent: %w", err)
	}

	var chunks []string

	agentResponse := result.GetStream()
	for {
		ev, ok := <-agentResponse.Events()
		if !ok {
			if agentResponse.Err() != nil {
				return "", fmt.Errorf("error receiving agent response: %w", agentResponse.Err())
			}
			break
		}

		switch v := ev.(type) {
		case *types.InlineAgentResponseStreamMemberChunk:
			// Record output chunks.
			chunks = append(chunks, string(v.Value.Bytes))
		case *types.InlineAgentResponseStreamMemberReturnControl:
			// If we get RETURN_CONTROL, call our functions and the invoke the agent again.
			var results []types.InvocationResultMember
			for _, rawInv := range v.Value.InvocationInputs {
				switch inv := rawInv.(type) {
				case *types.InvocationInputMemberMemberFunctionInvocationInput:
					resultValue, err := ba.invokeFunction(ctx, inv.Value)
					if err != nil {
						fmt.Println("Function invocation error:", err)
					}
					result := types.InvocationResultMemberMemberFunctionResult{
						Value: resultValue,
					}
					results = append(results, &result)
				default:
					log.Fatalf("Unsupported invocation input type")
				}
			}

			input := ba.makeBaseInput(sessionID)
			input.InlineSessionState = &types.InlineSessionState{
				InvocationId:                   v.Value.InvocationId,
				ReturnControlInvocationResults: results,
			}
			result, err := client.InvokeInlineAgent(context.TODO(), &input)
			if err != nil {
				return "", fmt.Errorf(
					"failed to invoke inline agent for return control: %v", err)
			}
			agentResponse = result.GetStream()
		default:
			fmt.Printf("Unexpected event type: %T\n", v)
		}
	}

	return strings.Join(chunks, ""), nil
}
