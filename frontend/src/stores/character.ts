import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Character, LifeNode, Profile, ResolveCandidate } from '@/api/client'

export const useCharacterStore = defineStore('character', () => {
  const character = ref<Character | null>(null)
  const profile = ref<Profile | null>(null)
  const candidates = ref<ResolveCandidate[]>([])
  const nodes = ref<LifeNode[]>([])
  const loading = ref(false)
  const jobProgress = ref(0)

  function reset() {
    character.value = null
    profile.value = null
    candidates.value = []
    nodes.value = []
    loading.value = false
    jobProgress.value = 0
  }

  return {
    character,
    profile,
    candidates,
    nodes,
    loading,
    jobProgress,
    reset,
  }
})
