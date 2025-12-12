# GitHub Daily Reporter (Go + Gemini)

A Go application that automatically generates daily summaries of your GitHub activity using Google's Gemini AI and sends them to Discord. Perfect for developers who want to track and share their coding progress.

## Features

- **Comprehensive Activity Tracking**: Monitors both public and private repository events
- **AI-Powered Summaries**: Uses Google Gemini 2.0 Flash to generate concise, engaging summaries
- **Discord Integration**: Automatically posts daily reports to Discord channels via webhooks
- **GitHub Actions Automation**: Runs daily at midnight (Brasília time) using GitHub Actions
- **Clean Architecture**: Well-organized code structure following Go best practices
- **Docker Support**: Containerized application for easy deployment
- **Event Filtering**: Captures activities from the last 24 hours including:
  - Push events with commit details
  - Repository creation and deletion
  - Issue management
  - Pull request activities
  - Branch operations

## Prerequisites

- Go 1.23 or higher
- Docker and Docker Compose (optional, for containerized deployment)
- GitHub Personal Access Token
- Google Gemini API Key
- Discord Webhook URL

## Installation

### Local Setup

1. **Clone the repository:**
```bash
git clone https://github.com/carvalhocaio/routines-in-the-night
cd routines-in-the-night
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Configure environment variables:**
```bash
cp .env.example .env
```

Edit `.env` with your credentials:
```env
GH_USER=your_github_username
GH_TOKEN=your_github_personal_access_token
GEMINI_API_KEY=your_gemini_api_key
DISCORD_WEBHOOK_URL=your_discord_webhook_url
```

### Docker Setup

1. **Configure environment variables** (same as above)

2. **Build and run with Docker Compose:**
```bash
docker-compose up --build
```

## Configuration

### GitHub Personal Access Token

Create a GitHub Personal Access Token with the following scopes:
- `repo` (for private repository access)
- `user` (for user events access)

[Create token here](https://github.com/settings/tokens)

### Gemini API Key

1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Create a new API key
3. Copy the key to your `.env` file

### Discord Webhook

1. Go to your Discord server settings
2. Navigate to Integrations > Webhooks
3. Create a new webhook and copy the URL

## Usage

### Local Execution

Using Make:
```bash
make run
```

Or directly:
```bash
go run ./cmd/reporter/main.go
```

### Docker Execution

```bash
docker-compose up
```

### Automated Execution

The project includes a GitHub Actions workflow that runs automatically every day at midnight (Brasília time).

To enable automation:

1. Add the required secrets to your GitHub repository:
   - `GH_USER`
   - `GH_TOKEN`
   - `GEMINI_API_KEY`
   - `DISCORD_WEBHOOK_URL`

2. The workflow will run automatically or can be triggered manually from the Actions tab.

## Project Structure

```
.
├── cmd/
│   └── reporter/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── github/
│   │   └── client.go            # GitHub API client
│   ├── gemini/
│   │   └── client.go            # Gemini AI client
│   └── discord/
│       └── client.go            # Discord webhook client
├── .github/
│   └── workflows/
│       └── daily-github-report.yml  # GitHub Actions workflow
├── .env.example                 # Environment variables template
├── .gitignore                   # Git ignore rules
├── docker-compose.yml           # Docker Compose configuration
├── Dockerfile                   # Docker image definition
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── Makefile                     # Build automation
└── README.md                    # This file
```

## 🛠️ Development

### Available Make Commands

```bash
make help          # Show available commands
make build         # Build the application
make run           # Build and run
make test          # Run tests
make coverage      # Generate coverage report
make clean         # Clean build artifacts
make docker-build  # Build Docker image
make docker-run    # Run with Docker Compose
make format        # Format code
make lint          # Run linter
make deps          # Download dependencies
make tidy          # Tidy go.mod
```

### Code Formatting

```bash
make format
```

### Running Tests

```bash
make test
```

### Linting

Install golangci-lint first:
```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

Then run:
```bash
make lint
```

## Architecture

The application follows clean architecture principles with clear separation of concerns:

- **cmd/reporter**: Application entry point and main workflow
- **internal/config**: Configuration loading and validation
- **internal/github**: GitHub API integration
- **internal/gemini**: Gemini AI integration
- **internal/discord**: Discord webhook integration

Each package is independent and testable, following Go best practices.

## How It Works

1. **Event Retrieval**: The GitHub client fetches events from the last 24 hours using the GitHub API
2. **Event Processing**: Events are filtered and formatted to extract relevant information
3. **AI Summary Generation**: Gemini AI processes the events and generates a detailed narrative summary (100-150 words)
4. **Discord Notification**: The summary is sent to Discord as an embedded message with timestamp

## Event Types Tracked

- **PushEvent**: Code pushes with commit messages
- **CreateEvent**: Repository, branch, or tag creation
- **DeleteEvent**: Repository, branch, or tag deletion
- **IssuesEvent**: Issue creation, closing, or updates
- **PullRequestEvent**: PR creation, merging, or updates

## Error Handling

The application includes comprehensive error handling:
- API rate limit management
- Network connection failures
- Malformed event data processing
- Discord webhook delivery failures
- Gemini API errors with fallback messages

All errors are logged and optionally sent to Discord for monitoring.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following Go best practices
4. Format your code (`make format`)
5. Run tests (`make test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## License

This project is open source and available under the MIT License.

## Acknowledgments

- [Google Gemini AI](https://deepmind.google/technologies/gemini/) for AI-powered summaries
- [GitHub API](https://docs.github.com/en/rest) for event tracking
- [Discord Webhooks](https://discord.com/developers/docs/resources/webhook) for notifications

## Support

If you encounter any issues or have questions:
1. Check the [Issues](../../issues) page
2. Create a new issue with details about your problem
3. Include logs and environment information
