import { afterAll, beforeAll, describe, expect, it } from "bun:test";
import { DiscordClient } from "./client";

describe("DiscordClient", () => {
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
    });

    await client.sendDailyReport("Test summary content");

    expect(lastPayload).toBeDefined();
    const payload = lastPayload as { embeds: Array<{ title: string; description: string; color: number; footer: { text: string } }> };
    expect(payload.embeds).toHaveLength(1);
    expect(payload.embeds[0].description).toBe("Test summary content");
    expect(payload.embeds[0].color).toBe(0x7289da);
    expect(payload.embeds[0].footer.text).toBe("GitHub Daily Reporter");
    expect(payload.embeds[0].title).toContain("GitHub Daily");
  });

  it("should send error with red color", async () => {
    const client = new DiscordClient({
      webhookUrl: `${webhookUrl}/webhook/success`,
    });

    await client.sendError("Test error message");

    expect(lastPayload).toBeDefined();
    const payload = lastPayload as { embeds: Array<{ title: string; description: string; color: number }> };
    expect(payload.embeds[0].description).toBe("Test error message");
    expect(payload.embeds[0].color).toBe(0xff0000);
    expect(payload.embeds[0].title).toBe("GitHub Daily Reporter - Error");
  });

  it("should throw error on webhook failure", async () => {
    const client = new DiscordClient({
      webhookUrl: `${webhookUrl}/webhook/error`,
    });

    await expect(client.sendDailyReport("Test")).rejects.toThrow(
      "Discord webhook error: 400"
    );
  });
});
