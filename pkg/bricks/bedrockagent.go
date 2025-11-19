package bricks

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
)

type BedrockInvokeInput = types.FunctionInvocationInput
type BedrockInvokeOutput = types.FunctionResult

// BedrockAgent is an agent implementation that uses AWS Bedrock Agent Runtime. The
// configuration is applied "inline", meaning you don't need to create an agent resource
// before invoking it. This allows you to do flexible/dynamic configuration per query.
type BedrockAgent struct {
	Config BedrockAgentConfig
}

// Configuration input for a BedrockAgent, passed to NewBedrockAgent.
type BedrockAgentConfig struct {
	AgentName      string
	Model          string
	Instruction    string
	Functions      *FunctionSet
	Knowledgebases []types.KnowledgeBase
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
			// TODO: have not experimented much with this. I have seen wierd results with
			// array fields in the past, for example, the model arbitrarily restricting
			// itself to one entry only.
			jsonMap[*param.Name] = *param.Value
		default:
			// Any missed cases, just use the string directly.
			jsonMap[*param.Name] = *param.Value
		}
	}

	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal function params; %w", err)
	}
	return jsonBytes, nil
}

// Invoke one of our functions and return the output. This is called during the
// RETURN_CONTROL flow, i.e., when control is returned to our side from Bedrock to invoke
// a tool.
//
// When an error is returned, the BedrockInvokeOutput returned will have a FAILURE
// response state that can be sent to Bedrock.
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

	if *input.ActionGroup != fs.Name {
		return out, fmt.Errorf("%w; unknown action group: %s",
			ErrInvalidArg, *input.ActionGroup)
	}

	result, err := fs.Invoke(ctx, *input.Function, bodyBytes)
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

// Create a base inline configuration for invoking a Bedrock Agent.
func (ba *BedrockAgent) makeBaseInput(sessionID string) bedrockagentruntime.InvokeInlineAgentInput {
	input := bedrockagentruntime.InvokeInlineAgentInput{
		// The session ID can be any string you want. Using the same string again will
		// continue the same session (conversation history). The session is deleted after
		// a configurable amount of time of no activity. Default is 15 minutes.
		SessionId:       aws.String(sessionID),
		FoundationModel: aws.String(ba.Config.Model),
		Instruction:     aws.String(ba.Config.Instruction),
		AgentName:       aws.String(ba.Config.AgentName),
		KnowledgeBases:  ba.Config.Knowledgebases,
	}
	if ba.Config.Functions != nil {
		input.ActionGroups = []types.AgentActionGroup{
			ba.Config.Functions.GetActionGroup(),
		}
	}
	return input
}

// Query the Bedrock Agent with the given prompt.
func (ba *BedrockAgent) Query(ctx context.Context, inputText string, sessionID string) (QueryResult, error) {

	client := getBedrockAgentRuntime()

	// We're using "inline" agents, meaning that we configure them each time we want to
	// invoke them. This is more flexible than creating persistent agent resources in AWS
	// ahead of time. More friendly to code-driven agents.
	//
	// However, static agent configurations do have their use cases and can be a viable
	// alternative.
	input := ba.makeBaseInput(sessionID)

	// Invoke the agent with the input text.
	input.InputText = aws.String(inputText)
	result, err := client.InvokeInlineAgent(ctx, &input)
	if err != nil {
		return QueryResult{}, fmt.Errorf("failed to invoke agent: %w", err)
	}

	var chunks []string
	refs := []Reference{}
	used_refs := make(map[string]bool)

	agentResponse := result.GetStream()
	for {
		// The response may appear in multiple events, especially longer responses.
		ev, ok := <-agentResponse.Events()
		if !ok {
			if agentResponse.Err() != nil {
				return QueryResult{}, fmt.Errorf("error receiving agent response: %w", agentResponse.Err())
			}
			break
		}

		switch v := ev.(type) {
		case *types.InlineAgentResponseStreamMemberChunk:
			// A "chunk" is a section of the text response from the model.

			// We'll record all chunks and concatenate them at the end.
			chunks = append(chunks, string(v.Value.Bytes))

			// In attribution, we can check for citations used. Citations are references
			// to sources that were used for retrieval augmented generation. For example,
			// if it used a knowledgebase document, the citations will contain metadata
			// of what documents were used.
			//
			// We have added our own metadata fields to the knowledgebase documents to
			// help identify where they can be found in the web UI.
			if v.Value.Attribution != nil {
				for _, citation := range v.Value.Attribution.Citations {

					for _, ref := range citation.RetrievedReferences {
						meta := make(map[string]string)
						for k, val := range ref.Metadata {
							var text string
							_ = val.UnmarshalSmithyDocument(&text)
							meta[k] = string(text)
						}

						// Dedup refs, in case multiple chunks are used from same source.
						// Note this is only valid for knowledgebase metadata. If we start
						// to support other sources, then we'd want a different
						// deduplication strategy.
						refKey := fmt.Sprintf("%s/%s",
							meta["x-amz-bedrock-kb-data-source-id"],
							meta["x-amz-bedrock-kb-source-uri"])
						if used_refs[refKey] {
							continue
						}
						used_refs[refKey] = true

						refs = append(refs, Reference{
							Type: "knowledgebase",
							Data: meta,
						})
					}
				}
			}
		case *types.InlineAgentResponseStreamMemberReturnControl:
			// In AWS Bedrock, a "RETURN_CONTROL" flow is when the agent defers to the
			// caller to execute some function (tool) and provide the result before
			// continuing.
			//
			// Models can call multiple tools at once. The system prompt will typically
			// encourage calling multiple tools at once.
			//
			// So, go through the list of invocation inputs, call the desired functions,
			// and then build a result for the model to continue with. This is basically
			// submitted as a follow up message: e.g., the model is responding with a
			// question, and our side is responding with an answer.
			var results []types.InvocationResultMember
			for _, rawInv := range v.Value.InvocationInputs {
				switch inv := rawInv.(type) {
				case *types.InvocationInputMemberMemberFunctionInvocationInput:
					resultValue, err := ba.invokeFunction(ctx, inv.Value)
					if err != nil {
						log.Errorln("Function invocation error:", err)
						// Fallthrough: the resultValue contains a valid FAILURE state
						// that is sent back to the model.
					}
					result := types.InvocationResultMemberMemberFunctionResult{
						Value: resultValue,
					}
					results = append(results, &result)

					// We could also consider including this tool call and parameters as a
					// Reference in the output.
				default:
					log.Panicln("Unsupported invocation input type")
				}
			}

			// Invoke the agent again with the tool call results.
			input := ba.makeBaseInput(sessionID)
			input.InlineSessionState = &types.InlineSessionState{
				InvocationId:                   v.Value.InvocationId,
				ReturnControlInvocationResults: results,
			}
			result, err := client.InvokeInlineAgent(ctx, &input)
			if err != nil {
				return QueryResult{}, fmt.Errorf(
					"failed to invoke inline agent for return control: %v", err)
			}
			agentResponse = result.GetStream()
		default:
			fmt.Printf("Unexpected event type: %T\n", v)
		}
	}

	return QueryResult{
		Response: strings.Join(chunks, ""),
		Refs:     refs,
	}, nil
}
