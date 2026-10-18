import { ref } from 'vue'
import { defineStore } from 'pinia'
import { api, APIError } from '../api'
import type { User } from '../types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const initialized = ref(false)
  const error = ref('')
  let initialization: Promise<void> | null = null
  async function restore() {
    if (initialized.value) return
    if (!initialization)
      initialization = (async () => {
        try {
          user.value = (await api.me()).user
          error.value = ''
        } catch (err) {
          user.value = null
          if (!(err instanceof APIError && err.status === 401)) error.value = (err as Error).message
        } finally {
          initialized.value = true
          initialization = null
        }
      })()
    await initialization
  }
  async function login(email: string, password: string) {
    user.value = (await api.login(email, password)).user
    initialized.value = true
    error.value = ''
  }
  async function logout() {
    await api.logout()
    user.value = null
    initialized.value = true
  }
  function expire() {
    user.value = null
    initialized.value = true
  }
  return { user, initialized, error, restore, login, logout, expire }
})
