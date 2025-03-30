# diffgpt

[![Run Go Tests](https://github.com/Kabilan108/diffgpt/actions/workflows/test.yml/badge.svg)](https://github.com/Kabilan108/diffgpt/actions/workflows/test.yml)

AI-powered commit message generator that learns from your repository's commit style.

## Features

- Generates conventional commit messages from staged changes
- Learns from your repository's existing commit history
- Supports both global and per-repository commit style examples
- Compatible with any OpenAI-compatible API provider

## Installation

### Using Pre-built Binary

```bash
# Download the latest release from GitHub
curl -Lo diffgpt "https://github.com/kabilan108/diffgpt/releases/latest/download/diffgpt"
chmod +x diffgpt
sudo mv diffgpt /usr/local/bin/
```

### Building from Source

```bash
go install github.com/kabilan108/diffgpt@latest
```

## Configuration

Set the following environment variables or use command-line flags:

```bash
DIFFGPT_API_KEY=<your-api-key>        # Required: API key for LLM provider
DIFFGPT_BASE_URL=<api-base-url>       # Optional: Base URL for API (default: OpenAI)
DIFFGPT_MODEL=<model-name>            # Optional: Model to use (default: gpt-4o-mini)
```

## Usage

### Basic Usage

Stage your changes and run:

```bash
diffgpt
```

### Learning from Repository History

```bash
# Learn from current repository
diffgpt learn

# Learn from specific repository
diffgpt learn /path/to/repo

# Learn globally (for use across all repos)
diffgpt learn --global

# Clear learned examples
diffgpt learn --clear
```

### Additional Options

```bash
# Generate detailed commit messages
diffgpt --detailed

# Use specific model
diffgpt --model gpt-4

# Use different API provider
diffgpt --base-url https://api.provider.com/v1
```

### Pipe Mode

```bash
git diff | diffgpt
```

## License

MIT
