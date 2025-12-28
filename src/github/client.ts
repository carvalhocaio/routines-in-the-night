import { FetchError, fetchWithRetry } from "../utils/fetch";
import type { FormattedEvent, GitHubEvent } from "./types";

const GITHUB_API_URL = "https://api.github.com";
const GITHUB_API_VERSION = "2022-11-28";

export interface GitHubClientOptions {
  username: string;
  token: string;
  baseUrl?: string;
}

export class GitHubClient {
  private username: string;
  private token: string;
  private baseUrl: string;

  constructor(options: GitHubClientOptions) {
    this.username = options.username;
    this.token = options.token;
    this.baseUrl = options.baseUrl || GITHUB_API_URL;
  }

  async getDailyEvents(): Promise<FormattedEvent[]> {
    const events = await this.fetchUserEvents();
    const yesterday = new Date(Date.now() - 24 * 60 * 60 * 1000);

    const recentEvents = events.filter(
      (event) => new Date(event.created_at) > yesterday
    );

    return this.formatEvents(recentEvents);
  }

  private async fetchUserEvents(): Promise<GitHubEvent[]> {
    const url = `${this.baseUrl}/users/${this.username}/events`;

    try {
      const response = await fetchWithRetry(url, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${this.token}`,
          Accept: "application/vnd.github.v3+json",
          "X-GitHub-Api-Version": GITHUB_API_VERSION,
        },
        timeout: 30000,
        maxRetries: 3,
      });

      return response.json();
    } catch (error) {
      if (error instanceof FetchError) {
        throw new Error(
          `Failed to fetch GitHub events for user '${this.username}': ${error.message}`
        );
      }
      throw new Error(
        `Failed to fetch GitHub events: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  private formatEvents(events: GitHubEvent[]): FormattedEvent[] {
    return events.map((event) => {
      const formatted: FormattedEvent = {
        type: event.type,
        repo: event.repo.name,
        createdAt: event.created_at,
        isPrivate: !event.public,
      };

      switch (event.type) {
        case "PushEvent":
          formatted.commits = event.payload.commits?.length || 0;
          formatted.branch = this.extractBranchName(event.payload.ref || "");
          formatted.commitMessages =
            event.payload.commits?.map((c) => c.message) || [];
          break;

        case "CreateEvent":
        case "DeleteEvent":
          formatted.refType = event.payload.ref_type;
          formatted.ref = event.payload.ref;
          break;

        case "IssuesEvent":
        case "PullRequestEvent":
          formatted.action = event.payload.action;
          if (event.payload.pull_request) {
            formatted.prTitle = event.payload.pull_request.title;
          }
          break;
      }

      return formatted;
    });
  }

  private extractBranchName(ref: string): string {
    const prefix = "refs/heads/";
    if (ref.startsWith(prefix)) {
      return ref.slice(prefix.length);
    }
    return ref;
  }
}
