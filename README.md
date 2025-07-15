# GitHub Daily Reporter

A Python application that automatically generates daily summaries of your GitHub activity using
OpenAI's GPT and sends them to Discord. Perfect for developers who want to track and share their
coding progress.

## Features

- **Comprehensive Activity Tracking**: Monitors both public and private repository events
- **AI-Powered Summaries**: Uses OpenAI GPT-4o-mini to generate concise, engaging summaries
- **Discord Integration**: Automatically posts daily reports to Discord channels via webhooks
- **GitHub Actions Automation**: Runs daily at midnight (Brasília time) using GitHub Actions
- **Event Filtering**: Captures activities from the last 24 hours including:
  - Push events with commit details
  - Repository creation and deletion
  - Issue management
  - Pull request activities
  - Branch operations

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd routines-in-the-night
```

2. Install dependencies:
```bash
pip install -r requirements.txt
```

3. Set up environment variables by creating a `.env` file:
```env
GH_USER=your_github_username
GH_TOKEN=your_github_personal_access_token
OPENAI_API_KEY=your_openai_api_key
DISCORD_WEBHOOK_URL=your_discord_webhook_url
```

## Configuration

### GitHub Personal Access Token
Create a GitHub Personal Access Token with the following scopes:
- `repo` (for private repository access)
- `user` (for user events access)

### Discord Webhook
1. Go to your Discord server settings
2. Navigate to Integrations > Webhooks
3. Create a new webhook and copy the URL

### OpenAI API Key
Obtain your API key from the OpenAI platform dashboard.

## Usage

### Local Execution
Run the script manually:
```bash
python github_daily.py
```

### Automated Execution
The project includes a GitHub Actions workflow that runs automatically every day at midnight 
(Brasília time). The workflow is defined in [`.github/workflows/daily-github-report.yml`](.github/workflows/daily-github-report.yml).

To enable automation:
1. Add the required secrets to your GitHub repository:
   - `GH_USER`
   - `GH_TOKEN`
   - `OPENAI_API_KEY`
   - `DISCORD_WEBHOOK_URL`

2. The workflow will run automatically or can be triggered manually from the Actions tab.

## Project Structure

```
.
├── github_daily.py          # Main application script
├── requirements.txt         # Python dependencies
├── ruff.toml                # Code formatting configuration
├── .env                     # Environment variables (not tracked)
├── .gitignore               # Git ignore rules
├── __init__.py              # Python package initialization
├── README.md                # Project documentation
└── .github/
    └── workflows/
        └── daily-github-report.yml  # GitHub Actions workflow
```

## Code Quality

The project uses Ruff for code formatting and linting with the following configuration:
- Line length: 79 characters
- Indent width: 4 spaces
- Selected rules: Import sorting, Error detection, PEP 8 compliance, Pylint recommendations, and pytest best practices

Run code formatting:
```bash
ruff format
```

Run linting:
```bash
ruff check
```

## How It Works

1. **Event Retrieval**: The [`GitHubDailyReporter`](github_daily.py) class fetches GitHub events from the last 24 hours using the GitHub API
2. **Event Processing**: Events are filtered and formatted to extract relevant information like commit messages, repository names, and activity types
3. **AI Summary Generation**: OpenAI GPT-4o-mini processes the events and generates a concise, Twitter-ready summary (max 280 characters)
4. **Discord Notification**: The summary is sent to Discord as an embedded message with timestamp and formatting

## Dependencies

The project relies on several key libraries:
- `requests`: HTTP requests for GitHub and Discord APIs
- `openai`: OpenAI API integration
- `python-dotenv`: Environment variable management
- `ruff`: Code formatting and linting

See [`requirements.txt`](requirements.txt) for the complete list of dependencies with specific versions.

## Error Handling

The application includes comprehensive error handling:
- API rate limit management
- Network connection failures
- Malformed event data processing
- Discord webhook delivery failures
- OpenAI API errors with fallback messages

Errors are logged to console and optionally sent to Discord for monitoring.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes following the existing code style
4. Run linting and formatting checks
5. Submit a pull request
