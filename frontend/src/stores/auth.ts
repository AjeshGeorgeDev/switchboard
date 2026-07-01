import { ref } from 'vue'
import { defineStore } from 'pinia'
import { api } from '../api'

export interface User {
  id: string
  username: string
  email: string
  display_name: string
  auth_type: string
  roles: string[]
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const loaded = ref(false)

  async function fetchMe() {
    try {
      user.value = await api.get<User>('/api/auth/me')
    } catch {
      try {
        await api.post('/api/auth/refresh', {})
        user.value = await api.get<User>('/api/auth/me')
      } catch {
        user.value = null
      }
    } finally {
      loaded.value = true
    }
  }

  async function login(email: string, password: string, rememberMe = true) {
    await api.post('/api/auth/login', { email, password, remember_me: rememberMe })
    await fetchMe()
  }

  async function logout() {
    await api.post('/api/auth/logout', {})
    user.value = null
  }

  function hasRole(role: string) {
    return user.value?.roles.includes(role) ?? false
  }

  function isAdmin() {
    return hasRole('admin')
  }

  function canViewSecurity() {
    return hasRole('admin') || hasRole('security-team')
  }

  return { user, loaded, fetchMe, login, logout, hasRole, isAdmin, canViewSecurity }
})
