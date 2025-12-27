export interface GitHubEvent {
  type: string;
  repo: {
    name: string;
  };
  created_at: string;
  public: boolean;
  payload: {
    ref?: string;
    ref_type?: string;
    action?: string;
    commits?: Array<{
      message: string;
      sha: string;
    }>;
    pull_request?: {
      title: string;
    };
  };
}

export interface FormattedEvent {
  type: string;
  repo: string;
  createdAt: string;
  isPrivate: boolean;
  branch?: string;
  commits?: number;
  commitMessages?: string[];
  refType?: string;
  ref?: string;
  action?: string;
  prTitle?: string;
}
