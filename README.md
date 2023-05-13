# Commit Message

Create a commit message suggestion from the git diff using the openAI API.

Note that this means that filename and lines changed is sent to openAI. If that
bothers you - don't use this tool.

Make sure environment variable `OPENAI_API_KEY` contains a valid API key. Note
that this tool uses the openAI API so it will incur a cost. It is recommended to
set a hard limit in the [openai account settings panel](https://platform.openai.com/account/billing/limits).

## Example Usage

Needs `commit-msg` (that is the binary from this repo) in PATH.

```sh
───────┬────────────────────────────────────────────────────────────────────────────────────
       │ File: .git/hooks/prepare-commit-msg
───────┼────────────────────────────────────────────────────────────────────────────────────
   1   │ #!/bin/sh
   2   │
   3   │ # Use CLI tool commit-msg to fetch a suggested commit message. Prepend the
   4   │ # suggested commit message to the commit message file.
   5   │
   6   │ COMMIT_MSG_FILE=$1
   7   │
   8   │ echo "Fetching suggested commit message..."
   9   │
  10   │ COMMIT_MSG=$(commit-msg --timeout=15s --file=$COMMIT_MSG_FILE)
  11   │
  12   │ if [ $? -ne 0 ]; then
  13   │     echo "❌ prepare-commit-msg: commit-msg failed. Doing nothing..."
  14   │     exit 0
  15   │ fi
  16   │
  17   │ printf '%s\n%s\n' "${COMMIT_MSG}" "$(cat $COMMIT_MSG_FILE)" >$COMMIT_MSG_FILE
───────┴────────────────────────────────────────────────────────────────────────────────────

```

## Todos

- [ ] Support a tone/style setting.
    Use the following:

    - Descriptive and Neutral: This style focuses on stating the changes as plainly and objectively as possible. It's typically preferred in most development environments.

    - Conversational and Casual: This style includes using casual language or even humor to describe changes. It's less common and more appropriate for less formal environments or small, close-knit teams.

    - Bullet-pointed or List-based: Changes are presented in a list format, often used when there are multiple distinct changes that are easier to understand when broken down.

    - Problem-Solution: This style first states the problem that was present and then details the solution that was implemented. It's especially useful when the commit addresses specific bugs or issues.

- [ ] Flag for conventional commit, on or off.
