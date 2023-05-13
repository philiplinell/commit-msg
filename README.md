# Commit Message

Create a commit message suggestion from the git diff using the openAI API.

Note that this means that filename and lines changed is sent to openAI. If that
bothers you - don't use this tool.

## Todos

- [ ] Support a tone/style setting.
    Use the following:

    - Descriptive and Neutral: This style focuses on stating the changes as plainly and objectively as possible. It's typically preferred in most development environments.

    - Conversational and Casual: This style includes using casual language or even humor to describe changes. It's less common and more appropriate for less formal environments or small, close-knit teams.

    - Bullet-pointed or List-based: Changes are presented in a list format, often used when there are multiple distinct changes that are easier to understand when broken down.

    - Problem-Solution: This style first states the problem that was present and then details the solution that was implemented. It's especially useful when the commit addresses specific bugs or issues.

- [ ] Flag for conventional commit, on or off.
