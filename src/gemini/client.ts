import { GoogleGenerativeAI } from "@google/generative-ai";
import { buildPrompt } from "./prompt";

const MAX_SUMMARY_CHARS = 4096;

export interface GeminiClientOptions {
  apiKey: string;
  model?: string;
}

export class GeminiClient {
  private genAI: GoogleGenerativeAI;
  private modelName: string;

  constructor(options: GeminiClientOptions) {
    this.genAI = new GoogleGenerativeAI(options.apiKey);
    this.modelName = options.model || "gemini-2.5-flash";
  }

  async generateDailySummary(eventsJson: string): Promise<string> {
    const model = this.genAI.getGenerativeModel({
      model: this.modelName,
      generationConfig: {
        temperature: 1.2,
        maxOutputTokens: 8192,
      },
    });

    const prompt = buildPrompt(eventsJson);
    const result = await model.generateContent(prompt);
    const response = result.response;
    const text = response.text();

    return this.truncateAtSentence(text, MAX_SUMMARY_CHARS);
  }

  private truncateAtSentence(text: string, maxLength: number): string {
    if (text.length <= maxLength) {
      return text;
    }

    const truncated = text.slice(0, maxLength);
    const lastPeriod = truncated.lastIndexOf(".");

    if (lastPeriod > maxLength * 0.5) {
      return truncated.slice(0, lastPeriod + 1);
    }

    return `${truncated.slice(0, -3)}...`;
  }
}
