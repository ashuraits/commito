<script lang="ts">
  import { appState } from '../state.svelte'

  let { path }: { path: string } = $props()

  const IMAGE_EXT = new Set(['png','jpg','jpeg','gif','webp','svg','bmp','ico','avif'])
  const VIDEO_EXT = new Set(['mp4','webm','ogg','mov','mkv','avi'])
  const AUDIO_EXT = new Set(['mp3','wav','ogg','flac','aac','m4a','opus'])

  function ext(p: string) {
    return p.split('.').pop()?.toLowerCase() ?? ''
  }

  const type = $derived.by(() => {
    const e = ext(path)
    if (IMAGE_EXT.has(e)) return 'image'
    if (VIDEO_EXT.has(e)) return 'video'
    if (AUDIO_EXT.has(e)) return 'audio'
    return 'unknown'
  })

  const src = $derived(`/api/file?path=${encodeURIComponent(path)}`)
  const dark = $derived(appState.theme !== 'light')
</script>

<div class="h-full flex flex-col items-center justify-center p-6 {dark ? 'bg-gray-950' : 'bg-gray-100'}">
  <div class="text-[11px] mb-3 {dark ? 'text-gray-500' : 'text-gray-400'}">{path}</div>

  {#if type === 'image'}
    <img
      {src}
      alt={path}
      class="max-w-full max-h-full object-contain rounded shadow-lg"
      style="max-height: calc(100vh - 120px)"
    />

  {:else if type === 'video'}
    <video
      {src}
      controls
      class="max-w-full rounded shadow-lg"
      style="max-height: calc(100vh - 120px)"
    >
      Your browser does not support video playback.
    </video>

  {:else if type === 'audio'}
    <div class="flex flex-col items-center gap-4">
      <svg class="w-16 h-16 {dark ? 'text-gray-600' : 'text-gray-300'}" fill="currentColor" viewBox="0 0 24 24">
        <path d="M12 3v10.55A4 4 0 1 0 14 17V7h4V3h-6z"/>
      </svg>
      <audio {src} controls class="w-72">
        Your browser does not support audio playback.
      </audio>
    </div>

  {:else}
    <div class="text-sm {dark ? 'text-gray-500' : 'text-gray-400'}">Cannot preview this file type</div>
  {/if}
</div>
