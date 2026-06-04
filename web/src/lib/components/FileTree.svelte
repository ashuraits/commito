<script lang="ts">
  import { untrack } from 'svelte'
  import type { Snippet } from 'svelte'
  import { appState } from '../state.svelte'

  export interface TreeNode {
    name: string
    path: string
    isDir: boolean
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
  }

  let {
    root,
    selectedPath = null,
    focusPath = null,
    onFocusDone,
    onFile,
    fileActions,
    autoExpanded = new Set(),
  }: Props = $props()

  let container: HTMLDivElement
  let userToggled = $state(new Map<string, boolean>())

  function isOpen(path: string): boolean {
    if (userToggled.has(path)) return userToggled.get(path)!
    return autoExpanded.has(path)
  }

  function toggle(path: string) {
    userToggled.set(path, !isOpen(path))
    userToggled = new Map(userToggled)
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
      case 'added':
      case 'untracked': return 'text-emerald-300'
      case 'deleted':   return 'text-red-300'
      case 'renamed':   return 'text-yellow-300'
      case 'modified':  return 'text-blue-300'
      default:          return ''
    }
  }

  export function statusIcon(s: string | undefined): string {
    switch (s) {
      case 'added':
      case 'untracked': return 'A'
      case 'deleted':   return 'D'
      case 'renamed':   return 'R'
      case 'modified':  return 'M'
      default:          return ''
    }
  }

  function dirStatus(node: TreeNode): string | undefined {
    if (node.status) return node.status
    for (const child of node.children.values()) {
      const s = dirStatus(child)
      if (s) return s
    }
  }

  function sortedChildren(node: TreeNode): TreeNode[] {
    return [...node.children.values()].sort((a, b) => {
      const aDir = a.isDir ? 0 : 1
      const bDir = b.isDir ? 0 : 1
      return aDir - bDir || a.name.localeCompare(b.name)
    })
  }
</script>

{#snippet nodeTree(node: TreeNode, depth: number)}
  {#each sortedChildren(node) as child}
    {#if child.isDir}
      {@const open = isOpen(child.path)}
      {@const changed = dirStatus(child)}
      <button
        class="w-full text-left flex items-center gap-1.5 py-[3px] pr-2 {hoverBg} transition-colors"
        style="padding-left: {10 + depth * 12}px"
        data-path={child.path}
        onclick={() => toggle(child.path)}
      >
        <svg class="w-3 h-3 shrink-0 transition-transform {open ? 'rotate-90' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        <svg class="w-3.5 h-3.5 shrink-0 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
          <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
        </svg>
        <span class="truncate text-[12px] flex-1 {changed ? statusColor(changed) : textDir}">{child.name}</span>
      </button>
      {#if open}
        {@render nodeTree(child, depth + 1)}
      {/if}
    {:else}
      {@const isSelected = selectedPath === child.path}
      <div class="group relative flex items-center overflow-hidden {isSelected ? 'bg-blue-500/20' : hoverBg} transition-colors">
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
          <span class="truncate text-[12px] flex-1 {statusColor(child.status)}">{child.name}</span>
        </button>
        {#if child.status}
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

<div class="py-1 text-[12px]" bind:this={container}>
  {@render nodeTree(root, 0)}
</div>
