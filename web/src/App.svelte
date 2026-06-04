<script lang="ts">
  import { onMount } from 'svelte'
  import { api } from './lib/api'
  import type { StatusFile, SearchResult, Commit, DiffFile } from './lib/api'
  import { appState, toggleTheme } from './lib/state.svelte'
  import DiffTree from './lib/components/DiffTree.svelte'
  import ProjectTree from './lib/components/ProjectTree.svelte'
  import DiffViewer from './lib/components/DiffViewer.svelte'
  import Editor from './lib/components/Editor.svelte'
  import SearchOverlay from './lib/components/SearchOverlay.svelte'
  import ThemeToggle from './lib/components/ThemeToggle.svelte'
  import CommitLog from './lib/components/CommitLog.svelte'
  import CommitDiffViewer from './lib/components/CommitDiffViewer.svelte'
  import MediaViewer from './lib/components/MediaViewer.svelte'

  // apply saved theme
  if (appState.theme === 'light') document.body.classList.add('light')

  const MEDIA_EXT = new Set(['png','jpg','jpeg','gif','webp','svg','bmp','ico','avif','mp4','webm','mov','mkv','avi','mp3','wav','flac','aac','m4a','opus'])
  function isMedia(path: string) {
    return MEDIA_EXT.has(path.split('.').pop()?.toLowerCase() ?? '')
  }

  const diffCache = new Map<string, import('./lib/api').DiffFile>()

  function diffKey(file: StatusFile) {
    return `${file.path}:${file.staged}`
  }

  async function openFile(path: string) {
    const changed = appState.statusFiles.find(f => f.path === path)
    if (changed) {
      selectFile(changed)
    } else {
      appState.selectedFile = { path, status: 'modified', staged: false }
      appState.currentDiff = null
      appState.viewMode = 'edit'
    }
  }

  async function loadStatus() {
    const files = await api.status().catch(() => null)
    if (!files) return
    appState.statusFiles = files
    api.allFiles().then(f => { if (f) { appState.allFiles = f.tracked; ignoredPaths = f.ignored } }).catch(() => {})
    const [stagedDiffs, unstagedDiffs] = await Promise.all([
      api.allDiffs(true).catch(() => []),
      api.allDiffs(false).catch(() => []),
    ])
    for (const d of [...(stagedDiffs ?? []), ...(unstagedDiffs ?? [])]) {
      diffCache.set(`${d.path}:${d.staged}`, d)
    }
    // clear diff if selected file no longer has changes
    if (appState.selectedFile && appState.viewMode === 'diff') {
      const still = files.find(f => f.path === appState.selectedFile?.path && f.staged === appState.selectedFile?.staged)
      if (!still) {
        appState.selectedFile = null
        appState.currentDiff = null
      }
    }
  }

  async function selectFile(file: StatusFile) {
    revertConfirm = false
    appState.selectedFile = file
    appState.viewMode = 'diff'
    appState.diffContext = 3
    const cached = diffCache.get(diffKey(file))
    if (cached) {
      appState.currentDiff = cached
      appState.loading = false
    } else {
      appState.loading = true
      appState.currentDiff = await api.diff(file.path, file.staged, appState.diffContext, file.status === 'untracked')
      diffCache.set(diffKey(file), appState.currentDiff)
      appState.loading = false
    }
  }

  let focusFolder = $state<string | null>(null)

  async function handleSearchSelect(result: SearchResult) {
    appState.searchOpen = false
    if (result.type === 'folder') {
      sidebarMode = 'files'
      focusFolder = result.path
      return
    }
    const file = appState.statusFiles.find(f => f.path === result.path)
    if (file) {
      selectFile(file)
    } else {
      appState.selectedFile = { path: result.path, status: 'modified', staged: false }
      appState.viewMode = 'edit'
    }
  }

  function treeOrder(files: StatusFile[]): StatusFile[] {
    interface Node { name: string; file?: StatusFile; children: Map<string, Node> }
    const root: Node = { name: '', children: new Map() }
    for (const f of files) {
      const parts = f.path.split('/')
      let node = root
      for (let i = 0; i < parts.length; i++) {
        const part = parts[i]
        if (!node.children.has(part)) node.children.set(part, { name: part, children: new Map() })
        node = node.children.get(part)!
        if (i === parts.length - 1) node.file = f
      }
    }
    const result: StatusFile[] = []
    function walk(node: Node) {
      const sorted = [...node.children.values()].sort((a, b) => {
        const aDir = !a.file && a.children.size > 0 ? 0 : 1
        const bDir = !b.file && b.children.size > 0 ? 0 : 1
        return aDir - bDir || a.name.localeCompare(b.name)
      })
      for (const child of sorted) {
        if (child.file) result.push(child.file)
        else walk(child)
      }
    }
    walk(root)
    return result
  }

  function navigateFiles(dir: 'up' | 'down') {
    if (appState.statusFiles.length === 0) return
    const staged   = treeOrder(appState.statusFiles.filter(f => f.staged))
    const unstaged = treeOrder(appState.statusFiles.filter(f => !f.staged))
    const all = [...staged, ...unstaged]
    const cur = all.findIndex(f => f.path === appState.selectedFile?.path && f.staged === appState.selectedFile?.staged)
    const next = cur === -1
      ? (dir === 'down' ? 0 : all.length - 1)
      : (dir === 'down' ? Math.min(all.length - 1, cur + 1) : Math.max(0, cur - 1))
    selectFile(all[next])
  }

  let revertConfirm = $state(false)

  async function doRevert() {
    if (!appState.selectedFile) return
    revertConfirm = false
    await api.revert(appState.selectedFile.path)
    await loadStatus()
    appState.selectedFile = null
    appState.currentDiff = null
  }

  function handleKeydown(e: KeyboardEvent) {
    const meta = e.metaKey || e.ctrlKey

    if ((e.key === 'ArrowUp' || e.key === 'ArrowDown') && !appState.searchOpen && sidebarMode === 'changes') {
      e.preventDefault()
      navigateFiles(e.key === 'ArrowDown' ? 'down' : 'up')
      return
    }
    if (meta && e.key === 'p') {
      e.preventDefault()
      appState.searchOpen = true
      appState.searchMode = 'files'
      return
    }
    if (meta && e.key === 'f' && !appState.searchOpen) {
      e.preventDefault()
      appState.searchOpen = true
      appState.searchMode = 'content'
      return
    }
    if (meta && e.shiftKey && e.key === 'T') {
      e.preventDefault()
      toggleTheme()
      return
    }
    if (e.key === 'e' && !appState.searchOpen && appState.selectedFile && appState.viewMode === 'diff') {
      appState.viewMode = 'edit'
      return
    }
    if (e.key === 'd' && !appState.searchOpen && appState.selectedFile && !revertConfirm) {
      const f = appState.selectedFile
      if (f.status !== 'untracked' && !f.staged) { revertConfirm = true; return }
    }
    if (revertConfirm) {
      if (e.key === 'Enter' || e.key === 'y') { e.preventDefault(); doRevert(); return }
      if (e.key === 'Escape') { revertConfirm = false; return }
    }
    if (e.key === 'Escape') {
      if (appState.searchOpen) { appState.searchOpen = false; return }
      if (appState.viewMode === 'edit') { appState.viewMode = 'diff'; return }
    }
    if (e.key === 'r' && meta) {
      e.preventDefault()
      loadStatus()
    }
  }

  let repoName = $state('')
  let branch = $state('')
  let ignoredPaths = $state<string[]>([])
  let sidebarMode = $state<'changes' | 'files'>('changes')
  let selectedCommit = $state<Commit | null>(null)
  let commitDiffs = $state<DiffFile[]>([])
  let commitLoading = $state(false)
  let commitLogHeight = $state(parseInt(localStorage.getItem('commitLogHeight') || '90'))

  function startCommitResize(e: MouseEvent) {
    e.preventDefault()
    const startY = e.clientY
    const startH = commitLogHeight
    function onMove(e: MouseEvent) {
      commitLogHeight = Math.min(500, Math.max(60, startH - (e.clientY - startY)))
    }
    function onUp() {
      localStorage.setItem('commitLogHeight', String(commitLogHeight))
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
  }

  async function selectCommit(c: Commit) {
    selectedCommit = c
    appState.selectedFile = null
    appState.currentDiff = null
    commitLoading = true
    commitDiffs = await api.commitDiff(c.hash, appState.diffContext)
    commitLoading = false
  }
  let sidebarWidth = $state(parseInt(localStorage.getItem('sidebarWidth') || '224'))
  let resizing = $state(false)

  function startResize(e: MouseEvent) {
    resizing = true
    const startX = e.clientX
    const startW = sidebarWidth
    function onMove(e: MouseEvent) {
      sidebarWidth = Math.min(480, Math.max(120, startW + e.clientX - startX))
    }
    function onUp() {
      resizing = false
      localStorage.setItem('sidebarWidth', String(sidebarWidth))
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
  }

  onMount(async () => {
    const info = await api.info()
    repoName = info.repoPath.split('/').pop() || info.repoPath
    branch = info.branch
    document.title = `${repoName} — commito`
    loadStatus()
    const interval = setInterval(loadStatus, 3000)
    return () => clearInterval(interval)
  })
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="h-full flex flex-col {appState.theme === 'light' ? 'bg-white text-gray-900' : 'bg-[#0d1117] text-gray-200'}">
  <!-- topbar -->
  <div class="flex items-center gap-2 px-3 py-1.5 border-b {appState.theme === 'light' ? 'border-gray-200 bg-gray-50' : 'border-gray-800 bg-gray-900/60'} shrink-0 text-[12px]">
    <span class="font-semibold text-blue-400 select-none">commito</span>
    {#if repoName}
      <span class="text-gray-500 ml-1 select-none">~/…/{repoName}</span>
    {/if}
    <div class="ml-auto flex items-center gap-1">
      <button
        class="px-2 py-0.5 rounded text-[11px] {appState.theme === 'light' ? 'text-gray-500 hover:text-gray-800 hover:bg-gray-200' : 'text-gray-500 hover:text-gray-300 hover:bg-gray-800'} transition-colors flex items-center gap-1"
        onclick={() => { appState.searchOpen = true; appState.searchMode = 'files' }}
      >
        <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <span>Search</span>
        <kbd class="ml-0.5 opacity-50">⌘P</kbd>
      </button>
      <ThemeToggle />
      <button
        class="p-1.5 rounded {appState.theme === 'light' ? 'text-gray-500 hover:text-gray-800 hover:bg-gray-200' : 'text-gray-500 hover:text-gray-300 hover:bg-gray-800'} transition-colors"
        onclick={loadStatus}
        title="Refresh (⌘R)"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
      </button>
    </div>
  </div>

  <!-- main layout -->
  <div class="flex flex-1 overflow-hidden" class:select-none={resizing} class:cursor-col-resize={resizing}>
    <!-- left sidebar -->
    <div class="shrink-0 flex flex-col {appState.theme === 'light' ? 'bg-gray-50' : 'bg-gray-900/40'}" style="width: {sidebarWidth}px">
      <!-- mode toggle -->
      <div class="flex shrink-0 border-b {appState.theme === 'light' ? 'border-gray-200' : 'border-gray-800'}">
        <button
          class="flex-1 py-1.5 text-[11px] font-medium transition-colors
            {sidebarMode === 'changes'
              ? (appState.theme === 'light' ? 'text-gray-800 border-b-2 border-blue-500' : 'text-gray-200 border-b-2 border-blue-500')
              : 'text-gray-500 hover:text-gray-400'}"
          onclick={() => sidebarMode = 'changes'}
        >Changes {appState.statusFiles.length > 0 ? `(${appState.statusFiles.length})` : ''}</button>
        <button
          class="flex-1 py-1.5 text-[11px] font-medium transition-colors
            {sidebarMode === 'files'
              ? (appState.theme === 'light' ? 'text-gray-800 border-b-2 border-blue-500' : 'text-gray-200 border-b-2 border-blue-500')
              : 'text-gray-500 hover:text-gray-400'}"
          onclick={() => sidebarMode = 'files'}
        >Files</button>
      </div>
      <div class="flex-1 overflow-y-auto min-h-0">
        {#if sidebarMode === 'changes'}
          <DiffTree files={appState.statusFiles} onSelect={selectFile} onRefresh={loadStatus} />
        {:else}
          <ProjectTree statusFiles={appState.statusFiles} {ignoredPaths} onOpen={openFile} {focusFolder} onFocusDone={() => focusFolder = null} />
        {/if}
      </div>
      <!-- commit log resize handle -->
      <div
        class="h-[4px] shrink-0 cursor-row-resize relative group border-t {appState.theme === 'light' ? 'border-gray-200' : 'border-gray-800'}"
        onmousedown={startCommitResize}
        role="separator"
        aria-orientation="horizontal"
        aria-label="Resize commit log"
      >
        <div class="absolute inset-x-0 top-[1px] h-px {appState.theme === 'light' ? 'group-hover:bg-blue-400' : 'group-hover:bg-blue-500'} transition-colors"></div>
      </div>
      <div class="shrink-0 overflow-y-auto" style="height: {commitLogHeight}px">
        <CommitLog onSelect={selectCommit} />
      </div>
    </div>

    <!-- resize handle -->
    <div
      class="w-[5px] shrink-0 cursor-col-resize relative group"
      onmousedown={startResize}
      role="separator"
      aria-orientation="vertical"
      aria-label="Resize sidebar"
    >
      <div class="absolute inset-y-0 left-[2px] w-px {appState.theme === 'light' ? 'bg-gray-300' : 'bg-gray-700'} group-hover:bg-blue-500 transition-colors {resizing ? 'bg-blue-500' : ''}"></div>
    </div>

    <!-- right: diff or editor -->
    <div class="flex-1 overflow-hidden flex flex-col">
      {#if revertConfirm && appState.selectedFile}
        <div class="shrink-0 flex items-center gap-3 px-4 py-2 bg-red-900/30 border-b border-red-800 text-[12px]">
          <span class="text-red-300">Revert changes in <strong>{appState.selectedFile.path}</strong>?</span>
          <button class="px-2 py-0.5 rounded bg-red-700 hover:bg-red-600 text-white text-[11px] transition-colors" onclick={doRevert}>Yes (↵)</button>
          <button class="px-2 py-0.5 rounded text-red-400 hover:text-red-300 text-[11px]" onclick={() => revertConfirm = false}>Cancel (Esc)</button>
        </div>
      {/if}
      {#if appState.selectedFile && isMedia(appState.selectedFile.path)}
        <MediaViewer path={appState.selectedFile.path} />
      {:else if appState.viewMode === 'edit' && appState.selectedFile}
        <Editor path={appState.selectedFile.path} theme={appState.theme} />
      {:else if selectedCommit}
        <div class="shrink-0 flex items-center gap-2 px-4 py-2 border-b text-[12px]
          {appState.theme === 'light' ? 'border-gray-200 text-gray-600' : 'border-gray-800 text-gray-400'}">
          <span class="font-mono text-[11px] {appState.theme === 'light' ? 'text-blue-600' : 'text-blue-400'}">{selectedCommit.short}</span>
          <span class="font-medium {appState.theme === 'light' ? 'text-gray-800' : 'text-gray-200'} truncate">{selectedCommit.message}</span>
          <span class="ml-auto shrink-0">{selectedCommit.author} · {selectedCommit.date}</span>
          <button
            class="shrink-0 ml-1 text-[11px] {appState.theme === 'light' ? 'text-gray-400 hover:text-gray-700' : 'text-gray-600 hover:text-gray-300'} transition-colors"
            onclick={() => { selectedCommit = null; commitDiffs = [] }}
          >✕</button>
        </div>
        <div class="flex-1 overflow-hidden">
          <CommitDiffViewer files={commitDiffs} loading={commitLoading} />
        </div>
      {:else}
        <DiffViewer diff={appState.currentDiff} loading={appState.loading} />
      {/if}
    </div>
  </div>

  <!-- bottom bar -->
  <div class="shrink-0 flex items-center gap-4 px-3 py-1 border-t text-[10px] select-none
    {appState.theme === 'light' ? 'border-gray-200 bg-gray-50 text-gray-400' : 'border-gray-800 bg-gray-900/60 text-gray-600'}">
    {#snippet key(k: string, label: string)}
      <span class="flex items-center gap-1">
        <kbd class="px-1 py-px rounded border text-[9px]
          {appState.theme === 'light' ? 'border-gray-300 bg-white text-gray-500' : 'border-gray-700 bg-gray-800 text-gray-400'}">{k}</kbd>
        <span>{label}</span>
      </span>
    {/snippet}
    {#if branch}
      <span class="flex items-center gap-1 {appState.theme === 'light' ? 'text-blue-500' : 'text-blue-400'}">
        <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 3v12m0 0a3 3 0 100 6 3 3 0 000-6zm0 0h6m0-12a3 3 0 100 6 3 3 0 000-6zm0 0v6" />
        </svg>
        {branch}
      </span>
      <span class="{appState.theme === 'light' ? 'text-gray-300' : 'text-gray-700'}">|</span>
    {/if}
    {@render key('⌘P', 'search files')}
    {@render key('⌘F', 'search content')}
    {@render key('↑↓', 'navigate')}
    {@render key('e', 'edit')}
    {@render key('d', 'revert')}
    {@render key('Esc', 'back')}
    {@render key('⌘R', 'refresh')}
    {@render key('⌘⇧T', 'theme')}
  </div>
</div>

{#if appState.searchOpen}
  <SearchOverlay
    onClose={() => appState.searchOpen = false}
    onSelect={handleSearchSelect}
  />
{/if}
