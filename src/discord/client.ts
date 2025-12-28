import { FetchError, fetchWithRetry } from "../utils/fetch";
import type { DiscordWebhookPayload } from "./types";

const COLOR_BLUE = 0x7289da;
const COLOR_RED = 0xff0000;

export interface DiscordClientOptions {
  webhookUrl: string;
}

export class DiscordClient {
  private webhookUrl: string;

  constructor(options: DiscordClientOptions) {
    this.webhookUrl = options.webhookUrl;
  }

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
