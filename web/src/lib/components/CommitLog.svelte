<script lang="ts">
  import { api, type Commit } from '../api'
  import { appState } from '../state.svelte'

  let { onSelect }: { onSelect: (commit: Commit) => void } = $props()

  let commits = $state<Commit[]>([])
  let selected = $state<string | null>(null)

  async function load() {
    commits = await api.commits()
  }

  $effect(() => { load() })

  function pick(c: Commit) {
    selected = c.hash
    onSelect(c)
  }

  const dark = $derived(appState.theme !== 'light')
</script>

<div class="flex flex-col h-full">
  <div class="px-2 py-1.5 text-[10px] font-semibold uppercase tracking-wider {dark ? 'text-gray-500' : 'text-gray-400'}">
    Recent commits
  </div>
  <div class="flex-1 overflow-y-auto">
    {#each commits as c (c.hash)}
      <button
        class="w-full text-left px-2 py-1.5 flex flex-col gap-0.5 transition-colors
          {selected === c.hash
            ? (dark ? 'bg-blue-900/40 text-gray-100' : 'bg-blue-50 text-gray-900')
            : (dark ? 'hover:bg-gray-800/60 text-gray-300' : 'hover:bg-gray-100 text-gray-700')}"
        onclick={() => pick(c)}
      >
        <span class="text-[11px] leading-tight truncate">{c.message}</span>
        <span class="text-[10px] {dark ? 'text-gray-500' : 'text-gray-400'}">{c.short} · {c.author} · {c.date}</span>
      </button>
    {:else}
      <div class="px-2 py-3 text-[11px] {dark ? 'text-gray-600' : 'text-gray-400'}">No commits yet</div>
    {/each}
  </div>
</div>
