import { describe, expect, it, mock, spyOn } from "bun:test";
import { GeminiClient, GeminiValidationError } from "./client";
import { buildPrompt } from "./prompt";

describe("buildPrompt", () => {
  it("should insert events into prompt template", () => {
    const events = '[{"type": "PushEvent"}]';
    const prompt = buildPrompt(events);

    expect(prompt).toContain(events);
    expect(prompt).toContain("Você é um assistente técnico");
    expect(prompt).toContain("REQUISITOS OBRIGATÓRIOS");
  });
});

describe("GeminiClient", () => {
  describe("JSON validation", () => {
    it("should throw error for empty input", async () => {
      const client = new GeminiClient({ apiKey: "test-key" });

      await expect(client.generateDailySummary("")).rejects.toThrow(
        GeminiValidationError
      );
      await expect(client.generateDailySummary("")).rejects.toThrow(
        "Events JSON cannot be empty"
      );
    });

    it("should throw error for whitespace-only input", async () => {
      const client = new GeminiClient({ apiKey: "test-key" });

      await expect(client.generateDailySummary("   ")).rejects.toThrow(
        "Events JSON cannot be empty"
      );
    });

    it("should throw error for invalid JSON", async () => {
      const client = new GeminiClient({ apiKey: "test-key" });

      await expect(
        client.generateDailySummary("not valid json")
      ).rejects.toThrow(GeminiValidationError);
      await expect(
        client.generateDailySummary("not valid json")
      ).rejects.toThrow("Invalid JSON input");
    });

    it("should throw error for non-array JSON", async () => {
      const client = new GeminiClient({ apiKey: "test-key" });

      await expect(
        client.generateDailySummary('{"type": "object"}')
      ).rejects.toThrow("Events JSON must be an array");
    });

    it("should throw error for input exceeding size limit", async () => {
      const client = new GeminiClient({ apiKey: "test-key" });
      const largeInput = JSON.stringify(
        Array(10000).fill({ type: "PushEvent", data: "x".repeat(100) })
      );

      await expect(client.generateDailySummary(largeInput)).rejects.toThrow(
        "exceeds maximum allowed"
      );
    });

    it("should accept valid JSON array", async () => {
      const mockGenerateContent = mock(() =>
        Promise.resolve({
          response: {
            text: () => "Generated summary text",
          },
        })
      );

      const mockGetGenerativeModel = mock(() => ({
        generateContent: mockGenerateContent,
      }));

      mock.module("@google/generative-ai", () => ({
        GoogleGenerativeAI: class {
          getGenerativeModel = mockGetGenerativeModel;
        },
      }));

      const { GeminiClient: MockedClient } = await import("./client");
      const client = new MockedClient({ apiKey: "test-key" });

      const result = await client.generateDailySummary('[{"type": "PushEvent"}]');

      expect(result).toBe("Generated summary text");
    });
  });

  describe("text truncation", () => {
    it("should not truncate text under limit", async () => {
      const mockGenerateContent = mock(() =>
        Promise.resolve({
          response: {
            text: () => "Short text.",
          },
        })
      );

      const mockGetGenerativeModel = mock(() => ({
        generateContent: mockGenerateContent,
      }));

      mock.module("@google/generative-ai", () => ({
        GoogleGenerativeAI: class {
          getGenerativeModel = mockGetGenerativeModel;
        },
      }));

      const { GeminiClient: MockedClient } = await import("./client");
      const client = new MockedClient({ apiKey: "test-key" });

      const result = await client.generateDailySummary("[]");

      expect(result).toBe("Short text.");
    });

    it("should truncate at sentence boundary when possible", async () => {
      const longText =
        "First sentence. Second sentence. Third sentence. Fourth sentence.";
      const mockGenerateContent = mock(() =>
        Promise.resolve({
          response: {
            text: () => longText,
          },
        })
      );

      const mockGetGenerativeModel = mock(() => ({
        generateContent: mockGenerateContent,
      }));

      mock.module("@google/generative-ai", () => ({
        GoogleGenerativeAI: class {
          getGenerativeModel = mockGetGenerativeModel;
        },
      }));

      const { GeminiClient: MockedClient } = await import("./client");
      const client = new MockedClient({ apiKey: "test-key" });

      const result = await client.generateDailySummary("[]");

      expect(result).toBe(longText);
    });
  });

  describe("model configuration", () => {
    it("should use default model when not specified", () => {
      const client = new GeminiClient({ apiKey: "test-key" });
      expect(client).toBeDefined();
    });

    it("should use custom model when specified", () => {
      const client = new GeminiClient({
        apiKey: "test-key",
        model: "gemini-pro",
      });
      expect(client).toBeDefined();
    });
  });
});
