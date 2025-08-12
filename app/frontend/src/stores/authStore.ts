// SOHOAAS Authentication Store
// External state management using Zustand - outside React rendering cycle
// Following the SOHOAAS 5-agent PoC system architecture

import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import type { AuthState, User } from '../types'
import { sohoaasApi } from '../services/api'

interface AuthStore extends AuthState {
  // Actions
  checkAuth: () => Promise<void>
  login: () => Promise<void>
  logout: () => void
  refreshToken: () => Promise<void>
  setUser: (user: User) => void
  clearError: () => void
}

export const useAuthStore = create<AuthStore>()(
  subscribeWithSelector((set, get) => ({
    // Initial state
    isAuthenticated: false,
    token: null,
    user: null,
    loading: false,
    error: null,

    // Actions
    checkAuth: async () => {
      set({ loading: true, error: null })
      
      try {
        const token = await sohoaasApi.getAuthToken()
        
        if (token?.access_token && token.valid) {
          set({ 
            isAuthenticated: true, 
            token,
            loading: false 
          })
          
          // Try to get user capabilities to build user profile
          const capabilities = await sohoaasApi.getPersonalCapabilities(token.access_token)
          if (capabilities) {
            const user: User = {
              user_id: `frontend_user_${Date.now()}`,
              oauth_tokens: {
                google: {
                  access_token: token.access_token,
                  token_type: token.token_type
                }
              },
              connected_services: ['gmail', 'calendar', 'docs', 'drive']
            }
            set({ user })
          }
        } else {
          set({ 
            isAuthenticated: false, 
            token: null, 
            user: null,
            loading: false 
          })
        }
      } catch (error) {
        console.error('Auth check failed:', error)
        set({ 
          isAuthenticated: false, 
          token: null, 
          user: null,
          loading: false,
          error: error instanceof Error ? error.message : 'Authentication check failed'
        })
      }
    },

    login: async () => {
      set({ loading: true, error: null })
      
      try {
        await sohoaasApi.initiateGoogleAuth()
        
        // Poll for auth status after login attempt
        const pollAuth = async () => {
          const token = await sohoaasApi.getAuthToken()
          if (token?.access_token && token.valid) {
            set({ 
              isAuthenticated: true, 
              token,
              loading: false 
            })
            return true
          }
          return false
        }

        // Poll every 2 seconds for up to 30 seconds
        let attempts = 0
        const maxAttempts = 15
        
        const pollInterval = setInterval(async () => {
          attempts++
          const success = await pollAuth()
          
          if (success || attempts >= maxAttempts) {
            clearInterval(pollInterval)
            if (!success) {
              set({ 
                loading: false,
                error: 'Login timeout - please try again'
              })
            }
          }
        }, 2000)

      } catch (error) {
        console.error('Login failed:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Login failed'
        })
      }
    },

    logout: () => {
      set({ 
        isAuthenticated: false, 
        token: null, 
        user: null,
        loading: false,
        error: null
      })
    },

    refreshToken: async () => {
      const { checkAuth } = get()
      await checkAuth()
    },

    setUser: (user: User) => {
      set({ user })
    },

    clearError: () => {
      set({ error: null })
    }
  }))
)

// Auto-refresh token every 30 minutes
setInterval(() => {
  const { isAuthenticated, refreshToken } = useAuthStore.getState()
  if (isAuthenticated) {
    refreshToken()
  }
}, 30 * 60 * 1000)

// Subscribe to auth changes for logging
useAuthStore.subscribe(
  (state) => state.isAuthenticated,
  (isAuthenticated) => {
    console.log('Auth status changed:', isAuthenticated ? 'Authenticated' : 'Not authenticated')
  }
)
