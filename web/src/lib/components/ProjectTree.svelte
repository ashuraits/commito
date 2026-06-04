<script lang="ts">
  import type { StatusFile } from '../api'
  import { api } from '../api'
  import { appState } from '../state.svelte'
  import FileTree from './FileTree.svelte'
  import type { TreeNode } from './FileTree.svelte'

  interface Props {
    statusFiles: StatusFile[]
    ignoredPaths: string[]
    onOpen: (path: string) => void
    focusFolder?: string | null
    onFocusDone?: () => void
  }

  let { statusFiles, ignoredPaths, onOpen, focusFolder, onFocusDone }: Props = $props()

  const allFiles = $derived(appState.allFiles)
  const loading = $derived(allFiles.length === 0)

  const statusMap = $derived.by(() => {
    const m = new Map<string, string>()
    for (const f of statusFiles) {
      if (!m.has(f.path) || f.staged) m.set(f.path, f.status)
    }
    return m
  })

  function addPath(root: TreeNode, filePath: string, status: string | undefined, isIgnoredDir = false) {
    const parts = filePath.split('/')
    let node = root
    for (let i = 0; i < parts.length; i++) {
      const part = parts[i]
      const isLast = i === parts.length - 1
      const currentPath = parts.slice(0, i + 1).join('/')
      if (!node.children.has(part)) {
        node.children.set(part, {
          name: part,
          path: currentPath,
          isDir: !isLast,
          ignoredDir: isLast && isIgnoredDir,
          children: new Map(),
        })
      }
      const child = node.children.get(part)!
      if (isLast && status !== undefined) child.status = status
      node = child
    }
  }

  const tree = $derived.by((): TreeNode => {
    const map = statusMap
    const root: TreeNode = { name: '', path: '', isDir: true, children: new Map() }
    const trackedSet = new Set(allFiles)

    for (const filePath of allFiles) {
      addPath(root, filePath, map.get(filePath))
    }

    for (const f of statusFiles) {
      if (f.status === 'untracked' && !trackedSet.has(f.path)) {
        addPath(root, f.path, 'untracked')
      }
    }

    for (const p of ignoredPaths) {
      const clean = p.replace(/\/$/, '')
      const isDir = p.endsWith('/')
      addPath(root, clean, 'ignored', isDir)
    }

    return root
  })

  async function loadDir(path: string): Promise<TreeNode[]> {
    const entries = await api.listDir(path)
    return entries.map(e => ({
      name: e.path.split('/').pop()!,
      path: e.path,
      isDir: e.isDir,
      ignoredDir: false,
      children: new Map(),
      status: 'ignored',
    }))
  }

  // auto-expand folders containing changed files
  const autoExpanded = $derived.by(() => {
    const set = new Set<string>()
    for (const f of statusFiles) {
      const parts = f.path.split('/')
      for (let i = 1; i < parts.length; i++) set.add(parts.slice(0, i).join('/'))
    }
    return set
  })

  const selectedPath = $derived(appState.selectedFile?.path ?? null)
</script>

<div class="py-1 text-[12px]">
  {#if loading}
    <div class="px-3 py-4 text-gray-600 text-[11px]">Loading…</div>
  {:else}
    <FileTree
      root={tree}
      {selectedPath}
      {autoExpanded}
      focusPath={focusFolder}
      {onFocusDone}
      onFile={(n) => onOpen(n.path)}
      {loadDir}
    />
  {/if}
</div>
