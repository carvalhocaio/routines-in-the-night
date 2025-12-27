import { describe, expect, it, mock } from "bun:test";
import { buildPrompt } from "./prompt";

describe("buildPrompt", () => {
  it("should insert events into prompt template", () => {
    const events = '{"type": "PushEvent"}';
    const prompt = buildPrompt(events);

    expect(prompt).toContain(events);
    expect(prompt).toContain("Você é um assistente técnico");
    expect(prompt).toContain("REQUISITOS OBRIGATÓRIOS");
  });
});

describe("GeminiClient", () => {
  it("should truncate long text at sentence boundary", async () => {
    const mockGenerateContent = mock(() =>
      Promise.resolve({
        response: {
          text: () =>
            "This is a long text. It has multiple sentences. This is the third sentence. And the fourth one here.",
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

    const { GeminiClient } = await import("./client");
    const client = new GeminiClient({ apiKey: "test-key" });

    const result = await client.generateDailySummary('{"test": true}');

    expect(result).toBeTruthy();
    expect(mockGenerateContent).toHaveBeenCalled();
  });
});
