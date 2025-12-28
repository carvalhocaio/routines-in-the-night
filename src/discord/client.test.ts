import { afterAll, beforeAll, describe, expect, it } from "bun:test";
import { DiscordClient, DiscordWebhookError } from "./client";

describe("DiscordClient", () => {
  describe("webhook URL validation", () => {
    it("should accept valid Discord webhook URL", () => {
      const client = new DiscordClient({
        webhookUrl: "https://discord.com/api/webhooks/1234567890/abcdefghijklmnop",
      });
      expect(client).toBeDefined();
    });

    it("should reject invalid webhook URL", () => {
      expect(() => {
        new DiscordClient({
          webhookUrl: "https://example.com/webhook",
        });
      }).toThrow(DiscordWebhookError);
    });

    it("should reject HTTP webhook URL", () => {
      expect(() => {
        new DiscordClient({
          webhookUrl: "http://discord.com/api/webhooks/123/abc",
        });
      }).toThrow(DiscordWebhookError);
    });

    it("should skip validation when skipValidation is true", () => {
      const client = new DiscordClient({
        webhookUrl: "http://localhost:3000/webhook",
        skipValidation: true,
      });
      expect(client).toBeDefined();
    });
  });

  describe("webhook operations", () => {
    let mockServer: ReturnType<typeof Bun.serve>;
    let webhookUrl: string;
    let lastPayload: unknown = null;

    beforeAll(() => {
      mockServer = Bun.serve({
        port: 0,
        async fetch(req) {
          const url = new URL(req.url);

          if (url.pathname === "/webhook/success") {
            lastPayload = await req.json();
            return new Response(null, { status: 204 });
          }

          if (url.pathname === "/webhook/error") {
            return new Response("Bad Request", { status: 400 });
          }

          return new Response("Not Found", { status: 404 });
        },
      });
      webhookUrl = `http://localhost:${mockServer.port}`;
    });

    afterAll(() => {
      mockServer.stop();
    });

    it("should send daily report with correct embed format", async () => {
      const client = new DiscordClient({
        webhookUrl: `${webhookUrl}/webhook/success`,
        skipValidation: true,
      });

      await client.sendDailyReport("Test summary content");

      expect(lastPayload).toBeDefined();
      const payload = lastPayload as {
        embeds: Array<{
          title: string;
          description: string;
          color: number;
          footer: { text: string };
        }>;
      };
      expect(payload.embeds).toHaveLength(1);
      expect(payload.embeds[0].description).toBe("Test summary content");
      expect(payload.embeds[0].color).toBe(0x7289da);
      expect(payload.embeds[0].footer.text).toBe("GitHub Daily Reporter");
      expect(payload.embeds[0].title).toContain("GitHub Daily");
    });

    it("should send error with red color", async () => {
      const client = new DiscordClient({
        webhookUrl: `${webhookUrl}/webhook/success`,
        skipValidation: true,
      });

      await client.sendError("Test error message");

      expect(lastPayload).toBeDefined();
      const payload = lastPayload as {
        embeds: Array<{ title: string; description: string; color: number }>;
      };
      expect(payload.embeds[0].description).toBe("Test error message");
      expect(payload.embeds[0].color).toBe(0xff0000);
      expect(payload.embeds[0].title).toBe("GitHub Daily Reporter - Error");
    });

    it("should send no activity report", async () => {
      const client = new DiscordClient({
        webhookUrl: `${webhookUrl}/webhook/success`,
        skipValidation: true,
      });

      await client.sendNoActivityReport();

      expect(lastPayload).toBeDefined();
      const payload = lastPayload as {
        embeds: Array<{ title: string; description: string; color: number }>;
      };
      expect(payload.embeds[0].description).toContain(
        "planejamento e reflexão"
      );
      expect(payload.embeds[0].color).toBe(0x7289da);
    });

    it("should throw error on webhook failure", async () => {
      const client = new DiscordClient({
        webhookUrl: `${webhookUrl}/webhook/error`,
        skipValidation: true,
      });

      await expect(client.sendDailyReport("Test")).rejects.toThrow(
        "Failed to send daily report to Discord"
      );
    });
  });
});
