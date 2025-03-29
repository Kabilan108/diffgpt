package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

type Commit struct {
	Message string `json:"message" jsonschema_description:"Single-line commit message that adheres to conventional commits."`
}

type DetailedCommit struct {
	Message string `json:"message" jsonschema_description:"Single-line commit message that adheres to conventional commits."`
	Details string `json:"details" jsonschema_description:"Description of the changes made, written as concise bullet points in markdown"`
}

func GenerateSchema[T any]() any {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

func newResponseSchema[T any](name string, desc string) openai.ChatCompletionNewParamsResponseFormatUnion {
	jsonSchema := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        name,
		Description: openai.String(desc),
		Schema:      GenerateSchema[T](),
		Strict:      openai.Bool(true),
	}
	return openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
			JSONSchema: jsonSchema,
		},
	}
}

func NewClient(apiKey, baseURL string) openai.Client {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
	return client
}

func Generate[T Commit | DetailedCommit](
	ctx context.Context, client openai.Client, model, schemaName, schemaDesc, prompt, systemPrompt string,
) (T, error) {
	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(prompt),
		},
		ResponseFormat: newResponseSchema[T](schemaName, schemaDesc),
		Model:          shared.ChatModel(model),
	})
	if err != nil {
		var zero T
		return zero, err
	}

	var resp T
	err = json.Unmarshal([]byte(completion.Choices[0].Message.Content), &resp)
	if err != nil {
		var zero T
		return zero, err
	}

	return resp, nil
}

func GenerateCommitMessage(
	ctx context.Context, client openai.Client, model, diff string, detailed bool,
) (string, error) {
	systemMessage := `You are an expert programmer assisting with writing git commit messages.
Analyze the provided code diff and generate a concise, informative commit messag following
conventional commit standards (e.g., "feat: add user login functionality").
The commit message should accurately describe the changes. Do not include explanations or apologies.
`
	userMessage := fmt.Sprintf("Generate a commit message for the following diff:\n```diff\n%s\n```", diff)

	if detailed {
		r, err := Generate[DetailedCommit](
			ctx, client, model, "detailed_commit",
			"a git commit message with a description of the changes made",
			userMessage, systemMessage,
		)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s\n\n%s", r.Message, r.Details), nil
	}

	r, err := Generate[Commit](
		ctx, client, model, "commit", "a git commit message", userMessage, systemMessage,
	)
	if err != nil {
		return "", err
	}
	return r.Message, nil
}
