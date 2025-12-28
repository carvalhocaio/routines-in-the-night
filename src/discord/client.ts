import { FetchError, fetchWithRetry } from "../utils/fetch";
import type { DiscordWebhookPayload } from "./types";

const COLOR_BLUE = 0x7289da;
const COLOR_RED = 0xff0000;

// Strengthened regex: validates Discord webhook URL format more strictly
// - Must be HTTPS
// - Webhook ID must be 17-20 digits (snowflake format)
// - Token must be alphanumeric with dashes/underscores, 60-80 chars
const DISCORD_WEBHOOK_PATTERN =
  /^https:\/\/(?:(?:canary|ptb)\.)?discord\.com\/api\/webhooks\/\d{17,20}\/[\w-]{60,80}$/;

export class DiscordWebhookError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "DiscordWebhookError";
  }
}

export interface DiscordClientOptions {
  webhookUrl: string;
  /** Skip URL validation (for testing only) */
  skipValidation?: boolean;
}

/**
 * Client for sending messages to Discord via webhooks.
 */
export class DiscordClient {
  private webhookUrl: string;

  /**
   * Creates a new Discord client.
   * @param options - Configuration options including the webhook URL
   * @throws {DiscordWebhookError} If the webhook URL is invalid
   */
  constructor(options: DiscordClientOptions) {
    if (!options.skipValidation && !this.isValidWebhookUrl(options.webhookUrl)) {
      throw new DiscordWebhookError(
        "Invalid Discord webhook URL format. Expected: https://discord.com/api/webhooks/{id}/{token}"
      );
    }
    this.webhookUrl = options.webhookUrl;
  }

  /**
   * Sends a daily activity report to Discord.
   * @param summary - The AI-generated summary text
   */
  async sendDailyReport(summary: string): Promise<void> {
    const now = new Date();
    const title = `GitHub Daily - ${now.toLocaleDateString("pt-BR")}`;

    const payload: DiscordWebhookPayload = {
      embeds: [
        {
          title,
          description: summary,
          color: COLOR_BLUE,
          timestamp: now.toISOString(),
          footer: {
            text: "GitHub Daily Reporter",
          },
        },
      ],
    };

    await this.sendWebhook(payload, "daily report");
  }

  /**
   * Sends a no-activity report when no GitHub events are found.
   */
  async sendNoActivityReport(): Promise<void> {
    const now = new Date();
    const title = `GitHub Daily - ${now.toLocaleDateString("pt-BR")}`;

    const payload: DiscordWebhookPayload = {
      embeds: [
        {
          title,
          description:
            "Hoje foi um dia de planejamento e reflexão no código. Nenhuma atividade registrada no GitHub nas últimas 24 horas.",
          color: COLOR_BLUE,
          timestamp: now.toISOString(),
          footer: {
            text: "GitHub Daily Reporter",
          },
        },
      ],
    };

    await this.sendWebhook(payload, "no activity report");
  }

  /**
   * Sends an error notification to Discord.
   * @param error - The error message to send
   */
  async sendError(error: string): Promise<void> {
    const now = new Date();

    const payload: DiscordWebhookPayload = {
      embeds: [
        {
          title: "GitHub Daily Reporter - Error",
          description: error,
          color: COLOR_RED,
          timestamp: now.toISOString(),
          footer: {
            text: "GitHub Daily Reporter",
          },
        },
      ],
    };

    await this.sendWebhook(payload, "error notification");
  }

  private isValidWebhookUrl(url: string): boolean {
    return DISCORD_WEBHOOK_PATTERN.test(url);
  }

  private async sendWebhook(
    payload: DiscordWebhookPayload,
    operation: string
  ): Promise<void> {
    try {
      await fetchWithRetry(this.webhookUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
        timeout: 15000,
        maxRetries: 2,
      });
    } catch (error) {
      if (error instanceof FetchError) {
        throw new Error(
          `Failed to send ${operation} to Discord: ${error.message}`
        );
      }
      throw new Error(
        `Failed to send ${operation} to Discord: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }
}
