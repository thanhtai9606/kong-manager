import { defineStore } from 'pinia'
import { KM_TOKEN_KEY, setKmToken } from '@/services/apiService'
import { jwtSubject } from '@/utils/jwt'

function readStoredToken(): string | null {
  if (typeof sessionStorage === 'undefined') {
    return null
  }
  return sessionStorage.getItem(KM_TOKEN_KEY)
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: readStoredToken() as string | null,
  }),
  getters: {
    isAuthenticated: (s): boolean => !!s.token,
    username(state): string | null {
      return jwtSubject(state.token)
    },
  },
  actions: {
    setSession(token: string) {
      this.token = token
      setKmToken(token)
    },
    clearSession() {
      this.token = null
      setKmToken(null)
    },
  },
})
