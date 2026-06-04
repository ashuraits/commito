export interface DiffFile {
  path: string
  oldPath?: string
  status: 'modified' | 'added' | 'deleted' | 'renamed' | 'untracked'
  staged: boolean
  hunks: Hunk[]
  addCount: number
  delCount: number
}

export interface Hunk {
  header: string
  oldStart: number
  oldCount: number
  newStart: number
  newCount: number
  lines: Line[]
}

export interface Line {
  type: 'add' | 'del' | 'context'
  content: string
  oldNo?: number
  newNo?: number
}

export interface StatusFile {
  path: string
  status: string
  staged: boolean
}

export interface Commit {
  hash: string
  short: string
  message: string
  author: string
  date: string
}

export interface SearchResult {
  path: string
  line?: number
  content?: string
  type: 'file' | 'content' | 'folder'
}

const base = ''

export const api = {
  async info(): Promise<{ repoPath: string; branch: string }> {
    const r = await fetch(`${base}/api/info`)
    return r.json()
  },

  async status(): Promise<StatusFile[]> {
    const r = await fetch(`${base}/api/status`)
    return r.json()
  },

  async allFiles(): Promise<string[]> {
    const r = await fetch(`${base}/api/files`)
    return r.json()
  },

  async diff(file: string, staged: boolean, context = 3, untracked = false): Promise<DiffFile> {
    const r = await fetch(`${base}/api/diff?file=${encodeURIComponent(file)}&staged=${staged}&context=${context}&untracked=${untracked}`)
    return r.json()
  },

  async allDiffs(staged: boolean, context = 3): Promise<DiffFile[]> {
    const r = await fetch(`${base}/api/diffs?staged=${staged}&context=${context}`)
    return r.json()
  },

  async readFile(path: string): Promise<string> {
    const r = await fetch(`${base}/api/file?path=${encodeURIComponent(path)}`)
    return r.text()
  },

  async writeFile(path: string, content: string): Promise<void> {
    await fetch(`${base}/api/file?path=${encodeURIComponent(path)}`, {
      method: 'PUT',
      body: content,
    })
  },

  async stage(file: string)      { await fetch(`${base}/api/stage?file=${encodeURIComponent(file)}`, { method: 'POST' }) },
  async unstage(file: string)    { await fetch(`${base}/api/unstage?file=${encodeURIComponent(file)}`, { method: 'POST' }) },
  async revert(file: string)     { await fetch(`${base}/api/revert?file=${encodeURIComponent(file)}`, { method: 'POST' }) },
  async stageAll()               { await fetch(`${base}/api/stage-all`, { method: 'POST' }) },
  async unstageAll()             { await fetch(`${base}/api/unstage-all`, { method: 'POST' }) },
  async revertAll()              { await fetch(`${base}/api/revert-all`, { method: 'POST' }) },

  async searchFiles(q: string): Promise<SearchResult[]> {
    const r = await fetch(`${base}/api/search/files?q=${encodeURIComponent(q)}`)
    return r.json()
  },

  async searchContent(q: string): Promise<SearchResult[]> {
    const r = await fetch(`${base}/api/search/content?q=${encodeURIComponent(q)}`)
    return r.json()
  },

  async commits(): Promise<Commit[]> {
    const r = await fetch(`${base}/api/commits`)
    return r.json()
  },

  async commitDiff(hash: string, context = 3): Promise<DiffFile[]> {
    const r = await fetch(`${base}/api/commit-diff?hash=${hash}&context=${context}`)
    return r.json()
  },
}
