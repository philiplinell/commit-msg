package commitassist

import (
	"context"
	"fmt"
	"strings"

	"github.com/philiplinell/commit-msg/internal/openai"
)

type UnexpectedStateError struct {
	Msg string
}

func (e UnexpectedStateError) Error() string {
	return e.Msg
}

type UnsureError struct {
	Msg string
}

func (e UnsureError) Error() string {
	return e.Msg
}

type Client struct {
	client *openai.Client
}

func New(client *openai.Client) *Client {
	return &Client{
		client: client,
	}
}

type GetTypeResponse struct {
	Message string

	// Cost is the cost of the request in cent.
	Cost float64
}

func (o *Client) GetCommitMessage(ctx context.Context, gitDiff string) (GetTypeResponse, error) {
	return o.doChatCompletionRequest(ctx, []openai.Message{
		{
			Role:    openai.SystemRole,
			Content: "You are helpful assistant that suggest commit messages. The commit messages should explain the changes made in the files, including any breaking changes, which should be denoted with a '!' (e.g., 'feat!'). The structure of the commit message can be flexible, varying based on the size and complexity of the changes. You should only respond with the commit subject and the commit body separated by newlines. The commit subject should be in imperative mood.",
		},
		{
			Role: openai.UserRole,
			Content: `diff --git a/README.md b/README.md
new file mode 100644
index 0000000..ca34b6a
--- /dev/null
+++ b/README.md
@@ -0,0 +1,21 @@
+# Commit Message
+
+Create a commit message suggestion from the git diff using the openAI API.
+
+Note that this means that filename and lines changed is sent to openAI. If that
+bothers you - don't use this tool.`,
		},
		{
			Role:    openai.AssistantRole,
			Content: "feat: Add README.md to explain the tool usage\n\nThis commit introduces a new README.md file. The purpose of this file is to provide detailed instructions and important notes about the new tool that generates commit message suggestions using the OpenAI API. It highlights the tool's functionality and data it sends to OpenAI, including filenames and lines changed.",
		},

		// This is the final message that the assistant should respond to.
		{
			Role:    openai.UserRole,
			Content: gitDiff,
		},
	})
}

func (o *Client) doChatCompletionRequest(ctx context.Context, messages []openai.Message) (GetTypeResponse, error) {
	content, err := o.client.ChatCompletionRequest(ctx, messages, openai.GPT3_5Turbo, 0.2)
	if err != nil {
		return GetTypeResponse{}, fmt.Errorf("could not do ChatCompletionRequest: %w", err)
	}

	if len(content.Messages) != 1 {
		return GetTypeResponse{}, UnexpectedStateError{fmt.Sprintf("unexpected number of messages returned, got %d", len(content.Messages))}
	}

	message := content.Messages[0]

	if strings.Contains(message, "unsure") {
		return GetTypeResponse{}, UnsureError{message}
	}

	return GetTypeResponse{
		Message: message,
		Cost:    content.Cost * 100,
	}, nil
}
