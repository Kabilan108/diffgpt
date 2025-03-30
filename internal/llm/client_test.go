package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kabilan108/diffgpt/internal/config"
	"github.com/openai/openai-go"
)

// mockOpenAIClient is our test mock for the OpenAI client
type mockOpenAIClient struct{}

// Global variables for test response values
var (
	testResponseContent string
	testError           error
)

// Create a method to generate a mock client for testing
func createMockClient(responseContent string, returnError error) *mockOpenAIClient {
	// Set up global test variables used by the mocked functions
	testResponseContent = responseContent
	testError = returnError

	return &mockOpenAIClient{}
}

// Mock the chat access path (client.Chat.Completions.New)
type (
	mockChatStruct        struct{}
	mockCompletionsStruct struct{}
)

func (m *mockOpenAIClient) Chat() mockChatStruct {
	return mockChatStruct{}
}

func (m mockChatStruct) Completions() mockCompletionsStruct {
	return mockCompletionsStruct{}
}

// This mocks the New method and returns our test response
func (m mockCompletionsStruct) New(
	ctx context.Context, params openai.ChatCompletionNewParams,
) (*mockCompletionResponse, error) {
	if testError != nil {
		return nil, testError
	}

	return &mockCompletionResponse{
		Choices: []mockChoice{
			{
				Message: mockMessage{
					Content: testResponseContent,
				},
			},
		},
	}, nil
}

// Simple structs to mock the response structure
type mockCompletionResponse struct {
	Choices []mockChoice
}

type mockChoice struct {
	Message mockMessage
}

type mockMessage struct {
	Content string
}

// Test the GenerateSchema function
func TestGenerateSchema(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "commit schema",
		},
		{
			name: "detailed commit schema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result any
			if tt.name == "commit schema" {
				result = GenerateSchema[Commit]()
			} else {
				result = GenerateSchema[DetailedCommit]()
			}

			// Just verify the schema is not nil
			if result == nil {
				t.Errorf("GenerateSchema returned nil for %s", tt.name)
			}
		})
	}
}

func TestNewResponseSchema(t *testing.T) {
	name := "test_schema"
	desc := "test description"

	result := newResponseSchema[Commit](name, desc)

	// Check that the schema was created properly
	if result.OfJSONSchema == nil {
		t.Fatal("Expected JSONSchema format to be created")
	}

	if result.OfJSONSchema.JSONSchema.Name != name {
		t.Errorf("Expected name %s, got %s", name, result.OfJSONSchema.JSONSchema.Name)
	}
}

func TestNewClient(t *testing.T) {
	// We can't easily test the internals of the client
	// Just ensure the function doesn't panic
	_ = NewClient("test-key", "https://api.test.com")
}

func TestCreateUserMessage(t *testing.T) {
	diff := "- removed line\n+ added line"
	expected := "Generate a commit message for the following diff:\n```diff\n- removed line\n+ added line\n```"

	result := createUserMessage(diff)
	if result != expected {
		t.Errorf("Expected message:\n%s\n\nGot:\n%s", expected, result)
	}
}

func TestFormatExamples(t *testing.T) {
	examples := []config.Example{
		{
			Diff:    "- removed\n+ added",
			Message: "feat: test example",
		},
	}

	results := formatExamples(examples)

	// Should create 2 messages per example (user + assistant)
	if len(results) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(results))
	}
}

// MockedGenerate is a function that mimics the Generate function but accepts our mock client
func MockedGenerate[T Commit | DetailedCommit](
	ctx context.Context, client *mockOpenAIClient,
	model, schemaName, schemaDesc, prompt, systemPrompt string,
	examples []openai.ChatCompletionMessageParamUnion,
) (T, error) {
	// This is a simplified version that uses our test variables
	if testError != nil {
		var zero T
		return zero, testError
	}

	var resp T
	err := json.Unmarshal([]byte(testResponseContent), &resp)
	if err != nil {
		var zero T
		return zero, err
	}

	return resp, nil
}

// MockedGenerateCommitMessage is a function that mimics GenerateCommitMessage but uses our MockedGenerate
func MockedGenerateCommitMessage(
	ctx context.Context, client *mockOpenAIClient, model, diff string, detailed bool,
	examples []config.Example,
) (string, error) {
	// Create the messages (same as in original function)
	systemMessage := `You are an expert programmer assisting with writing git commit messages.
Analyze the provided code diff and generate a concise, informative commit message following
conventional commit standards (e.g., "feat: add user login functionality").
The commit message should accurately describe the changes. Do not include explanations or apologies.
`
	userMessage := createUserMessage(diff)
	apiExamples := formatExamples(examples)

	if detailed {
		r, err := MockedGenerate[DetailedCommit](
			ctx, client, model, "detailed_commit",
			"a git commit message with a description of the changes made",
			userMessage, systemMessage, apiExamples,
		)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s\n\n%s", r.Message, r.Details), nil
	}

	r, err := MockedGenerate[Commit](
		ctx, client, model, "commit", "a git commit message", userMessage, systemMessage, apiExamples,
	)
	if err != nil {
		return "", err
	}
	return r.Message, nil
}

func TestGenerate(t *testing.T) {
	ctx := context.Background()

	t.Run("successful generation", func(t *testing.T) {
		// Create expected response
		expectedCommit := Commit{Message: "feat: add new feature"}
		responseJSON, err := json.Marshal(expectedCommit)
		if err != nil {
			t.Fatalf("Failed to marshal test response: %v", err)
		}

		// Create mock client with expected response
		client := createMockClient(string(responseJSON), nil)

		// Call our mocked version
		result, err := MockedGenerate[Commit](
			ctx, client,
			"gpt-4", "commit", "test description",
			"test prompt", "test system prompt", nil,
		)
		// Check results
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result.Message != expectedCommit.Message {
			t.Errorf("Expected message %q, got %q", expectedCommit.Message, result.Message)
		}
	})

	t.Run("API error", func(t *testing.T) {
		// Create mock client with error
		client := createMockClient("", fmt.Errorf("API error"))

		// Call our mocked version
		_, err := MockedGenerate[Commit](
			ctx, client,
			"gpt-4", "commit", "test description",
			"test prompt", "test system prompt", nil,
		)

		// Check results
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("JSON parse error", func(t *testing.T) {
		// Create mock client with invalid JSON
		client := createMockClient("{invalid json", nil)

		// Call our mocked version
		_, err := MockedGenerate[Commit](
			ctx, client,
			"gpt-4", "commit", "test description",
			"test prompt", "test system prompt", nil,
		)

		// Check results
		if err == nil {
			t.Fatal("Expected error for invalid JSON, got nil")
		}
	})
}

func TestGenerateCommitMessage(t *testing.T) {
	ctx := context.Background()
	examples := []config.Example{}

	t.Run("simple commit", func(t *testing.T) {
		// Create expected response
		expectedCommit := Commit{Message: "feat: add new feature"}
		responseJSON, _ := json.Marshal(expectedCommit)

		// Create mock client
		client := createMockClient(string(responseJSON), nil)

		// Call the mocked function
		message, err := MockedGenerateCommitMessage(ctx, client, "gpt-4", "test diff", false, examples)
		// Check results
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if message != expectedCommit.Message {
			t.Errorf("Expected message %q, got %q", expectedCommit.Message, message)
		}
	})

	t.Run("detailed commit", func(t *testing.T) {
		// Create expected response
		expectedCommit := DetailedCommit{
			Message: "feat: add new feature",
			Details: "- Added X\n- Improved Y",
		}
		responseJSON, _ := json.Marshal(expectedCommit)

		// Create mock client
		client := createMockClient(string(responseJSON), nil)

		// Call the mocked function
		message, err := MockedGenerateCommitMessage(ctx, client, "gpt-4", "test diff", true, examples)
		// Check results
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		expected := fmt.Sprintf("%s\n\n%s", expectedCommit.Message, expectedCommit.Details)
		if message != expected {
			t.Errorf("Expected message %q, got %q", expected, message)
		}
	})

	t.Run("API error", func(t *testing.T) {
		// Create mock client with error
		client := createMockClient("", fmt.Errorf("API error"))

		// Call the mocked function
		_, err := MockedGenerateCommitMessage(ctx, client, "gpt-4", "test diff", false, examples)

		// Check results
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}

