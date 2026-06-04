<script lang="ts">
  import type { DiffFile } from '../api'
  import { appState } from '../state.svelte'
  import HunkView from './HunkView.svelte'

  let { files, loading }: { files: DiffFile[]; loading: boolean } = $props()

  const dark = $derived(appState.theme !== 'light')
</script>

<div class="h-full flex flex-col overflow-hidden">
  {#if loading}
    <div class="flex-1 flex items-center justify-center text-sm {dark ? 'text-gray-500' : 'text-gray-400'}">Loading…</div>

  {:else if files.length === 0}
    <div class="flex-1 flex items-center justify-center text-sm {dark ? 'text-gray-600' : 'text-gray-400'}">No changes</div>

  {:else}
    <div class="flex-1 overflow-auto font-mono text-[12px]">
      {#each files as diff (diff.path)}
        <!-- file header -->
        <div class="flex items-center gap-3 px-4 py-2 border-b sticky top-0 z-10 text-[12px]
          {dark ? 'border-gray-800 bg-gray-900/90 text-gray-300' : 'border-gray-200 bg-gray-50/90 text-gray-700'}">
          {#if diff.oldPath}
            <span class="{dark ? 'text-gray-500' : 'text-gray-400'}">{diff.oldPath}</span>
            <span class="{dark ? 'text-gray-600' : 'text-gray-400'}">→</span>
          {/if}
          <span class="font-medium">{diff.path}</span>
          <span class="ml-auto text-[11px] px-1.5 py-px rounded
            {diff.status === 'added'   ? 'bg-green-900/40 text-green-400' :
             diff.status === 'deleted' ? 'bg-red-900/40 text-red-400'     :
             diff.status === 'renamed' ? 'bg-yellow-900/40 text-yellow-400' :
             (dark ? 'bg-gray-800 text-gray-500' : 'bg-gray-200 text-gray-500')}">
            {diff.status}
          </span>
          <span class="text-green-400">+{diff.addCount}</span>
          <span class="text-red-400">-{diff.delCount}</span>
        </div>

        {#if diff.hunks.length === 0}
          <div class="px-4 py-2 text-[11px] {dark ? 'text-gray-600' : 'text-gray-400'}">No diff available</div>
        {:else}
          {#each diff.hunks as hunk}
            <HunkView {hunk} />
          {/each}
        {/if}
      {/each}
    </div>
  {/if}
</div>
