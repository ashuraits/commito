<script lang="ts">
  import { api } from '../api'
  import type { StatusFile } from '../api'
  import { appState } from '../state.svelte'
  import FileTree from './FileTree.svelte'
  import type { TreeNode } from './FileTree.svelte'

  interface Props {
    files: StatusFile[]
    onSelect: (file: StatusFile) => void
    onRefresh: () => void
  }

  let { files, onSelect, onRefresh }: Props = $props()

  const staged   = $derived((files ?? []).filter(f => f.staged))
  const unstaged = $derived((files ?? []).filter(f => !f.staged))

  function buildTree(items: StatusFile[]): TreeNode {
    const root: TreeNode = { name: '', path: '', isDir: true, children: new Map() }
    for (const f of items) {
      const parts = f.path.split('/')
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
        if (isLast) { child.status = f.status; child.data = f }
        node = child
      }
    }
    return root
  }

  const stagedTree   = $derived(buildTree(staged))
  const unstagedTree = $derived(buildTree(unstaged))

  function allDirs(items: StatusFile[]): Set<string> {
    const s = new Set<string>()
    for (const f of items) {
      const parts = f.path.split('/')
      for (let i = 1; i < parts.length; i++) s.add(parts.slice(0, i).join('/'))
    }
    return s
  }

  const stagedDirs   = $derived(allDirs(staged))
  const unstagedDirs = $derived(allDirs(unstaged))

  let reverting = $state<string | null>(null)

  async function act(fn: () => Promise<void>) {
    await fn()
    onRefresh()
  }

  const selectedPath = $derived(appState.selectedFile?.path ?? null)

  function fileOf(node: TreeNode): StatusFile {
    return node.data as StatusFile
  }
</script>

{#snippet stagedActions(node: TreeNode)}
  <button
    class="px-1.5 py-px rounded text-[10px] opacity-0 group-hover:opacity-100 transition-opacity text-gray-500 hover:text-gray-300 hover:bg-white/5"
    title="Unstage"
    onclick={(e) => { e.stopPropagation(); act(() => api.unstage(node.path)) }}
  >−</button>
{/snippet}

{#snippet unstagedActions(node: TreeNode)}
  {@const f = fileOf(node)}
  {#if reverting === node.path}
    <span class="text-[10px] text-red-300 pr-1 opacity-100">revert?</span>
    <button class="text-[10px] text-red-400 hover:text-red-300 px-1" onclick={(e) => { e.stopPropagation(); reverting = null; act(() => api.revert(node.path)) }}>yes</button>
    <button class="text-[10px] text-gray-500 hover:text-gray-300 px-1" onclick={(e) => { e.stopPropagation(); reverting = null }}>no</button>
  {:else}
    <button
      class="px-1.5 py-px rounded text-[10px] opacity-0 group-hover:opacity-100 transition-opacity text-gray-500 hover:text-gray-300 hover:bg-white/5"
      title="Stage"
      onclick={(e) => { e.stopPropagation(); act(() => api.stage(node.path)) }}
    >+</button>
    {#if f?.status !== 'untracked'}
      <button
        class="px-1.5 py-px rounded text-[10px] opacity-0 group-hover:opacity-100 transition-opacity text-red-400 hover:text-red-300 hover:bg-red-500/10"
        title="Revert"
        onclick={(e) => { e.stopPropagation(); reverting = node.path }}
      >↩</button>
    {/if}
  {/if}
{/snippet}

<div class="py-1 text-[12px]">
  {#if staged.length > 0}
    <div class="flex items-center px-2 pt-2 pb-1 select-none">
      <span class="text-[10px] font-semibold uppercase tracking-wider text-gray-600 flex-1">Staged ({staged.length})</span>
      <button class="text-[10px] text-gray-600 hover:text-gray-400 px-1" onclick={() => act(() => api.unstageAll())}>unstage all</button>
    </div>
    <FileTree root={stagedTree} {selectedPath} autoExpanded={stagedDirs} onFile={(n) => onSelect(fileOf(n))} fileActions={stagedActions} />
  {/if}

  {#if unstaged.length > 0}
    <div class="flex items-center px-2 pt-3 pb-1 select-none">
      <span class="text-[10px] font-semibold uppercase tracking-wider text-gray-600 flex-1">Changes ({unstaged.length})</span>
      <button class="text-[10px] text-gray-600 hover:text-gray-400 px-1" onclick={() => act(() => api.stageAll())}>+all</button>
      <button class="text-[10px] text-red-700 hover:text-red-500 px-1" onclick={() => act(() => api.revertAll())}>↩all</button>
    </div>
    <FileTree root={unstagedTree} {selectedPath} autoExpanded={unstagedDirs} onFile={(n) => onSelect(fileOf(n))} fileActions={unstagedActions} />
  {/if}

  {#if files.length === 0}
    <div class="px-3 py-4 text-gray-600 text-[11px]">No changes</div>
  {/if}
</div>
