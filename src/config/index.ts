export interface Config {
  ghUser: string;
  ghToken: string;
  geminiApiKey: string;
  geminiModel: string;
  discordWebhookUrl: string;
}

export class ConfigError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "ConfigError";
  }
}

export function loadConfig(): Config {
  const ghUser = Bun.env.GH_USER;
  const ghToken = Bun.env.GH_TOKEN;
  const geminiApiKey = Bun.env.GEMINI_API_KEY;
  const geminiModel = Bun.env.GEMINI_MODEL || "gemini-2.5-flash";
  const discordWebhookUrl = Bun.env.DISCORD_WEBHOOK_URL;

  const missingVars: string[] = [];

  if (!ghUser) missingVars.push("GH_USER");
  if (!ghToken) missingVars.push("GH_TOKEN");
  if (!geminiApiKey) missingVars.push("GEMINI_API_KEY");
  if (!discordWebhookUrl) missingVars.push("DISCORD_WEBHOOK_URL");

  if (missingVars.length > 0) {
    throw new ConfigError(
      `Missing required environment variables: ${missingVars.join(", ")}`
    );
  }

  return {
    ghUser: ghUser as string,
    ghToken: ghToken as string,
    geminiApiKey: geminiApiKey as string,
    geminiModel,
    discordWebhookUrl: discordWebhookUrl as string,
  };
}
