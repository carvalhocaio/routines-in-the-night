export interface DiscordEmbed {
  title: string;
  description: string;
  color: number;
  timestamp: string;
  footer: {
    text: string;
  };
}

export interface DiscordWebhookPayload {
  embeds: DiscordEmbed[];
}
