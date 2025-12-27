import { afterAll, beforeAll, describe, expect, it } from "bun:test";
import { GitHubClient } from "./client";

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

    await expect(client.getDailyEvents()).rejects.toThrow("GitHub API error: 401");
  });
});
