// SOHOAAS Service Store
// External state management using Zustand - outside React rendering cycle
// Following the SOHOAAS 5-agent PoC system architecture

import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import type { ServiceState } from '../types'
import { sohoaasApi } from '../services/api'
import { useAuthStore } from './authStore'

interface ServiceStore extends ServiceState {
  // Actions
  loadServiceCatalog: () => Promise<void>
  loadAgents: () => Promise<void>
  validateCatalog: () => Promise<any>
  refreshServices: () => Promise<void>
  clearError: () => void
}

export const useServiceStore = create<ServiceStore>()(
  subscribeWithSelector((set, get) => ({
    // Initial state
    catalog: null,
    agents: [],
    loading: false,
    error: null,

    // Actions
    loadServiceCatalog: async () => {
      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const catalog = await sohoaasApi.getServiceCatalog(token.access_token)
        set({ 
          catalog,
          loading: false 
        })
      } catch (error) {
        console.error('Failed to load service catalog:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to load service catalog'
        })
      }
    },

    loadAgents: async () => {
      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const agents = await sohoaasApi.getAgents(token.access_token)
        set({ 
          agents,
          loading: false 
        })
      } catch (error) {
        console.error('Failed to load agents:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to load agents'
        })
      }
    },

    validateCatalog: async () => {
      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const result = await sohoaasApi.validateServiceCatalog(token.access_token)
        set({ loading: false })
        return result
      } catch (error) {
        console.error('Failed to validate catalog:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to validate catalog'
        })
        return null
      }
    },

    refreshServices: async () => {
      const { loadServiceCatalog, loadAgents } = get()
      await Promise.all([
        loadServiceCatalog(),
        loadAgents()
      ])
    },

    clearError: () => {
      set({ error: null })
    }
  }))
)

// Subscribe to service changes for logging
useServiceStore.subscribe(
  (state) => state.catalog?.count,
  (serviceCount) => {
    if (serviceCount !== undefined) {
      console.log('Service catalog loaded with', serviceCount, 'services')
    }
  }
)
