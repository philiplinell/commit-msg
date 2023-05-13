package openai_test

import (
	"context"
	"embed"
	"net/http"
	"testing"

	"github.com/philiplinell/commit-msg/internal/openai"
)

//go:embed testdata
var testdata embed.FS

type mockHTTPClient struct {
	DoFn func(req *http.Request) (*http.Response, error)
}

func (f mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return f.DoFn(req)
}

func TestAIModelCost(t *testing.T) {
	testCases := []struct {
		totalTokens  int
		expectedCost float64
	}{
		{
			totalTokens:  1000,
			expectedCost: 0.002,
		},
		{
			totalTokens:  100,
			expectedCost: 0.0002,
		},
		{
			totalTokens:  1000000,
			expectedCost: 2,
		},
		{
			totalTokens:  0,
			expectedCost: 0.0,
		},
		{
			totalTokens:  -1,
			expectedCost: 0.0,
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable

		t.Run("", func(t *testing.T) {
			got := openai.GPT3_5Turbo.Cost(tc.totalTokens)

			if got != tc.expectedCost {
				t.Errorf("got %v, want %v", got, tc.expectedCost)
			}
		})
	}
}

func TestInvalidTemperatureReturnsErr(t *testing.T) {
	testCases := []struct {
		temperature float32
	}{
		{
			temperature: -5,
		},
		{
			temperature: 1.1,
		},
		{
			temperature: 10,
		},
		{
			temperature: 10000,
		},
	}

	httpClient := createFakeHTTPClient(t, http.StatusOK, "testdata/chat_completion_response.json")

	client := openai.NewClient(httpClient, "")

	for _, tc := range testCases {
		tc := tc // capture range variable

		t.Run("", func(t *testing.T) {
			_, err := client.ChatCompletionRequest(context.Background(), []openai.Message{}, openai.GPT3_5Turbo, tc.temperature)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestSuccessfulChatCompletionRequest(t *testing.T) {
	httpClient := createFakeHTTPClient(t, http.StatusOK, "testdata/chat_completion_response.json")

	client := openai.NewClient(httpClient, "")

	response, err := client.ChatCompletionRequest(context.Background(), []openai.Message{}, openai.GPT3_5Turbo, 0.5)
	if err != nil {
		t.Fatal(err)
	}

	if len(response.Messages) != 1 {
		t.Error("expected message in the response")
	}
}

func createFakeHTTPClient(t *testing.T, expectedStatusCode int, testdataFile string) openai.Doer {
	t.Helper()

	file, err := testdata.Open(testdataFile)
	if err != nil {
		t.Fatal(err)
	}

	defer file.Close()

	return mockHTTPClient{
		DoFn: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: expectedStatusCode,
				Body:       file,
				Header:     make(http.Header),
			}, nil
		},
	}
}
