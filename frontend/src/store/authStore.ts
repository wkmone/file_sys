import { create } from 'zustand'
import type { User } from '../types/user'

interface AuthState {
  accessToken: string | null
  user: User | null
  isAuthenticated: boolean
  setAccessToken: (token: string) => void
  setUser: (user: User) => void
  login: (accessToken: string, user: User) => void
  logout: () => void
}

// Persist token to localStorage so new tabs can restore auth.
// refresh token is httpOnly cookie — managed by the backend, not stored here.
const TOKEN_KEY = 'fs_access_token'
const USER_KEY = 'fs_user'

function loadFromStorage() {
  try {
    const token = localStorage.getItem(TOKEN_KEY)
    const user = localStorage.getItem(USER_KEY)
    return {
      accessToken: token,
      user: user ? JSON.parse(user) : null,
      isAuthenticated: !!(token && user),
    }
  } catch {
    return { accessToken: null, user: null, isAuthenticated: false }
  }
}

const initial = loadFromStorage()

export const useAuthStore = create<AuthState>((set) => ({
  ...initial,

  setAccessToken: (token: string) => {
    localStorage.setItem(TOKEN_KEY, token)
    set({ accessToken: token })
  },

  setUser: (user: User) => {
    localStorage.setItem(USER_KEY, JSON.stringify(user))
    set({ user })
  },

  login: (accessToken: string, user: User) => {
    localStorage.setItem(TOKEN_KEY, accessToken)
    localStorage.setItem(USER_KEY, JSON.stringify(user))
    set({ accessToken, user, isAuthenticated: true })
  },

  logout: () => {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
    set({ accessToken: null, user: null, isAuthenticated: false })
  },
}))
