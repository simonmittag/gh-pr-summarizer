# gh-pr-summarizer

CLI utility to generate pull request summaries from commit history and issue trackers.

> [!WARNING]
> Active development. Expect breaking changes.

## Features

- Commit Summarization: Collects subjects between current and base branch.
- Issue Tracker Integration: Fetches ticket details from Linear, GitHub, or Jira.
- AI-Powered: Optional OpenAI integration for polished descriptions.
- Configurable: Support for global and project-specific settings.

## Installation

```bash
go install github.com/simonmittag/gh-pr-summarizer/cmd/gh-pr-summarizer@latest
```

## Usage

Run from any git repository:

```bash
gh-pr-summarizer
```

### Flags

- `--current <branch>`: Override detected current branch.
- `--base <branch>`: Override detected base branch (defaults to main/master).
- `--version`: Show version.

### Environment Variables

- `OPENAI_API_KEY`: Enables AI-generated summaries.
- `GITHUB_TOKEN`, `LINEAR_API_KEY`, `ATLASSIAN_TOKEN`: Defaults for Tracker authentication (these are configurable per tracker if you have a complex workstation setup).


## Configuration

Hierarchical system: local `.ghpr.toml` overrides global `~/.config/gh-pr-summarizer/config.toml`.

### Example config.toml

```toml
tracker = "linear"

[linear]
ticket_url_stem = "https://linear.app/my-org/issue/"
token_env = "MY_CUSTOM_LINEAR_TOKEN_ENV" # Defaults to LINEAR_API_KEY

[github]
token_env = "GITHUB_TOKEN"

[jira]
ticket_url_stem = "https://my-org.atlassian.net/browse/"
token_env = "ATLASSIAN_TOKEN"
```

### Ticket Detection

Detects ticket IDs in branch names (e.g., `feature/PROJ-123-add-login`) to fetch context.

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) and our [Code of Conduct](CODE_OF_CONDUCT.md).

## License

[Apache 2.0](LICENSE)
