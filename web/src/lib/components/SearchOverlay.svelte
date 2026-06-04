<script lang="ts">
  import { onMount } from 'svelte'
  import Fuse from 'fuse.js'
  import type { SearchResult } from '../api'
  import { api } from '../api'
  import { appState } from '../state.svelte'

  interface Props {
    onClose: () => void
    onSelect: (result: SearchResult) => void
  }

  let { onClose, onSelect }: Props = $props()

  let query = $state('')
  let results = $state<SearchResult[]>([])
  let activeIdx = $state(0)
  let input: HTMLInputElement
  let mode = $state<'files' | 'content'>(appState.searchMode)
  let debounce: ReturnType<typeof setTimeout>
  let allFiles = $state<SearchResult[]>([])
  let fuse: Fuse<SearchResult>

  onMount(async () => {
    input?.focus()
    const paths = await api.allFiles()
    const dirs = new Set<string>()
    for (const p of paths) {
      const parts = p.split('/')
      for (let i = 1; i < parts.length; i++) dirs.add(parts.slice(0, i).join('/'))
    }
    const dirEntries: SearchResult[] = [...dirs].sort().map(p => ({ path: p, type: 'folder' }))
    const fileEntries: SearchResult[] = paths.map(p => ({ path: p, type: 'file' }))
    allFiles = [...dirEntries, ...fileEntries]
    fuse = new Fuse(allFiles, { keys: ['path'], threshold: 0.4 })
  })

  $effect(() => {
    activeIdx = 0
    clearTimeout(debounce)
    if (!query) {
      results = allFiles.slice(0, 20)
      return
    }
    if (mode === 'files') {
      results = fuse ? fuse.search(query).map(r => r.item).slice(0, 20) : []
    } else {
      debounce = setTimeout(async () => {
        results = await api.searchContent(query)
      }, 300)
    }
  })

  function handleKey(e: KeyboardEvent) {
    if (e.key === 'Escape') { onClose(); return }
    if (e.key === 'ArrowDown') { e.preventDefault(); activeIdx = Math.min(activeIdx + 1, results.length - 1) }
    if (e.key === 'ArrowUp') { e.preventDefault(); activeIdx = Math.max(activeIdx - 1, 0) }
    if (e.key === 'Enter') { if (results[activeIdx]) onSelect(results[activeIdx]) }
    if (e.key === 'Tab') { e.preventDefault(); mode = mode === 'files' ? 'content' : 'files' }
  }
</script>

<!-- backdrop -->
<div
  class="fixed inset-0 bg-black/50 z-50 flex items-start justify-center pt-[15vh]"
  onclick={onClose}
  onkeydown={e => e.key === 'Escape' && onClose()}
  role="dialog"
  aria-modal="true"
  tabindex="-1"
>
  <div
    class="bg-gray-900 border border-gray-700 rounded-lg w-[600px] max-h-[60vh] flex flex-col shadow-2xl overflow-hidden"
    onclick={e => e.stopPropagation()}
    onkeydown={() => {}}
    role="none"
  >
    <!-- input row -->
    <div class="flex items-center gap-2 px-3 py-2 border-b border-gray-800">
      <svg class="w-4 h-4 text-gray-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <input
        bind:this={input}
        bind:value={query}
        onkeydown={handleKey}
        class="flex-1 bg-transparent outline-none text-[13px] text-gray-200 placeholder-gray-600"
        placeholder={mode === 'files' ? 'Search files…' : 'Search in content…'}
      />
      <div class="flex gap-1 shrink-0">
        <button
          class="px-2 py-0.5 rounded text-[11px] transition-colors {mode === 'files' ? 'bg-blue-600 text-white' : 'text-gray-500 hover:text-gray-300'}"
          onclick={() => { mode = 'files'; input?.focus() }}
        >files</button>
        <button
          class="px-2 py-0.5 rounded text-[11px] transition-colors {mode === 'content' ? 'bg-blue-600 text-white' : 'text-gray-500 hover:text-gray-300'}"
          onclick={() => { mode = 'content'; input?.focus() }}
        >content</button>
      </div>
    </div>

    <!-- results -->
    <div class="overflow-auto">
      {#each results as result, i}
        <button
          class="w-full text-left px-3 py-1.5 flex items-center gap-2 text-[12px] transition-colors {i === activeIdx ? 'bg-blue-600/30 text-white' : 'text-gray-300 hover:bg-gray-800'}"
          onclick={() => onSelect(result)}
          onmouseenter={() => activeIdx = i}
        >
          {#if result.type === 'folder'}
            <svg class="w-3.5 h-3.5 shrink-0 text-yellow-600/70" fill="currentColor" viewBox="0 0 20 20">
              <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
            </svg>
            <span class="truncate">{result.path}/</span>
          {:else if result.type === 'file'}
            <svg class="w-3.5 h-3.5 shrink-0 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <span class="truncate">{result.path}</span>
          {:else}
            <svg class="w-3.5 h-3.5 shrink-0 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h8" />
            </svg>
            <span class="text-gray-500 shrink-0">{result.path}:{result.line}</span>
            <span class="truncate text-gray-400">{result.content}</span>
          {/if}
        </button>
      {/each}

      {#if results.length === 0 && query}
        <div class="px-3 py-4 text-gray-600 text-[12px] text-center">No results</div>
      {/if}
    </div>

    <!-- footer hints -->
    <div class="px-3 py-1.5 border-t border-gray-800 flex gap-3 text-[11px] text-gray-600 select-none">
      <span><kbd class="text-gray-500">↑↓</kbd> navigate</span>
      <span><kbd class="text-gray-500">↵</kbd> open</span>
      <span><kbd class="text-gray-500">Tab</kbd> switch mode</span>
      <span><kbd class="text-gray-500">Esc</kbd> close</span>
    </div>
  </div>
</div>
