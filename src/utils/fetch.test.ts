import { afterAll, beforeAll, describe, expect, it } from "bun:test";
import { FetchError, fetchWithRetry } from "./fetch";

describe("fetchWithRetry", () => {
  let mockServer: ReturnType<typeof Bun.serve>;
  let baseUrl: string;
  let requestCount: number;

  beforeAll(() => {
    requestCount = 0;
    mockServer = Bun.serve({
      port: 0,
      fetch(req) {
        const url = new URL(req.url);

        if (url.pathname === "/success") {
          return Response.json({ message: "ok" });
        }

        if (url.pathname === "/error") {
          return new Response("Bad Request", { status: 400 });
        }

        if (url.pathname === "/retry") {
          requestCount++;
          if (requestCount < 3) {
            return new Response("Server Error", { status: 500 });
          }
          return Response.json({ message: "success after retry" });
        }

        if (url.pathname === "/always-fail") {
          return new Response("Server Error", { status: 500 });
        }

        return new Response("Not Found", { status: 404 });
      },
    });
    baseUrl = `http://localhost:${mockServer.port}`;
  });

  afterAll(() => {
    mockServer.stop();
  });

  it("should fetch successfully", async () => {
    const response = await fetchWithRetry(`${baseUrl}/success`);
    const data = await response.json();

    expect(data.message).toBe("ok");
  });

  it("should throw FetchError on non-retryable error", async () => {
    await expect(fetchWithRetry(`${baseUrl}/error`)).rejects.toThrow(FetchError);
    await expect(fetchWithRetry(`${baseUrl}/error`)).rejects.toThrow(
      "HTTP 400: Bad Request"
    );
  });

  it("should retry on 500 errors and succeed", async () => {
    requestCount = 0;
    const response = await fetchWithRetry(`${baseUrl}/retry`, {
      maxRetries: 3,
      retryDelay: 10,
    });
    const data = await response.json();

    expect(data.message).toBe("success after retry");
    expect(requestCount).toBe(3);
  });

  it("should throw after max retries exceeded", async () => {
    await expect(
      fetchWithRetry(`${baseUrl}/always-fail`, {
        maxRetries: 2,
        retryDelay: 10,
      })
    ).rejects.toThrow(FetchError);
  });

});
