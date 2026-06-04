<script lang="ts">
  import type { Hunk } from '../api'

  interface Props {
    hunk: Hunk
    onExpand?: () => void
  }

  let { hunk, onExpand }: Props = $props()
</script>

<div class="font-mono text-[12px]">
  <!-- hunk header -->
  <div class="flex items-center gap-2 bg-blue-900/20 text-blue-400 px-3 py-0.5 select-none">
    <span class="text-blue-500/60">{hunk.header.split('@@')[1]?.trim() || ''}</span>
    <span class="text-blue-400/40 text-[11px] ml-auto">
      @@ -{hunk.oldStart},{hunk.oldCount} +{hunk.newStart},{hunk.newCount} @@
    </span>
  </div>

  <!-- lines -->
  {#each hunk.lines as line}
    <div
      class="flex group leading-5 {line.type === 'add' ? 'diff-add' : line.type === 'del' ? 'diff-del' : 'diff-context'}"
    >
      <span class="line-num shrink-0 select-none text-[11px] leading-5 pl-2 pr-1 w-10 text-right">
        {line.type !== 'add' ? (line.oldNo ?? '') : ''}
      </span>
      <span class="line-num shrink-0 select-none text-[11px] leading-5 pl-1 pr-2 w-10 text-right">
        {line.type !== 'del' ? (line.newNo ?? '') : ''}
      </span>
      <span class="px-1 text-xs leading-5 shrink-0 w-4 select-none">
        {line.type === 'add' ? '+' : line.type === 'del' ? '-' : ' '}
      </span>
      <span class="px-2 whitespace-pre overflow-hidden">{line.content}</span>
    </div>
  {/each}
</div>
