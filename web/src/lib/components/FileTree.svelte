<script lang="ts">
  import { untrack } from 'svelte'
  import type { Snippet } from 'svelte'
  import { appState } from '../state.svelte'
  import ContextMenu from './ContextMenu.svelte'

  export interface FileOps {
    clipboard: { path: string; op: 'cut' | 'copy' } | null
    onCut: (path: string) => void
    onCopy: (path: string) => void
    onPaste: (targetDir: string) => void
    onDelete: (path: string) => void
    onRename: (oldPath: string, newName: string) => void
  }

  export interface TreeNode {
    name: string
    path: string
    isDir: boolean
    ignoredDir?: boolean
    children: Map<string, TreeNode>
    status?: string
    data?: unknown
  }

  interface Props {
    root: TreeNode
    selectedPath?: string | null
    focusPath?: string | null
    onFocusDone?: () => void
    onFile: (node: TreeNode) => void
    fileActions?: Snippet<[TreeNode]>
    autoExpanded?: Set<string>
    loadDir?: (path: string) => Promise<TreeNode[]>
    onToggle?: (path: string, open: boolean) => void
    fileOps?: FileOps
  }

  let {
    root,
    selectedPath = null,
    focusPath = null,
    onFocusDone,
    onFile,
    fileActions,
    autoExpanded = new Set(),
    loadDir,
    onToggle,
    fileOps,
  }: Props = $props()

  let lazyChildren = $state(new Map<string, TreeNode[]>())
  let lazyLoading = $state(new Set<string>())

  async function toggleIgnoredDir(path: string) {
    if (userToggled.get(path)) {
      userToggled.set(path, false)
      userToggled = new Map(userToggled)
      lazyChildren.delete(path)
      lazyChildren = new Map(lazyChildren)
      onToggle?.(path, false)
      return
    }
    if (!lazyChildren.has(path) && loadDir) {
      lazyLoading = new Set(lazyLoading).add(path)
      const children = await loadDir(path)
      lazyChildren = new Map(lazyChildren).set(path, children)
      lazyLoading.delete(path)
      lazyLoading = new Set(lazyLoading)
    }
    userToggled.set(path, true)
    userToggled = new Map(userToggled)
    onToggle?.(path, true)
  }

  let container: HTMLDivElement
  let userToggled = $state(new Map<string, boolean>())
  let renamingPath = $state<string | null>(null)
  let renameValue = $state('')
  let contextMenu = $state<{ x: number; y: number; node: TreeNode } | null>(null)

  function openContextMenu(e: MouseEvent, node: TreeNode) {
    if (!fileOps) return
    e.preventDefault()
    e.stopPropagation()
    contextMenu = { x: e.clientX, y: e.clientY, node }
  }

  function startRename(node: TreeNode) {
    renamingPath = node.path
    renameValue = node.name
  }

  function commitRename(node: TreeNode) {
    const name = renameValue.trim()
    if (name && name !== node.name) {
      fileOps?.onRename(node.path, name)
    }
    renamingPath = null
  }

  function contextMenuItems(node: TreeNode) {
    const ops = fileOps!
    const isDir = node.isDir || node.ignoredDir
    const parentDir = node.path.includes('/') ? node.path.split('/').slice(0, -1).join('/') : ''
    const pasteDir = isDir ? node.path : parentDir
    return [
      { label: 'Rename', shortcut: 'F2', action: () => startRename(node) },
      { separator: true as const },
      ...(!isDir ? [
        { label: 'Cut', shortcut: '⌘X', action: () => ops.onCut(node.path) },
        { label: 'Copy', shortcut: '⌘C', action: () => ops.onCopy(node.path) },
      ] : []),
      { label: 'Paste', shortcut: '⌘V', action: () => ops.onPaste(pasteDir), disabled: !ops.clipboard },
      { separator: true as const },
      { label: 'Delete', shortcut: '⌫', action: () => ops.onDelete(node.path), danger: true },
    ]
  }

  function isOpen(path: string): boolean {
    if (userToggled.has(path)) return userToggled.get(path)!
    return autoExpanded.has(path)
  }

  function toggle(path: string) {
    const open = !isOpen(path)
    userToggled.set(path, open)
    userToggled = new Map(userToggled)
    onToggle?.(path, open)
  }

  $effect(() => {
    if (!focusPath) return
    const target = focusPath
    const parts = target.split('/')
    untrack(() => {
      const next = new Map(userToggled)
      for (let i = 1; i <= parts.length; i++) next.set(parts.slice(0, i).join('/'), true)
      userToggled = next
    })
    setTimeout(() => {
      const el = container?.querySelector(`[data-path="${CSS.escape(target)}"]`)
      el?.scrollIntoView({ block: 'center', behavior: 'smooth' })
      onFocusDone?.()
    }, 100)
  })

  $effect(() => {
    const sel = selectedPath
    if (!sel) return
    const parts = sel.split('/')
    untrack(() => {
      const next = new Map(userToggled)
      for (let i = 1; i < parts.length; i++) next.set(parts.slice(0, i).join('/'), true)
      userToggled = next
    })
    setTimeout(() => {
      const el = container?.querySelector(`[data-path="${CSS.escape(sel)}"]`)
      el?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
    }, 100)
  })

  const light = $derived(appState.theme === 'light')
  const textFile = $derived(light ? 'text-gray-800' : 'text-gray-300')
  const textDir  = $derived(light ? 'text-gray-700' : 'text-gray-200')
  const hoverBg  = $derived(light ? 'hover:bg-black/5' : 'hover:bg-white/5')

  export function statusColor(s: string | undefined): string {
    switch (s) {
      case 'added':     return 'text-emerald-300'
      case 'untracked': return 'text-gray-500'
      case 'ignored':   return 'text-gray-600'
      case 'deleted':   return 'text-red-300'
      case 'renamed':   return 'text-yellow-300'
      case 'modified':  return 'text-blue-300'
      default:          return ''
    }
  }

  export function statusIcon(s: string | undefined): string {
    switch (s) {
      case 'added':     return 'A'
      case 'untracked': return '?'
      case 'ignored':   return '!'
      case 'deleted':   return 'D'
      case 'renamed':   return 'R'
      case 'modified':  return 'M'
      default:          return ''
    }
  }

  function dirStatus(node: TreeNode): string | undefined {
    if (node.status && node.status !== 'ignored') return node.status
    for (const child of node.children.values()) {
      const s = dirStatus(child)
      if (s) return s
    }
  }

  function sortedChildren(node: TreeNode): TreeNode[] {
    return [...node.children.values()].sort((a, b) => {
      const aDir = (a.isDir || a.ignoredDir) ? 0 : 1
      const bDir = (b.isDir || b.ignoredDir) ? 0 : 1
      return aDir - bDir || a.name.localeCompare(b.name)
    })
  }

  function findNode(node: TreeNode, path: string): TreeNode | null {
    if (node.path === path) return node
    for (const child of node.children.values()) {
      const found = findNode(child, path)
      if (found) return found
    }
    return null
  }

  function focusInput(node: HTMLInputElement) {
    node.focus()
    node.select()
  }
</script>

{#snippet nodeTree(node: TreeNode, depth: number)}
  {#each sortedChildren(node) as child}
    {#if child.ignoredDir}
      {@const open = userToggled.get(child.path) ?? false}
      {@const loading = lazyLoading.has(child.path)}
      <button
        class="w-full text-left flex items-center gap-1.5 py-[3px] pr-2 {hoverBg} transition-colors"
        style="padding-left: {10 + depth * 12}px"
        data-path={child.path}
        onclick={() => toggleIgnoredDir(child.path)}
        oncontextmenu={(e) => openContextMenu(e, child)}
      >
        {#if loading}
          <span class="w-3 h-3 shrink-0 text-[10px]">…</span>
        {:else}
          <svg class="w-3 h-3 shrink-0 transition-transform {open ? 'rotate-90' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        {/if}
        <svg class="w-3.5 h-3.5 shrink-0 text-yellow-700" fill="currentColor" viewBox="0 0 20 20">
          <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
        </svg>
        {#if renamingPath === child.path}
          <input
            class="flex-1 min-w-0 bg-transparent border border-blue-500 rounded px-1 text-[12px] outline-none"
            bind:value={renameValue}
            onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); commitRename(child) } else if (e.key === 'Escape') { renamingPath = null } }}
            onblur={() => commitRename(child)}
            use:focusInput
          />
        {:else}
          <span class="truncate text-[12px] flex-1 text-gray-500">{child.name}</span>
        {/if}
      </button>
      {#if open && lazyChildren.has(child.path)}
        {@render lazyTree(lazyChildren.get(child.path)!, depth + 1)}
      {/if}
    {:else if child.isDir}
      {@const open = isOpen(child.path)}
      {@const changed = dirStatus(child)}
      <button
        class="w-full text-left flex items-center gap-1.5 py-[3px] pr-2 {hoverBg} transition-colors"
        style="padding-left: {10 + depth * 12}px"
        data-path={child.path}
        onclick={() => toggle(child.path)}
        oncontextmenu={(e) => openContextMenu(e, child)}
      >
        <svg class="w-3 h-3 shrink-0 transition-transform {open ? 'rotate-90' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        <svg class="w-3.5 h-3.5 shrink-0 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
          <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
        </svg>
        {#if renamingPath === child.path}
          <input
            class="flex-1 min-w-0 bg-transparent border border-blue-500 rounded px-1 text-[12px] outline-none"
            bind:value={renameValue}
            onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); commitRename(child) } else if (e.key === 'Escape') { renamingPath = null } }}
            onblur={() => commitRename(child)}
            use:focusInput
          />
        {:else}
          <span class="truncate text-[12px] flex-1 {changed ? statusColor(changed) : textDir}">{child.name}</span>
        {/if}
      </button>
      {#if open}
        {@render nodeTree(child, depth + 1)}
      {/if}
    {:else}
      {@const isSelected = selectedPath === child.path}
      {@const isCut = fileOps?.clipboard?.op === 'cut' && fileOps.clipboard.path === child.path}
      <div
        class="group relative flex items-center overflow-hidden {isSelected ? 'bg-blue-500/20' : hoverBg} transition-colors {isCut ? 'opacity-40' : ''}"
        oncontextmenu={(e) => openContextMenu(e, child)}
      >
        <button
          class="flex items-center gap-1.5 py-[3px] flex-1 min-w-0 text-left {textFile}"
          style="padding-left: {10 + depth * 12}px; padding-right: {fileActions ? '24px' : '8px'}"
          data-path={child.path}
          onclick={() => onFile(child)}
        >
          <svg class="w-3.5 h-3.5 shrink-0 {child.status ? statusColor(child.status) : 'text-gray-600'}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          {#if renamingPath === child.path}
            <input
              class="flex-1 min-w-0 bg-transparent border border-blue-500 rounded px-1 text-[12px] outline-none"
              bind:value={renameValue}
              onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); commitRename(child) } else if (e.key === 'Escape') { renamingPath = null } }}
              onblur={() => commitRename(child)}
              use:focusInput
            />
          {:else}
            <span class="truncate text-[12px] flex-1 {statusColor(child.status)}">{child.name}</span>
          {/if}
        </button>
        {#if child.status && renamingPath !== child.path}
          <span class="absolute right-1.5 text-[10px] font-bold w-4 text-center pointer-events-none
            {fileActions ? 'group-hover:opacity-0' : ''} transition-opacity {statusColor(child.status)}">
            {statusIcon(child.status)}
          </span>
        {/if}
        {#if fileActions}
          <div class="absolute right-0 flex items-center pr-1 gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
            {@render fileActions(child)}
          </div>
        {/if}
      </div>
    {/if}
  {/each}
{/snippet}

{#snippet lazyTree(nodes: TreeNode[], depth: number)}
  {#each nodes.sort((a, b) => (a.isDir ? 0 : 1) - (b.isDir ? 0 : 1) || a.name.localeCompare(b.name)) as child}
    {#if child.isDir}
      {@const open = userToggled.get(child.path) ?? false}
      <button
        class="w-full text-left flex items-center gap-1.5 py-[3px] pr-2 hover:bg-white/5 transition-colors"
        style="padding-left: {10 + depth * 12}px"
        data-path={child.path}
        onclick={() => { userToggled.set(child.path, !open); userToggled = new Map(userToggled) }}
        oncontextmenu={(e) => openContextMenu(e, child)}
      >
        <svg class="w-3 h-3 shrink-0 transition-transform {open ? 'rotate-90' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        <svg class="w-3.5 h-3.5 shrink-0 text-yellow-700" fill="currentColor" viewBox="0 0 20 20">
          <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
        </svg>
        <span class="truncate text-[12px] flex-1 text-gray-500">{child.name}</span>
      </button>
      {#if open}
        {#await (loadDir ? loadDir(child.path) : Promise.resolve([])) then kids}
          {@render lazyTree(kids, depth + 1)}
        {/await}
      {/if}
    {:else}
      {@const isSelected = selectedPath === child.path}
      {@const isCut = fileOps?.clipboard?.op === 'cut' && fileOps.clipboard.path === child.path}
      <div
        class="group relative flex items-center overflow-hidden {isSelected ? 'bg-blue-500/20' : 'hover:bg-white/5'} transition-colors {isCut ? 'opacity-40' : ''}"
        oncontextmenu={(e) => openContextMenu(e, child)}
      >
        <button
          class="w-full flex items-center gap-1.5 py-[3px] text-left text-gray-600 transition-colors"
          style="padding-left: {10 + depth * 12}px; padding-right: 8px"
          data-path={child.path}
          onclick={() => onFile(child)}
        >
          <svg class="w-3.5 h-3.5 shrink-0 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          {#if renamingPath === child.path}
            <input
              class="flex-1 min-w-0 bg-transparent border border-blue-500 rounded px-1 text-[12px] outline-none"
              bind:value={renameValue}
              onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); commitRename(child) } else if (e.key === 'Escape') { renamingPath = null } }}
              onblur={() => commitRename(child)}
              use:focusInput
            />
          {:else}
            <span class="truncate text-[12px] flex-1">{child.name}</span>
          {/if}
        </button>
      </div>
    {/if}
  {/each}
{/snippet}

<div
  class="py-1 text-[12px]"
  bind:this={container}
  onkeydown={(e) => {
    if (!fileOps || renamingPath) return
    if (e.key === 'F2' && selectedPath) {
      const treeNode = findNode(root, selectedPath)
      if (treeNode) { e.preventDefault(); startRename(treeNode) }
    }
    if ((e.key === 'Delete' || e.key === 'Backspace') && selectedPath) {
      const treeNode = findNode(root, selectedPath)
      if (treeNode) { e.preventDefault(); fileOps.onDelete(treeNode.path) }
    }
  }}
>
  {@render nodeTree(root, 0)}
</div>

{#if contextMenu && fileOps}
  <ContextMenu
    x={contextMenu.x}
    y={contextMenu.y}
    items={contextMenuItems(contextMenu.node)}
    onClose={() => contextMenu = null}
  />
{/if}
