package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// https://platform.openai.com/docs/guides/chat/introduction

const (
	chatCompletionURL = "https://api.openai.com/v1/chat/completions"
)

// Client is the OpenAI API client.
type Client struct {
	httpClient Doer
	apiKey     string
}

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewClient creates a new OpenAI API client.
func NewClient(httpClient Doer, apiKey string) *Client {
	return &Client{
		httpClient: httpClient,
		apiKey:     apiKey,
	}
}

type aiModel string

const (
	// GPT3_5Turbo - The most capable GPT-3.5 model and optimized for chat at
	// 1/10th the cost of text-davinci-003. Will be updated with the latest
	// model iteration.
	// gpt-3.5-turbo is recomennded over the other GPT-3.5 model due to its
	// (lowest) cost.
	GPT3_5Turbo aiModel = "gpt-3.5-turbo"
)

// Coster is an interface that models can implement to calculate the cost of
// request based on the total tokens used.
type Coster interface {
	// Cost returns the cost in dollars of the request based on the total
	// tokens used.
	Cost(totalTokens int) float64
}

func (m aiModel) Cost(totalTokens int) float64 {
	if totalTokens <= 0 {
		return 0.0
	}

	switch m {
	case GPT3_5Turbo:
		// $0.002 / 1K tokens
		return float64(totalTokens) * 0.002 / 1000
	default:
		return 0.0
	}
}

// aiRole defines the role of the message. Typically a conversation is
// formatted with a system message first, followed by alternating user and
// assistant messages.
//
// Example:
//  "role": "system",	 "content": "You are a helpful assistant."
//  "role": "user", 	 "content": "Who won the world series in 2020?"
//  "role": "assistant", "content": "The Los Angeles Dodgers won the World Series in 2020."
//  "role": "user", 	 "content": "Where was it played?"
type aiRole string

const (
	// The user messages help instruct the assistant. They can be generated by
	// the end users of an application, or set by a developer as an
	// instruction.
	UserRole aiRole = "user"

	// The system message helps set the behavior of the assistant. E.g. the
	// assistant can be instructed with "You are a helpful assistant."
	SystemRole aiRole = "system"

	// The assistant messages help store prior responses. They can also be
	// written by a developer to help give examples of desired behavior.
	AssistantRole aiRole = "assistant"
)

// Message is used to interact with the OpenAI model.
type Message struct {
	// Role is the role of the message.
	Role aiRole `json:"role"`
	// Content is the message to send to the OpenAI model.
	Content string `json:"content"`
}

// ChatCompletionRequest does a request to the openai chat completion API.
//
// temperature decides how deterministic the model is in generating a
// response. It must be a value between 0 and 1 (inclusive). A lower
// temperature means that completions will be more accurate and deterministic.
// A higher temperature value means that the completions will be more diverse.
// See more about temperature here:
// https://platform.openai.com/docs/quickstart/adjust-your-settings
func (c *Client) ChatCompletionRequest(ctx context.Context, messages []Message, model aiModel, temperature float32) (chatCompletionResponse, error) {
	if temperature < 0 || temperature > 1 {
		return chatCompletionResponse{}, fmt.Errorf("temperature must be between 0 and 1 (inclusive), got %f", temperature)
	}

	requestBody := chatCompletionRequest{
		Model:    string(model),
		Messages: messages,
	}

	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		return chatCompletionResponse{}, fmt.Errorf("could not marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatCompletionURL, bytes.NewBuffer(requestBytes))
	if err != nil {
		return chatCompletionResponse{}, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return chatCompletionResponse{}, fmt.Errorf("could not do request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return chatCompletionResponse{}, fmt.Errorf("got status code %q, expected %d", resp.Status, http.StatusOK)
	}

	var cResponse rawChatCompletionResponse

	err = json.NewDecoder(resp.Body).Decode(&cResponse)
	if err != nil {
		return chatCompletionResponse{}, fmt.Errorf("could not decode response: %w", err)
	}

	cost := calculateCost(cResponse.Usage.TotalTokens, model)
	answers := []string{}

	for _, choice := range cResponse.Choices {
		answers = append(answers, choice.Content())
	}

	return chatCompletionResponse{
		Created:  time.Unix(cResponse.Created, 0),
		Model:    model,
		Cost:     cost,
		Messages: answers,
	}, nil
}

type chatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature"`
}

type rawChatCompletionUsageResponse struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type rawChatCompletionMessageResponse struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type rawChatCompletionChoiceResponse struct {
	Message      rawChatCompletionMessageResponse `json:"message"`
	FinishReason string                           `json:"finish_reason"`
	Index        int                              `json:"index"`
}

type rawChatCompletionResponse struct {
	ID      string                            `json:"id"`
	Object  string                            `json:"object"`
	Created int64                             `json:"created"`
	Model   string                            `json:"model"`
	Usage   rawChatCompletionUsageResponse    `json:"usage"`
	Choices []rawChatCompletionChoiceResponse `json:"choices"`
}

type chatCompletionResponse struct {
	Created time.Time
	Model   aiModel

	// Cost is the cost for the request in dollars.
	Cost float64

	Messages []string
}

func (c rawChatCompletionChoiceResponse) Content() string {
	return c.Message.Content
}

func calculateCost(totalTokens int, model Coster) float64 {
	return model.Cost(totalTokens)
}