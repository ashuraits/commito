<script lang="ts">
  import { appState } from '../state.svelte'

  interface MenuItem {
    label: string
    shortcut?: string
    action: () => void
    disabled?: boolean
    danger?: boolean
    separator?: never
  }
  interface Separator { separator: true }

  interface Props {
    x: number
    y: number
    items: (MenuItem | Separator)[]
    onClose: () => void
  }

  let { x, y, items, onClose }: Props = $props()

  const light = $derived(appState.theme === 'light')

  function handleKey(e: KeyboardEvent) {
    if (e.key === 'Escape') onClose()
  }
</script>

<svelte:window onkeydown={handleKey} />

<!-- backdrop -->
<div class="fixed inset-0 z-40" onmousedown={onClose}></div>

<!-- menu -->
<div
  class="fixed z-50 min-w-[160px] py-1 rounded-md shadow-lg border text-[12px]
    {light ? 'bg-white border-gray-200' : 'bg-gray-800 border-gray-700'}"
  style="left: {x}px; top: {y}px"
>
  {#each items as item}
    {#if 'separator' in item}
      <div class="my-1 border-t {light ? 'border-gray-200' : 'border-gray-700'}"></div>
    {:else}
      <button
        class="w-full text-left flex items-center justify-between px-3 py-1.5 transition-colors
          {item.disabled ? 'opacity-40 cursor-not-allowed' : (item.danger
            ? 'text-red-400 hover:bg-red-500/10'
            : (light ? 'text-gray-700 hover:bg-gray-100' : 'text-gray-200 hover:bg-gray-700'))}"
        disabled={item.disabled}
        onclick={() => { if (!item.disabled) { item.action(); onClose() } }}
      >
        <span>{item.label}</span>
        {#if item.shortcut}
          <span class="ml-6 text-[10px] {light ? 'text-gray-400' : 'text-gray-500'}">{item.shortcut}</span>
        {/if}
      </button>
    {/if}
  {/each}
</div>
