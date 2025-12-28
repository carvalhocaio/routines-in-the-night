import { GoogleGenerativeAI } from "@google/generative-ai";
import { buildPrompt } from "./prompt";

const MAX_SUMMARY_CHARS = 4096;
const MAX_INPUT_SIZE_BYTES = 100000; // 100KB limit for input
const DEFAULT_TEMPERATURE = 0.7; // Lower temperature for more consistent output

export class GeminiValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "GeminiValidationError";
  }
}

export interface GeminiClientOptions {
  apiKey: string;
  model?: string;
  temperature?: number;
}

/**
 * Client for generating AI summaries using Google's Gemini API.
 */
export class GeminiClient {
  private genAI: GoogleGenerativeAI;
  private modelName: string;
  private temperature: number;

  /**
   * Creates a new Gemini client.
   * @param options - Configuration options including API key, optional model name, and temperature
   */
  constructor(options: GeminiClientOptions) {
    this.genAI = new GoogleGenerativeAI(options.apiKey);
    this.modelName = options.model || "gemini-2.5-flash";
    this.temperature = options.temperature ?? DEFAULT_TEMPERATURE;
  }

  /**
   * Generates a daily summary from GitHub events.
   * @param eventsJson - JSON string containing the GitHub events
   * @returns The AI-generated summary text, truncated to fit Discord's limits
   * @throws {GeminiValidationError} If the input is not valid JSON or exceeds size limits
   */
  async generateDailySummary(eventsJson: string): Promise<string> {
    this.validateJsonInput(eventsJson);

    const model = this.genAI.getGenerativeModel({
      model: this.modelName,
      generationConfig: {
        temperature: this.temperature,
        maxOutputTokens: 8192,
      },
    });

    const prompt = buildPrompt(eventsJson);
    const result = await model.generateContent(prompt);
    const response = result.response;
    const text = response.text();

    return this.truncateAtSentence(text, MAX_SUMMARY_CHARS);
  }

  /**
   * Validates that the input is valid JSON and within size limits.
   * @param input - The input string to validate
   * @throws {GeminiValidationError} If the input is not valid JSON or exceeds limits
   */
  private validateJsonInput(input: string): void {
    if (!input || input.trim() === "") {
      throw new GeminiValidationError("Events JSON cannot be empty");
    }

    const inputSizeBytes = new TextEncoder().encode(input).length;
    if (inputSizeBytes > MAX_INPUT_SIZE_BYTES) {
      throw new GeminiValidationError(
        `Input size (${inputSizeBytes} bytes) exceeds maximum allowed (${MAX_INPUT_SIZE_BYTES} bytes)`
      );
    }

    try {
      const parsed = JSON.parse(input);
      if (!Array.isArray(parsed)) {
        throw new GeminiValidationError("Events JSON must be an array");
      }
    } catch (error) {
      if (error instanceof GeminiValidationError) {
        throw error;
      }
      throw new GeminiValidationError(
        `Invalid JSON input: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  /**
   * Truncates text at a sentence boundary to fit within the maximum length.
   * @param text - The text to truncate
   * @param maxLength - The maximum allowed length
   * @returns The truncated text
   */
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
