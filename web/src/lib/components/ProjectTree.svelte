<script lang="ts">
  import type { StatusFile } from '../api'
  import { appState } from '../state.svelte'
  import FileTree from './FileTree.svelte'
  import type { TreeNode } from './FileTree.svelte'

  interface Props {
    statusFiles: StatusFile[]
    onOpen: (path: string) => void
    focusFolder?: string | null
    onFocusDone?: () => void
  }

  let { statusFiles, onOpen, focusFolder, onFocusDone }: Props = $props()

  const allFiles = $derived(appState.allFiles)
  const loading = $derived(allFiles.length === 0)

  const statusMap = $derived.by(() => {
    const m = new Map<string, string>()
    for (const f of statusFiles) {
      if (!m.has(f.path) || f.staged) m.set(f.path, f.status)
    }
    return m
  })

  const tree = $derived.by((): TreeNode => {
    const map = statusMap
    const root: TreeNode = { name: '', path: '', isDir: true, children: new Map() }
    for (const filePath of allFiles) {
      const parts = filePath.split('/')
      let node = root
      for (let i = 0; i < parts.length; i++) {
        const part = parts[i]
        const isLast = i === parts.length - 1
        if (!node.children.has(part)) {
          node.children.set(part, {
            name: part,
            path: parts.slice(0, i + 1).join('/'),
            isDir: !isLast,
            children: new Map(),
          })
        }
        const child = node.children.get(part)!
        if (isLast) child.status = map.get(filePath)
        node = child
      }
    }
    return root
  })

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
    />
  {/if}
</div>
