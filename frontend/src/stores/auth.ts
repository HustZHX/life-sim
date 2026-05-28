import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type AuthUser } from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const enabled = ref(false)
  const gatePassed = ref(false)
  const user = ref<AuthUser | null>(null)
  const checking = ref(true)

  async function loadStatus() {
    checking.value = true
    try {
      const status = await api.authStatus()
      enabled.value = status.enabled
      gatePassed.value = status.gate_passed
      if (status.logged_in) {
        user.value = await api.authMe()
      } else {
        user.value = null
      }
    } catch {
      enabled.value = false
      gatePassed.value = false
      user.value = null
    } finally {
      checking.value = false
    }
  }

  async function verifyGate(code: string) {
    await api.authGate(code)
    gatePassed.value = true
  }

  async function login(username: string, password: string) {
    const me = await api.authLogin(username, password)
    user.value = me
  }

  async function refreshSession() {
    await api.authRefresh()
    user.value = await api.authMe()
  }

  async function logout() {
    try {
      await api.authLogout()
    } finally {
      user.value = null
    }
  }

  function clearGate() {
    gatePassed.value = false
  }

  return {
    enabled,
    gatePassed,
    user,
    checking,
    loadStatus,
    verifyGate,
    login,
    refreshSession,
    logout,
    clearGate,
  }
})
