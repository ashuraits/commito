<script lang="ts">
  import { EditorView, keymap, lineNumbers, highlightActiveLine } from '@codemirror/view'
  import { EditorState } from '@codemirror/state'
  import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
  import { indentOnInput, syntaxHighlighting, defaultHighlightStyle, bracketMatching } from '@codemirror/language'
  import { searchKeymap, highlightSelectionMatches } from '@codemirror/search'
  import { oneDark } from '@codemirror/theme-one-dark'
  import { javascript } from '@codemirror/lang-javascript'
  import { python } from '@codemirror/lang-python'
  import { go } from '@codemirror/lang-go'
  import { css } from '@codemirror/lang-css'
  import { html } from '@codemirror/lang-html'
  import { json } from '@codemirror/lang-json'
  import { markdown } from '@codemirror/lang-markdown'
  import { sql } from '@codemirror/lang-sql'
  import { rust } from '@codemirror/lang-rust'
  import { api } from '../api'
  import { appState } from '../state.svelte'

  const IMAGE_EXTS = new Set(['png','jpg','jpeg','gif','webp','svg','ico','bmp','tiff','avif'])

  function isImage(filePath: string) {
    return IMAGE_EXTS.has(filePath.split('.').pop()?.toLowerCase() ?? '')
  }

  interface Props {
    path: string
    theme: 'dark' | 'light'
  }

  let { path, theme }: Props = $props()

  const imageUrl = $derived(isImage(path) ? `/api/file?path=${encodeURIComponent(path)}` : null)

  let container: HTMLDivElement | undefined = $state(undefined)
  let saving = $state(false)
  let saved = $state(false)
  let isBinary = $state(false)
  let saveTimer: ReturnType<typeof setTimeout>

  function getLanguage(filePath: string) {
    const ext = filePath.split('.').pop()?.toLowerCase()
    switch (ext) {
      case 'js': case 'jsx': case 'mjs': return javascript({ jsx: true })
      case 'ts': case 'tsx': return javascript({ typescript: true, jsx: true })
      case 'py': return python()
      case 'go': return go()
      case 'css': case 'scss': return css()
      case 'html': case 'svelte': return html()
      case 'json': return json()
      case 'md': return markdown()
      case 'sql': return sql()
      case 'rs': return rust()
      default: return []
    }
  }

  const lightTheme = EditorView.theme({
    '&': { background: '#ffffff', color: '#1f2328', height: '100%' },
    '.cm-scroller': { overflow: 'auto' },
    '.cm-content': { caretColor: '#0969da' },
    '.cm-gutters': { background: '#f6f8fa', color: '#9198a1', border: 'none', borderRight: '1px solid #d0d7de' },
    '.cm-activeLineGutter': { background: '#eef1f4' },
    '.cm-activeLine': { background: '#eef1f4' },
  }, { dark: false })

  const darkThemeOverride = EditorView.theme({
    '&': { height: '100%' },
    '.cm-scroller': { overflow: 'auto' },
  })

  $effect(() => {
    // read reactive deps first so Svelte tracks them
    const el = container
    const p = path
    const t = theme
    if (!el || isImage(p)) return

    let view: EditorView | null = null
    let cancelled = false
    isBinary = false

    api.readFile(p).then(content => {
      if (cancelled || !el) return
      if (content.includes('\0')) { isBinary = true; return }

      const lang = getLanguage(p)
      const updateListener = EditorView.updateListener.of((update) => {
        if (!update.docChanged) return
        clearTimeout(saveTimer)
        saving = true
        saved = false
        saveTimer = setTimeout(async () => {
          await api.writeFile(p, update.state.doc.toString())
          saving = false
          saved = true
          setTimeout(() => saved = false, 1500)
        }, 500)
      })

      const extensions = [
        lineNumbers(),
        highlightActiveLine(),
        history(),
        indentOnInput(),
        bracketMatching(),
        highlightSelectionMatches(),
        syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
        keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap]),
        EditorView.lineWrapping,
        updateListener,
        ...(Array.isArray(lang) ? lang : [lang]),
        t === 'dark' ? [oneDark, darkThemeOverride] : lightTheme,
      ]

      const state = EditorState.create({ doc: content, extensions })
      view = new EditorView({ state, parent: el })
      view.focus()
    })

    return () => {
      cancelled = true
      clearTimeout(saveTimer)
      view?.destroy()
    }
  })
</script>

<div class="h-full flex flex-col overflow-hidden">
  <div class="flex items-center gap-3 px-4 py-2 border-b border-gray-800 shrink-0 bg-gray-900/50 text-[12px]">
    <span class="text-gray-300 font-medium truncate">{path}</span>
    <div class="ml-auto flex items-center gap-2 shrink-0">
      {#if saving}
        <span class="text-gray-500 text-[11px]">saving…</span>
      {:else if saved}
        <span class="text-green-500 text-[11px]">saved</span>
      {/if}
      <button
        class="px-2 py-0.5 rounded border border-gray-700 text-gray-400 hover:text-gray-200 hover:border-gray-500 transition-colors text-[11px]"
        onclick={() => appState.viewMode = 'diff'}
        title="Back to diff (Escape)"
      >
        ← diff
      </button>
    </div>
  </div>

  {#if imageUrl}
    <div class="flex-1 overflow-auto flex items-center justify-center p-6 bg-[repeating-conic-gradient(#3d444d_0%_25%,transparent_0%_50%)] bg-[length:20px_20px]">
      <img
        src={imageUrl}
        alt={path}
        class="max-w-full max-h-full object-contain shadow-2xl"
        style="image-rendering: pixelated"
      />
    </div>
  {:else if isBinary}
    <div class="flex-1 flex items-center justify-center text-gray-500 text-sm select-none">
      This file is a binary
    </div>
  {:else}
    <div bind:this={container} class="flex-1 overflow-auto text-[13px] [&_.cm-editor]:h-full [&_.cm-editor]:text-[13px] [&_.cm-editor]:outline-none"></div>
  {/if}
</div>
