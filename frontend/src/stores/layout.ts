import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export type LayoutMode = 'desktop' | 'mobile'

const STORAGE_KEY = 'life-sim-layout-mode'

function readStoredMode(): LayoutMode | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw === 'desktop' || raw === 'mobile') return raw
    return null
  } catch {
    return null
  }
}

function writeStoredMode(mode: LayoutMode) {
  try {
    localStorage.setItem(STORAGE_KEY, mode)
  } catch {
    // ignore storage errors (private mode / disabled storage)
  }
}

function defaultModeByScreen(): LayoutMode {
  if (typeof window === 'undefined') return 'desktop'
  try {
    return window.matchMedia('(max-width: 768px)').matches ? 'mobile' : 'desktop'
  } catch {
    return 'desktop'
  }
}

export const useLayoutStore = defineStore('layout', () => {
  const mode = ref<LayoutMode>(readStoredMode() ?? defaultModeByScreen())
  const isMobile = computed(() => mode.value === 'mobile')

  function setMode(next: LayoutMode) {
    mode.value = next
    writeStoredMode(next)
  }

  function toggleMode() {
    setMode(mode.value === 'mobile' ? 'desktop' : 'mobile')
  }

  return { mode, isMobile, setMode, toggleMode }
})
