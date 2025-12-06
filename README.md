# Seer

**Seer** is a self-hosted tool to detect market opportunities for indie developers.

> "The Seer sees what others miss."

## Philosophy

- You control your data
- You choose where it runs
- You decide which AI to use

## Quick Start

### Using Docker

```bash
docker-compose up
```

### Using Binary

```bash
# Download the latest release
./seer
```

Visit `http://localhost:8080` to access the dashboard.

## Configuration

Copy the example configuration:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` to customize your settings.

## Development

### Prerequisites

- Go 1.23+
- Bun (for frontend)

### Commands

```bash
# Run in development mode
make dev

# Run Go server only
make dev-server

# Build CE binary
make build-ce

# Build Pro binary
make build-pro

# Run tests
make test

# Clean build artifacts
make clean
```

## Sources

Seer monitors multiple sources for opportunities:

| Source | Description | CE | Pro |
|--------|-------------|-----|-----|
| Hacker News | Algolia API | Yes | Yes |
| GitHub | Trending/Issues | Yes | Yes |
| npm | New packages | Yes | Yes |
| DEV.to | Articles | Yes | Yes |
| RSS | Custom feeds | Max 2 | Unlimited |
| Reddit | Subreddits | No | Yes |
| Twitter/X | Keywords | No | Yes |
| Custom API | JSON endpoint | No | Yes |

## AI Providers (Pro)

| Provider | Type | Notes |
|----------|------|-------|
| Ollama | Local | Free, private |
| OpenAI | Cloud | GPT-4, GPT-3.5 |
| Anthropic | Cloud | Claude |
| DeepSeek | Cloud | Cheap, good for code |
| Groq | Cloud | Fast, rate limited |
| Google AI | Cloud | Gemini |
| Mistral | Cloud | European, privacy |
| OpenRouter | Cloud | Multi-provider gateway |

## License

See [LICENSE](LICENSE) file.

## Support

- Email: seer@mendex.io
- Issues: [GitHub Issues](https://github.com/mx-seer/seer-ce/issues)
