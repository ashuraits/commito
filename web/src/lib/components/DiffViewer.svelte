<script lang="ts">
  import { untrack } from 'svelte'
  import type { DiffFile } from '../api'
  import { appState } from '../state.svelte'
  import { api } from '../api'
  import HunkView from './HunkView.svelte'

  interface Props {
    diff: DiffFile | null
    loading: boolean
  }

  let { diff, loading }: Props = $props()

  let expandedAbove = $state(new Map<number, number>())
  let expandedBelow = $state(new Map<number, number>())
  let fileLines = $state<string[] | null>(null)

  $effect(() => {
    diff?.path
    untrack(() => {
      expandedAbove = new Map()
      expandedBelow = new Map()
      fileLines = null
    })
  })

  const canExpand = $derived(
    !!diff &&
    diff.status !== 'added' &&
    diff.status !== 'deleted' &&
    diff.status !== 'untracked'
  )

  async function loadFile() {
    if (fileLines !== null || !appState.selectedFile || !canExpand) return
    const content = await api.readFile(appState.selectedFile.path)
    const lines = content.split('\n')
    if (lines.at(-1) === '') lines.pop()
    fileLines = lines
  }

  // hidden lines in gap before hunk[i]
  function gapBefore(i: number): number {
    if (!diff) return 0
    const h = diff.hunks
    if (i === 0) return h[0].newStart - 1
    const prev = h[i - 1]
    return h[i].newStart - (prev.newStart + prev.newCount)
  }

  // hidden lines in gap after hunk[i]; -1 = unknown (last hunk, file not loaded)
  function gapAfter(i: number): number {
    if (!diff) return 0
    const h = diff.hunks
    if (i < h.length - 1) return gapBefore(i + 1)
    if (!fileLines) return -1
    const last = h[i]
    return Math.max(0, fileLines.length - (last.newStart + last.newCount - 1))
  }

  function showExpandAbove(i: number): boolean {
    if (!canExpand) return false
    const gap = gapBefore(i)
    if (gap <= 0) return false
    const ea = expandedAbove.get(i) ?? 0
    const eb = i > 0 ? (expandedBelow.get(i - 1) ?? 0) : 0
    return ea + eb < gap
  }

  function showExpandBelow(i: number): boolean {
    if (!canExpand) return false
    const gap = gapAfter(i)
    if (gap === 0) return false
    if (gap === -1) return true
    const eb = expandedBelow.get(i) ?? 0
    const ea = i < (diff?.hunks?.length ?? 0) - 1 ? (expandedAbove.get(i + 1) ?? 0) : 0
    return eb + ea < gap
  }

  const STEP = 10

  async function doExpandAbove(i: number) {
    await loadFile()
    const ea = expandedAbove.get(i) ?? 0
    const eb = i > 0 ? (expandedBelow.get(i - 1) ?? 0) : 0
    const add = Math.min(STEP, gapBefore(i) - ea - eb)
    if (add > 0) expandedAbove = new Map(expandedAbove).set(i, ea + add)
  }

  async function doExpandBelow(i: number) {
    await loadFile()
    const eb = expandedBelow.get(i) ?? 0
    const ea = i < (diff?.hunks?.length ?? 0) - 1 ? (expandedAbove.get(i + 1) ?? 0) : 0
    const gap = gapAfter(i)
    const add = Math.min(STEP, gap === -1 ? STEP : gap - eb - ea)
    if (add > 0) expandedBelow = new Map(expandedBelow).set(i, eb + add)
  }

  function getFileLines(start: number, count: number): { lineNo: number; content: string }[] {
    if (!fileLines || count <= 0) return []
    return Array.from({ length: count }, (_, j) => {
      const n = start + j
      return { lineNo: n, content: fileLines[n - 1] ?? '' }
    })
  }

  // lines grown upward from hunk[i] start
  function linesAbove(i: number): { lineNo: number; content: string }[] {
    if (!diff || !fileLines) return []
    const ea = expandedAbove.get(i) ?? 0
    if (ea === 0) return []
    return getFileLines(diff.hunks[i].newStart - ea, ea)
  }

  // lines grown downward from hunk[i] end
  function linesBelow(i: number): { lineNo: number; content: string }[] {
    if (!diff || !fileLines) return []
    const eb = expandedBelow.get(i) ?? 0
    if (eb === 0) return []
    const h = diff.hunks[i]
    return getFileLines(h.newStart + h.newCount, eb)
  }
</script>

<div class="h-full flex flex-col overflow-hidden">
  {#if loading}
    <div class="flex-1 flex items-center justify-center text-gray-500 text-sm">Loading…</div>

  {:else if !diff}
    <div class="flex-1 flex flex-col items-center justify-center gap-3 text-gray-600 select-none">
      <svg class="w-10 h-10 opacity-30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
          d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
      <span class="text-sm">Select a file to view diff</span>
      <span class="text-xs text-gray-600">Use <kbd class="px-1.5 py-0.5 rounded border border-gray-700 text-[11px]">⌘P</kbd> to search files</span>
    </div>

  {:else if !diff.hunks?.length}
    <div class="flex-1 flex items-center justify-center text-gray-500 text-sm">
      No changes in this file
    </div>

  {:else}
    <!-- file header -->
    <div class="flex items-center gap-3 px-4 py-2 border-b border-gray-800 shrink-0 bg-gray-900/50 text-[12px]">
      <span class="text-gray-300 font-medium">{diff.path}</span>
      <span class="ml-auto text-[11px] px-1.5 py-px rounded
        {diff.status === 'added' ? 'bg-green-900/40 text-green-400' :
         diff.status === 'deleted' ? 'bg-red-900/40 text-red-400' :
         diff.status === 'renamed' ? 'bg-yellow-900/40 text-yellow-400' :
         'bg-gray-800 text-gray-500'}">
        {diff.status}
      </span>
      <span class="text-green-400">+{diff.addCount}</span>
      <span class="text-red-400">-{diff.delCount}</span>
      <button
        class="ml-1 px-2 py-0.5 rounded border border-gray-700 text-gray-400 hover:text-gray-200 hover:border-gray-500 transition-colors text-[11px]"
        onclick={() => appState.viewMode = 'edit'}
        title="Edit file (e)"
      >edit</button>
    </div>

    <!-- hunks -->
    <div class="flex-1 overflow-auto font-mono text-[12px]">
      {#each (diff.hunks ?? []) as hunk, i}

        {#if showExpandAbove(i)}
          <button
            class="w-full py-0.5 text-[11px] text-blue-400/50 hover:text-blue-300 hover:bg-blue-500/5 transition-colors flex items-center justify-center gap-1 border-y border-gray-800/40 select-none"
            onclick={() => doExpandAbove(i)}
          >↑ show more above</button>
        {/if}

        {#each linesAbove(i) as { lineNo, content }}
          <div class="flex leading-5 diff-context">
            <span class="line-num shrink-0 select-none text-[11px] leading-5 pl-2 pr-1 w-10 text-right">{lineNo}</span>
            <span class="line-num shrink-0 select-none text-[11px] leading-5 pl-1 pr-2 w-10 text-right">{lineNo}</span>
            <span class="px-1 text-xs leading-5 shrink-0 w-4 select-none"> </span>
            <span class="px-2 whitespace-pre overflow-hidden">{content}</span>
          </div>
        {/each}

        <HunkView {hunk} />

        {#each linesBelow(i) as { lineNo, content }}
          <div class="flex leading-5 diff-context">
            <span class="line-num shrink-0 select-none text-[11px] leading-5 pl-2 pr-1 w-10 text-right">{lineNo}</span>
            <span class="line-num shrink-0 select-none text-[11px] leading-5 pl-1 pr-2 w-10 text-right">{lineNo}</span>
            <span class="px-1 text-xs leading-5 shrink-0 w-4 select-none"> </span>
            <span class="px-2 whitespace-pre overflow-hidden">{content}</span>
          </div>
        {/each}

        {#if showExpandBelow(i)}
          <button
            class="w-full py-0.5 text-[11px] text-blue-400/50 hover:text-blue-300 hover:bg-blue-500/5 transition-colors flex items-center justify-center gap-1 border-y border-gray-800/40 select-none"
            onclick={() => doExpandBelow(i)}
          >↓ show more below</button>
        {/if}

      {/each}
    </div>
  {/if}
</div>
