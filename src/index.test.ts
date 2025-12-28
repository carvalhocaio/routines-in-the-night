import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it } from "bun:test";

describe("Application Integration", () => {
  const originalEnv = { ...Bun.env };

  beforeEach(() => {
    Bun.env.GH_USER = "testuser";
    Bun.env.GH_TOKEN = "testtoken";
    Bun.env.GEMINI_API_KEY = "testapikey";
    Bun.env.DISCORD_WEBHOOK_URL =
      "https://discord.com/api/webhooks/12345678901234567890/abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12345678";
  });

  afterEach(() => {
    Object.assign(Bun.env, originalEnv);
  });

  describe("configuration loading", () => {
    it("should load configuration from environment", async () => {
      const { loadConfig } = await import("./config");
      const config = loadConfig();

      expect(config.ghUser).toBe("testuser");
      expect(config.ghToken).toBe("testtoken");
      expect(config.geminiApiKey).toBe("testapikey");
      expect(config.geminiModel).toBe("gemini-2.5-flash");
    });

    it("should fail with missing environment variables", async () => {
      Bun.env.GH_USER = undefined;
      const { loadConfig, ConfigError } = await import("./config");

      expect(() => loadConfig()).toThrow(ConfigError);
    });
  });

  describe("client initialization", () => {
    it("should create GitHub client with config", async () => {
      const { loadConfig } = await import("./config");
      const { GitHubClient } = await import("./github/client");

      const config = loadConfig();
      const client = new GitHubClient({
        username: config.ghUser,
        token: config.ghToken,
      });

      expect(client).toBeDefined();
      expect(client.getRateLimitInfo()).toBeNull();
    });

    it("should create Gemini client with config", async () => {
      const { loadConfig } = await import("./config");
      const { GeminiClient } = await import("./gemini/client");

      const config = loadConfig();
      const client = new GeminiClient({
        apiKey: config.geminiApiKey,
        model: config.geminiModel,
      });

      expect(client).toBeDefined();
    });

    it("should create Discord client with valid webhook", async () => {
      const { loadConfig } = await import("./config");
      const { DiscordClient } = await import("./discord/client");

      const config = loadConfig();
      const client = new DiscordClient({
        webhookUrl: config.discordWebhookUrl,
      });

      expect(client).toBeDefined();
    });
  });

  describe("workflow integration", () => {
    let mockGitHubServer: ReturnType<typeof Bun.serve>;
    let mockDiscordServer: ReturnType<typeof Bun.serve>;
    let discordPayloads: unknown[] = [];

    beforeAll(() => {
      mockGitHubServer = Bun.serve({
        port: 0,
        fetch(req) {
          const url = new URL(req.url);
          if (url.pathname.includes("/users/") && url.pathname.includes("/events")) {
            return Response.json([
              {
                type: "PushEvent",
                repo: { name: "user/repo" },
                created_at: new Date().toISOString(),
                public: true,
                payload: {
                  ref: "refs/heads/main",
                  commits: [{ message: "test commit", sha: "abc123" }],
                },
              },
            ], {
              headers: {
                "X-RateLimit-Limit": "5000",
                "X-RateLimit-Remaining": "4999",
                "X-RateLimit-Reset": String(Math.floor(Date.now() / 1000) + 3600),
              },
            });
          }
          return new Response("Not Found", { status: 404 });
        },
      });

      mockDiscordServer = Bun.serve({
        port: 0,
        async fetch(req) {
          discordPayloads.push(await req.json());
          return new Response(null, { status: 204 });
        },
      });
    });

    afterAll(() => {
      mockGitHubServer.stop();
      mockDiscordServer.stop();
    });

    beforeEach(() => {
      discordPayloads = [];
    });

    it("should fetch events and track rate limits", async () => {
      const { GitHubClient } = await import("./github/client");

      const client = new GitHubClient({
        username: "testuser",
        token: "testtoken",
        baseUrl: `http://localhost:${mockGitHubServer.port}`,
      });

      const events = await client.getDailyEvents();

      expect(events.length).toBe(1);
      expect(events[0].type).toBe("PushEvent");

      const rateLimit = client.getRateLimitInfo();
      expect(rateLimit).not.toBeNull();
      expect(rateLimit?.remaining).toBe(4999);
    });

    it("should send discord webhook with correct payload", async () => {
      const { DiscordClient } = await import("./discord/client");

      const client = new DiscordClient({
        webhookUrl: `http://localhost:${mockDiscordServer.port}/webhook`,
        skipValidation: true,
      });

      await client.sendDailyReport("Test summary");

      expect(discordPayloads.length).toBe(1);
      const payload = discordPayloads[0] as {
        embeds: Array<{ description: string; color: number }>;
      };
      expect(payload.embeds[0].description).toBe("Test summary");
      expect(payload.embeds[0].color).toBe(0x7289da);
    });

    it("should handle no activity gracefully", async () => {
      const { DiscordClient } = await import("./discord/client");

      const client = new DiscordClient({
        webhookUrl: `http://localhost:${mockDiscordServer.port}/webhook`,
        skipValidation: true,
      });

      await client.sendNoActivityReport();

      expect(discordPayloads.length).toBe(1);
      const payload = discordPayloads[0] as {
        embeds: Array<{ description: string }>;
      };
      expect(payload.embeds[0].description).toContain("planejamento e reflexão");
    });
  });
});
