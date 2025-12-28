const DEFAULT_TIMEOUT = 30000;
const MAX_RETRIES = 3;
const RETRY_DELAY = 1000;

export interface FetchWithRetryOptions extends RequestInit {
  timeout?: number;
  maxRetries?: number;
  retryDelay?: number;
}

export class FetchError extends Error {
  constructor(
    message: string,
    public status?: number,
    public statusText?: string
  ) {
    super(message);
    this.name = "FetchError";
  }
}

function isRetryableStatus(status: number): boolean {
  return status === 429 || status >= 500;
}

async function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function fetchWithRetry(
  url: string,
  options: FetchWithRetryOptions = {}
): Promise<Response> {
  const {
    timeout = DEFAULT_TIMEOUT,
    maxRetries = MAX_RETRIES,
    retryDelay = RETRY_DELAY,
    ...fetchOptions
  } = options;

  let lastError: Error | null = null;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
      const response = await fetch(url, {
        ...fetchOptions,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        if (isRetryableStatus(response.status) && attempt < maxRetries) {
          const delay = retryDelay * 2 ** attempt;
          await sleep(delay);
          continue;
        }

        throw new FetchError(
          `HTTP ${response.status}: ${response.statusText}`,
          response.status,
          response.statusText
        );
      }

      return response;
    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof FetchError) {
        throw error;
      }

      if (error instanceof Error && error.name === "AbortError") {
        lastError = new FetchError(`Request timeout after ${timeout}ms`);
      } else {
        lastError = error instanceof Error ? error : new Error(String(error));
      }

      if (attempt < maxRetries) {
        const delay = retryDelay * 2 ** attempt;
        await sleep(delay);
      }
    }
  }

  throw lastError || new FetchError("Request failed after retries");
}
