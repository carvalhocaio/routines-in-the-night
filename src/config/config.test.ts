import { afterEach, beforeEach, describe, expect, it } from "bun:test";
import { ConfigError, loadConfig } from "./index";

describe("loadConfig", () => {
  const originalEnv = { ...Bun.env };

  beforeEach(() => {
    Bun.env.GH_USER = "testuser";
    Bun.env.GH_TOKEN = "testtoken";
    Bun.env.GEMINI_API_KEY = "testapikey";
    Bun.env.DISCORD_WEBHOOK_URL = "https://discord.com/webhook";
  });

  afterEach(() => {
    Object.assign(Bun.env, originalEnv);
  });

  it("should load config with all required variables", () => {
    const config = loadConfig();

    expect(config.ghUser).toBe("testuser");
    expect(config.ghToken).toBe("testtoken");
    expect(config.geminiApiKey).toBe("testapikey");
    expect(config.discordWebhookUrl).toBe("https://discord.com/webhook");
  });

  it("should use default gemini model when not specified", () => {
    Bun.env.GEMINI_MODEL = undefined;
    const config = loadConfig();

    expect(config.geminiModel).toBe("gemini-2.5-flash");
  });

  it("should use custom gemini model when specified", () => {
    Bun.env.GEMINI_MODEL = "gemini-pro";
    const config = loadConfig();

    expect(config.geminiModel).toBe("gemini-pro");
  });

  it("should throw ConfigError when GH_USER is missing", () => {
    Bun.env.GH_USER = undefined;

    expect(() => loadConfig()).toThrow(ConfigError);
    expect(() => loadConfig()).toThrow("GH_USER");
  });

  it("should throw ConfigError when GH_TOKEN is missing", () => {
    Bun.env.GH_TOKEN = undefined;

    expect(() => loadConfig()).toThrow(ConfigError);
    expect(() => loadConfig()).toThrow("GH_TOKEN");
  });

  it("should throw ConfigError when GEMINI_API_KEY is missing", () => {
    Bun.env.GEMINI_API_KEY = undefined;

    expect(() => loadConfig()).toThrow(ConfigError);
    expect(() => loadConfig()).toThrow("GEMINI_API_KEY");
  });

  it("should throw ConfigError when DISCORD_WEBHOOK_URL is missing", () => {
    Bun.env.DISCORD_WEBHOOK_URL = undefined;

    expect(() => loadConfig()).toThrow(ConfigError);
    expect(() => loadConfig()).toThrow("DISCORD_WEBHOOK_URL");
  });

  it("should list all missing variables in error message", () => {
    Bun.env.GH_USER = undefined;
    Bun.env.GH_TOKEN = undefined;

    expect(() => loadConfig()).toThrow("GH_USER");
    expect(() => loadConfig()).toThrow("GH_TOKEN");
  });
});
