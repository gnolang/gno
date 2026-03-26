export interface ParsedGithubTarget {
  owner: string
  repo: string
  branch?: string
  filePath?: string
}

/**
 * Parse GitHub repository URLs into constituent parts.
 * Supports repo root, tree branch, and blob file URLs.
 */
const GITHUB_URL_PATTERN =
  /^(?:https?:\/\/)?(?:www\.)?github\.com\/(?<owner>[\w.-]+)\/(?<repo>[\w.-]+)(?:\/(?<type>blob|tree)\/(?<branch>[\w./-]+?)(?:\/(?<path>.+))?)?(?:#.*)?$/i

export function parseGithubUrl(raw: string): ParsedGithubTarget | null {
  if (!raw) return null

  const match = raw.trim().match(GITHUB_URL_PATTERN)
  if (!match || !match.groups) return null

  const { owner, repo, type, branch, path } = match.groups
  if (!owner || !repo) return null

  if (type === 'blob') {
    if (!path) return null
    return { owner, repo, branch: (branch || 'main').replace(/\/$/, ''), filePath: path }
  }

  if (type === 'tree') {
    return { owner, repo, branch: branch?.replace(/\/$/, '') || 'main' }
  }

  return { owner, repo }
}

/**
 * Build a Playground GitHub import URL using the configured public base.
 */
export function buildPlaygroundGithubUrl(base: string | undefined, target: ParsedGithubTarget): string {
  const { origin } = new URL(import.meta.url)
  const url = new URL(`/github/${target.owner}/${target.repo}`, base || origin)

  if (target.filePath) {
    url.searchParams.set('file', target.filePath)
    url.searchParams.set('branch', target.branch || 'main')
  }

  return url.href
}
