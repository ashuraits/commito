export interface PersistedState {
  sidebarMode: 'changes' | 'files'
  selectedFilePath: string | null
  selectedFileStaged: boolean
  viewMode: 'diff' | 'edit'
  sidebarWidth: number
  commitLogHeight: number
  expandedDirs: string[]
}

const DEFAULTS: PersistedState = {
  sidebarMode: 'changes',
  selectedFilePath: null,
  selectedFileStaged: false,
  viewMode: 'diff',
  sidebarWidth: 224,
  commitLogHeight: 90,
  expandedDirs: [],
}

export function load(repo: string): PersistedState {
  try {
    return { ...DEFAULTS, ...JSON.parse(localStorage.getItem(`commito:${repo}`) ?? '{}') }
  } catch {
    return { ...DEFAULTS }
  }
}

export function save(repo: string, patch: Partial<PersistedState>) {
  const current = load(repo)
  localStorage.setItem(`commito:${repo}`, JSON.stringify({ ...current, ...patch }))
}
