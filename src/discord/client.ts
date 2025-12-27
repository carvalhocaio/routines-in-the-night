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

    await this.sendWebhook(payload);
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

    await this.sendWebhook(payload);
  }

  private async sendWebhook(payload: DiscordWebhookPayload): Promise<void> {
    const response = await fetch(this.webhookUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      throw new Error(`Discord webhook error: ${response.status}`);
    }
  }
}
