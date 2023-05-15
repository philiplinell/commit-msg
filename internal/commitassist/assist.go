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

type Style string

const (
	// DescriptiveAndNeutral: This style focuses on stating the changes as
	// plainly and objectively as possible. It's typically preferred in most
	// development environments.
	DescriptiveAndNeutral Style = "DescriptiveAndNeutral"

	// ConversationalAndCasual: This style includes using casual language or
	// even humor to describe changes. It's less common and more appropriate
	// for less formal environments or small, close-knit teams.
	ConversationalAndCasual Style = "ConversationalAndCasual"

	// ListBased: Changes are presented in a list format, often
	// used when there are multiple distinct changes that are easier to
	// understand when broken down.
	ListBased Style = "ListBased"

	// ProblemSolution: This style first states the problem that was present
	// and then details the solution that was implemented. It's especially
	// useful when the commit addresses specific bugs or issues.
	ProblemSolution Style = "ProblemSolution"
)

// ValidateMessageStyle returns an error if the assumedStyle is not a valid.
func ValidateMessageStyle(assumedStyle string) (Style, error) {
	switch assumedStyle {
	case string(DescriptiveAndNeutral), string(ConversationalAndCasual), string(ListBased), string(ProblemSolution):
		return Style(assumedStyle), nil
	default:
		return "", fmt.Errorf("invalid style %q", assumedStyle)
	}
}

type MessageConfig struct {
	Style                       Style
	ConventionalCommitCompliant bool
}

// GetCommitMessage returns a commit message based on the git diff provided.
//nolint:funlen
func (o *Client) GetCommitMessage(ctx context.Context, gitDiff string, cfg *MessageConfig) (GetTypeResponse, error) {
	// styleDescriptions is a map of the style to a description of the style,
	// to be used in the prompt to the OpenAI API. It should be used after "The
	// style of the commit messages should be ".
	styleDescriptions := map[Style]string{
		DescriptiveAndNeutral:   "descriptive and neutral. Use clear, concise language to describe the changes. The message should be objective and factual, focusing solely on what was done, without injecting personal opinions or unnecessary context",
		ConversationalAndCasual: "conversational and casual using informal language or even a touch of humor to describe the changes. You should aim to make the commit messages engaging, yet still professional and informative",
		ListBased:               "list-based. Use bullet points or numbered lists to itemize the changes made. Each point should be concise, specific, and self-explanatory. This style is particularly suitable for commits that involve multiple changes or updates. Despite the structured format, ensure the message provides enough context to understand the changes without having to look at the code",
		ProblemSolution:         "problem-solution oriented. Begin by clearly outlining the problem or issue that was addressed. Follow this with a concise explanation of the solution implemented to fix the problem. This style encourages a logical and methodical approach to describing changes, and is particularly effective for commits aimed at fixing bugs or improving functionality",
	}

	if cfg == nil {
		cfg = &MessageConfig{
			Style: DescriptiveAndNeutral,
		}
	}

	if _, err := ValidateMessageStyle(string(cfg.Style)); err != nil {
		return GetTypeResponse{}, err
	}

	conventionalCommitContent := ""
	if cfg.ConventionalCommitCompliant {
		conventionalCommitContent = "Use the conventional commit standard, including any breaking changes, which should be denoted with a '!' (e.g., 'feat!')."
	}

	return o.doChatCompletionRequest(ctx, []openai.Message{
		{
			Role: openai.SystemRole,
			Content: fmt.Sprintf(`You are an insightful assistant that crafts
commit messages. The commit messages should accurately and succinctly explain
the changes made in the files, detailing the reason for changes and the effect
they will have on the project. Your responses should consist of the commit
subject and the commit body, separated by newlines.

The commit subject should:
- Be brief (50 characters or less)
- Use the imperative mood (e.g., "Add", "Fix", "Change")

The commit body should:
- Further explain the changes in detail if necessary
- Be wrapped at 72 characters
- Be separated from the commit subject by a blank line

Make sure they provide enough context to understand the changes without having to look at the code.

The style of the commit message should be %s.
%s
`, styleDescriptions[cfg.Style], conventionalCommitContent),
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
			Content: getExpectedMessage(cfg.Style, cfg.ConventionalCommitCompliant),
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

func getExpectedMessage(style Style, conventionalCommitCompliant bool) string {
	var expectedMessage string
	switch style {
	case DescriptiveAndNeutral:
		expectedMessage = "Add README.md to explain the tool usage\n\n" +
			"This commit adds a new README.md file that serves as a comprehensive guide for utilizing the recently developed tool. The README.md file contains explicit instructions and essential information regarding the functionality of the tool, as well as the details of its interaction with the OpenAI API. It provides insights into the tool's capabilities, along with specific details on the files and lines that are affected during its operation"

	case ConversationalAndCasual:
		expectedMessage = "Unleashing a brand new README.md to demystify our OpenAI-powered commit message wizardry!\n\n" +
			"Hey folks,\n" +
			"We just slapped a shiny new README.md into the mix! 🎉\n" +
			"This bad boy's job is to school you all about our super cool, freshly baked tool that spits out commit message suggestions - all powered by the magic of OpenAI (no wizards were harmed in the process, promise! 🧙.\n" +
			"It's got everything - the ins, the outs, the what-have-yous about our tool. Oh, and it's also gonna give you the lowdown on the stuff we're sending over to OpenAI (don't worry, it's just filenames and changed lines, not your secret cookie recipes! 🍪).\n" +
			"So strap in, take a gander at the README, and let's get those commit messages singing! 🎵"

	case ListBased:
		expectedMessage = "Introducing README.md to illuminate tool usage\n\n " +
			"In this commit:\n\n" +
			"- A new README.md file has been added\n" +
			"- Its purpose: to offer detailed instructions and critical notes about our fresh tool that generates commit message suggestions\n" +
			"- What's covered in the README:\n" +
			"  - The tool's functionality\n" +
			"  - The type of data sent to OpenAI, like filenames and lines changed\n"

	case ProblemSolution:
		expectedMessage = "Addressing the lack of clarity with new README.md\n\n" +
			"Problem: Users were left in the dark about how to use our new commit message suggestion tool, and there was ambiguity regarding what data was being sent to OpenAI.\n\n" +
			"Solution: In this commit, we've introduced a README.md file that does the following:\n\n" +
			"- Provides detailed instructions and important notes about the usage of the tool\n" +
			"- Sheds light on the tool's functionality\n" +
			"- Outlines the specific data it sends to OpenAI, such as filenames and lines changed"
	}

	if conventionalCommitCompliant {
		expectedMessage = "feat: " + expectedMessage
	}

	return expectedMessage
}
