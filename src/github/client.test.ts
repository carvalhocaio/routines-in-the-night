import { afterAll, beforeAll, beforeEach, describe, expect, it, spyOn } from "bun:test";
import { GitHubClient, GitHubRateLimitError } from "./client";

describe("GitHubClient", () => {
  let mockServer: ReturnType<typeof Bun.serve>;
  let baseUrl: string;

  beforeAll(() => {
    mockServer = Bun.serve({
      port: 0,
      fetch(req) {
        const url = new URL(req.url);

        if (url.pathname === "/users/testuser/events") {
          const now = new Date();
          const yesterday = new Date(now.getTime() - 12 * 60 * 60 * 1000);
          const twoDaysAgo = new Date(now.getTime() - 48 * 60 * 60 * 1000);

          return Response.json([
            {
              type: "PushEvent",
              repo: { name: "user/repo1" },
              created_at: yesterday.toISOString(),
              public: true,
              payload: {
                ref: "refs/heads/main",
                commits: [
                  { message: "feat: add feature", sha: "abc123" },
                  { message: "fix: bug fix", sha: "def456" },
                ],
              },
            },
            {
              type: "CreateEvent",
              repo: { name: "user/repo2" },
              created_at: yesterday.toISOString(),
              public: false,
              payload: {
                ref_type: "branch",
                ref: "feature/new-feature",
              },
            },
            {
              type: "PullRequestEvent",
              repo: { name: "user/repo1" },
              created_at: yesterday.toISOString(),
              public: true,
              payload: {
                action: "opened",
                pull_request: {
                  title: "Add new feature",
                },
              },
            },
            {
              type: "PushEvent",
              repo: { name: "user/old-repo" },
              created_at: twoDaysAgo.toISOString(),
              public: true,
              payload: {
                ref: "refs/heads/main",
                commits: [{ message: "old commit", sha: "old123" }],
              },
            },
          ]);
        }

        if (url.pathname === "/users/emptyuser/events") {
          return Response.json([]);
        }

        if (url.pathname === "/users/erroruser/events") {
          return new Response("Unauthorized", { status: 401 });
        }

        if (url.pathname === "/users/ratelimituser/events") {
          return Response.json([], {
            headers: {
              "X-RateLimit-Limit": "5000",
              "X-RateLimit-Remaining": "50",
              "X-RateLimit-Reset": String(Math.floor(Date.now() / 1000) + 3600),
            },
          });
        }

        return new Response("Not Found", { status: 404 });
      },
    });
    baseUrl = `http://localhost:${mockServer.port}`;
  });

  afterAll(() => {
    mockServer.stop();
  });

  it("should fetch and filter events from last 24 hours", async () => {
    const client = new GitHubClient({
      username: "testuser",
      token: "testtoken",
      baseUrl,
    });

    const events = await client.getDailyEvents();

    expect(events.length).toBe(3);
    expect(events.every((e) => e.type !== "PushEvent" || e.repo !== "user/old-repo")).toBe(true);
  });

  it("should format PushEvent correctly", async () => {
    const client = new GitHubClient({
      username: "testuser",
      token: "testtoken",
      baseUrl,
    });

    const events = await client.getDailyEvents();
    const pushEvent = events.find((e) => e.type === "PushEvent");

    expect(pushEvent).toBeDefined();
    expect(pushEvent?.commits).toBe(2);
    expect(pushEvent?.branch).toBe("main");
    expect(pushEvent?.commitMessages).toEqual([
      "feat: add feature",
      "fix: bug fix",
    ]);
  });

  it("should format CreateEvent correctly", async () => {
    const client = new GitHubClient({
      username: "testuser",
      token: "testtoken",
      baseUrl,
    });

    const events = await client.getDailyEvents();
    const createEvent = events.find((e) => e.type === "CreateEvent");

    expect(createEvent).toBeDefined();
    expect(createEvent?.refType).toBe("branch");
    expect(createEvent?.ref).toBe("feature/new-feature");
    expect(createEvent?.isPrivate).toBe(true);
  });

  it("should format PullRequestEvent correctly", async () => {
    const client = new GitHubClient({
      username: "testuser",
      token: "testtoken",
      baseUrl,
    });

    const events = await client.getDailyEvents();
    const prEvent = events.find((e) => e.type === "PullRequestEvent");

    expect(prEvent).toBeDefined();
    expect(prEvent?.action).toBe("opened");
    expect(prEvent?.prTitle).toBe("Add new feature");
  });

  it("should return empty array when no events", async () => {
    const client = new GitHubClient({
      username: "emptyuser",
      token: "testtoken",
      baseUrl,
    });

    const events = await client.getDailyEvents();

    expect(events).toEqual([]);
  });

  it("should throw error on API failure", async () => {
    const client = new GitHubClient({
      username: "erroruser",
      token: "testtoken",
      baseUrl,
    });

    await expect(client.getDailyEvents()).rejects.toThrow(
      "Failed to fetch GitHub events for user 'erroruser'"
    );
  });

  describe("rate limit handling", () => {
    it("should track rate limit info from response headers", async () => {
      const client = new GitHubClient({
        username: "ratelimituser",
        token: "testtoken",
        baseUrl,
      });

      expect(client.getRateLimitInfo()).toBeNull();

      await client.getDailyEvents();

      const rateLimitInfo = client.getRateLimitInfo();
      expect(rateLimitInfo).not.toBeNull();
      expect(rateLimitInfo?.limit).toBe(5000);
      expect(rateLimitInfo?.remaining).toBe(50);
      expect(rateLimitInfo?.reset).toBeInstanceOf(Date);
    });

    it("should throw GitHubRateLimitError when rate limit is exhausted before request", async () => {
      const client = new GitHubClient({
        username: "ratelimituser",
        token: "testtoken",
        baseUrl,
      });

      // First request to populate rate limit info
      await client.getDailyEvents();

      // Manually set rate limit to 0 with future reset
      const futureReset = new Date(Date.now() + 3600000);
      // Access private property for testing
      (client as unknown as { lastRateLimitInfo: { limit: number; remaining: number; reset: Date } }).lastRateLimitInfo = {
        limit: 5000,
        remaining: 0,
        reset: futureReset,
      };

      await expect(client.getDailyEvents()).rejects.toThrow(GitHubRateLimitError);
      await expect(client.getDailyEvents()).rejects.toThrow("rate limit exceeded");
    });

    it("should warn when approaching rate limit", async () => {
      const client = new GitHubClient({
        username: "ratelimituser",
        token: "testtoken",
        baseUrl,
      });

      // First request to populate rate limit info
      await client.getDailyEvents();

      // Spy on console.warn
      const warnSpy = spyOn(console, "warn");

      // Set rate limit to a low value (below threshold of 100)
      (client as unknown as { lastRateLimitInfo: { limit: number; remaining: number; reset: Date } }).lastRateLimitInfo = {
        limit: 5000,
        remaining: 50,
        reset: new Date(Date.now() + 3600000),
      };

      // This should trigger the warning
      await client.getDailyEvents();

      expect(warnSpy).toHaveBeenCalled();
      expect(warnSpy.mock.calls[0][0]).toContain("rate limit warning");
      expect(warnSpy.mock.calls[0][0]).toContain("50 requests remaining");

      warnSpy.mockRestore();
    });

    it("should include reset time in GitHubRateLimitError", async () => {
      const client = new GitHubClient({
        username: "ratelimituser",
        token: "testtoken",
        baseUrl,
      });

      await client.getDailyEvents();

      const futureReset = new Date(Date.now() + 3600000);
      (client as unknown as { lastRateLimitInfo: { limit: number; remaining: number; reset: Date } }).lastRateLimitInfo = {
        limit: 5000,
        remaining: 0,
        reset: futureReset,
      };

      try {
        await client.getDailyEvents();
        expect(true).toBe(false); // Should not reach here
      } catch (error) {
        expect(error).toBeInstanceOf(GitHubRateLimitError);
        expect((error as GitHubRateLimitError).resetAt).toEqual(futureReset);
      }
    });
  });
});
