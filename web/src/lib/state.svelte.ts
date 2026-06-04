import type { DiffFile, StatusFile } from './api'

export const appState = $state({
  theme: (localStorage.getItem('theme') as 'dark' | 'light') || 'dark',
  selectedFile: null as StatusFile | null,
  viewMode: 'diff' as 'diff' | 'edit',
  stagedView: false,
  diffContext: 3,
  statusFiles: [] as StatusFile[],
  currentDiff: null as DiffFile | null,
  loading: false,
  searchOpen: false,
  searchMode: 'files' as 'files' | 'content',
  allFiles: [] as string[],
})

export function toggleTheme() {
  appState.theme = appState.theme === 'dark' ? 'light' : 'dark'
  localStorage.setItem('theme', appState.theme)
  document.body.className = appState.theme === 'light' ? 'light' : ''
}
