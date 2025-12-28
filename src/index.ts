import { loadConfig } from "./config";
import { DiscordClient } from "./discord/client";
import { GeminiClient } from "./gemini/client";
import { GitHubClient } from "./github/client";

async function run(): Promise<void> {
  const config = loadConfig();

  const githubClient = new GitHubClient({
    username: config.ghUser,
    token: config.ghToken,
  });

  const geminiClient = new GeminiClient({
    apiKey: config.geminiApiKey,
    model: config.geminiModel,
  });

  const discordClient = new DiscordClient({
    webhookUrl: config.discordWebhookUrl,
  });

  console.log("Fetching GitHub events...");
  const events = await githubClient.getDailyEvents();

  if (events.length === 0) {
    console.log("No events found in the last 24 hours");
    console.log("Sending no activity report to Discord...");
    await discordClient.sendNoActivityReport();
    console.log("No activity report sent successfully!");
    return;
  }

  console.log(`Found ${events.length} events`);

  console.log("Generating summary with Gemini...");
  const eventsJson = JSON.stringify(events);
  const summary = await geminiClient.generateDailySummary(eventsJson);

  console.log("Sending report to Discord...");
  await discordClient.sendDailyReport(summary);

  console.log("Daily report sent successfully!");
}

async function main(): Promise<void> {
  try {
    await run();
  } catch (error) {
    console.error("Error:", error);

    try {
      const webhookUrl = Bun.env.DISCORD_WEBHOOK_URL;
      if (webhookUrl) {
        const discordClient = new DiscordClient({ webhookUrl });
        const errorMessage =
          error instanceof Error ? error.message : String(error);
        await discordClient.sendError(errorMessage);
      }
    } catch (discordError) {
      console.error("Failed to send error to Discord:", discordError);
    }

    process.exit(1);
  }
}

main();
