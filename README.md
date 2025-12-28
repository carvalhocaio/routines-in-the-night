# GitHub Daily Reporter (Bun + Gemini)

A TypeScript application that automatically generates daily summaries of your GitHub activity using Google's Gemini AI and sends them to Discord. Perfect for developers who want to track and share their coding progress.

## Features

- **Comprehensive Activity Tracking**: Monitors both public and private repository events
- **AI-Powered Summaries**: Uses Google Gemini to generate concise, engaging summaries
- **Discord Integration**: Automatically posts daily reports to Discord channels via webhooks
- **GitHub Actions Automation**: Runs daily at midnight (Brasília time) using GitHub Actions
- **Robust Error Handling**: Request timeouts, retry logic with exponential backoff
- **Event Filtering**: Captures activities from the last 24 hours including:
  - Push events with commit details
  - Repository creation and deletion
  - Issue management
  - Pull request activities
  - Branch operations

## Prerequisites

- [Bun](https://bun.sh) v1.0 or higher
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
bun install
```

3. **Configure environment variables:**
```bash
cp .env.sample .env
```

Edit `.env` with your credentials:
```env
GH_USER=your_github_username
GH_TOKEN=your_github_personal_access_token
GEMINI_API_KEY=your_gemini_api_key
GEMINI_MODEL=gemini-2.5-flash  # Optional, defaults to gemini-2.5-flash
DISCORD_WEBHOOK_URL=your_discord_webhook_url
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

```bash
bun run start
```

Or with watch mode for development:
```bash
bun run dev
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
├── src/
│   ├── index.ts              # Application entry point
│   ├── config/
│   │   └── index.ts          # Configuration management
│   ├── github/
│   │   ├── client.ts         # GitHub API client
│   │   └── types.ts          # GitHub event types
│   ├── gemini/
│   │   ├── client.ts         # Gemini AI client
│   │   └── prompt.ts         # AI prompt template
│   ├── discord/
│   │   ├── client.ts         # Discord webhook client
│   │   └── types.ts          # Discord embed types
│   └── utils/
│       └── fetch.ts          # Fetch with retry and timeout
├── .github/
│   └── workflows/
│       ├── ci.yml                   # CI workflow (test, lint, typecheck)
│       └── daily-github-report.yml  # Daily report workflow
├── .env.sample               # Environment variables template
├── biome.json                # Linter/formatter configuration
├── package.json              # Project dependencies
├── tsconfig.json             # TypeScript configuration
└── README.md                 # This file
```

## Development

### Available Commands

```bash
bun run start      # Run the application
bun run dev        # Run with watch mode
bun test           # Run tests
bun run lint       # Run Biome linter
bun run format     # Format code with Biome
bun run check      # Run all Biome checks
bun run tsc        # Type check
```

### Running Tests

```bash
bun test
```

### Linting & Formatting

```bash
bun run lint      # Check for issues
bun run format    # Fix formatting
```

## Architecture

The application follows clean architecture principles with clear separation of concerns:

- **src/config**: Configuration loading and validation
- **src/github**: GitHub API integration with retry logic
- **src/gemini**: Gemini AI integration for summary generation
- **src/discord**: Discord webhook integration
- **src/utils**: Shared utilities (fetch with retry/timeout)

Each module is independent and testable.

## How It Works

1. **Event Retrieval**: The GitHub client fetches events from the last 24 hours using the GitHub API
2. **Event Processing**: Events are filtered and formatted to extract relevant information
3. **AI Summary Generation**: Gemini AI processes the events and generates a detailed technical summary in Portuguese
4. **Discord Notification**: The summary is sent to Discord as an embedded message with timestamp

If no events are found, a "no activity" message is sent to Discord instead.

## Event Types Tracked

- **PushEvent**: Code pushes with commit messages
- **CreateEvent**: Repository, branch, or tag creation
- **DeleteEvent**: Repository, branch, or tag deletion
- **IssuesEvent**: Issue creation, closing, or updates
- **PullRequestEvent**: PR creation, merging, or updates

## Error Handling

The application includes comprehensive error handling:
- Request timeouts (30s for GitHub/Gemini, 15s for Discord)
- Retry logic with exponential backoff for transient errors (429, 5xx)
- Detailed error messages with context
- Error notifications sent to Discord for monitoring

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Format your code (`bun run format`)
5. Run tests (`bun test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## License

This project is open source and available under the MIT License.

## Acknowledgments

- [Google Gemini AI](https://deepmind.google/technologies/gemini/) for AI-powered summaries
- [GitHub API](https://docs.github.com/en/rest) for event tracking
- [Discord Webhooks](https://discord.com/developers/docs/resources/webhook) for notifications
- [Bun](https://bun.sh) for the fast JavaScript runtime

## Support

If you encounter any issues or have questions:
1. Check the [Issues](../../issues) page
2. Create a new issue with details about your problem
3. Include logs and environment information
