package main

import (
	"fmt"
	"log"
	"os"

	"github.com/carvalhocaio/routines-in-the-night/internal/config"
	"github.com/carvalhocaio/routines-in-the-night/internal/discord"
	"github.com/carvalhocaio/routines-in-the-night/internal/gemini"
	"github.com/carvalhocaio/routines-in-the-night/internal/github"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize services
	githubClient := github.NewClient(cfg.GitHubUser, cfg.GitHubToken)
	geminiClient := gemini.NewClient(cfg.GeminiAPIKey, cfg.GeminiModel)
	discordClient := discord.NewClient(cfg.DiscordWebhookURL)

	// Execute the daily report workflow
	if err := run(githubClient, geminiClient, discordClient); err != nil {
		log.Printf("Error running daily report: %v", err)

		// Try to send error to Discord
		if sendErr := discordClient.SendError(err); sendErr != nil {
			log.Printf("Failed to send error to Discord: %v", sendErr)
		}

		os.Exit(1)
	}

	log.Println("Daily report completed successfully!")
}

func run(
	githubClient *github.Client,
	geminiClient *gemini.Client,
	discordClient *discord.Client,
) error {
	// Fetch GitHub events from last 24 hours
	log.Println("Fetching GitHub events...")
	events, err := githubClient.GetDailyEvents()
	if err != nil {
		return fmt.Errorf("failed to fetch GitHub events: %w", err)
	}

	log.Printf("Found %d events in the last 24 hours", len(events))

	if len(events) == 0 {
		log.Println("No events found, sending default message")
		return discordClient.SendDailyReport(
			"Hoje foi um dia de planejamento e reflexão no código.",
		)
	}

	// Generate summary using Gemini
	log.Println("Generating summary with Gemini AI...")
	summary, err := geminiClient.GenerateDailySummary(events)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	log.Printf("Generated summary (%d characters)", len(summary))

	// Send to Discord
	log.Println("Sending report to Discord...")
	if err := discordClient.SendDailyReport(summary); err != nil {
		return fmt.Errorf("failed to send to Discord: %w", err)
	}

	log.Println("Report sent successfully!")
	return nil
}
