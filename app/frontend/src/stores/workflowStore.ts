// SOHOAAS Workflow Store
// External state management using Zustand - outside React rendering cycle
// Following the SOHOAAS 5-agent PoC system architecture

import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import type { 
  WorkflowState, 
  Workflow, 
  PipelineResult,
  IntentAnalysis 
} from '../types'
import { sohoaasApi } from '../services/api'
import { useAuthStore } from './authStore'

interface WorkflowStore extends WorkflowState {
  // Actions
  loadWorkflows: () => Promise<void>
  createWorkflow: (userMessage: string) => Promise<PipelineResult | null>
  startDiscovery: (userMessage: string) => Promise<void>
  continueDiscovery: (userResponse: string) => Promise<void>
  executeWorkflow: (workflowId: string, parameters?: Record<string, any>) => Promise<any>
  getWorkflow: (workflowId: string) => Promise<void>
  analyzeIntent: (userMessage: string) => Promise<IntentAnalysis | null>
  generateWorkflow: (userMessage: string, intentAnalysis?: IntentAnalysis) => Promise<any>
  testPipeline: (userMessage: string) => Promise<PipelineResult | null>
  setCurrentWorkflow: (workflow: Workflow | null) => void
  clearError: () => void
  reset: () => void
}

export const useWorkflowStore = create<WorkflowStore>()(
  subscribeWithSelector((set, get) => ({
    // Initial state
    workflows: [],
    currentWorkflow: null,
    discovery: null,
    loading: false,
    error: null,

    // Actions
    loadWorkflows: async () => {
      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const workflows = await sohoaasApi.getWorkflows(token.access_token)
        set({ 
          workflows,
          loading: false 
        })
      } catch (error) {
        console.error('Failed to load workflows:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to load workflows'
        })
      }
    },

    createWorkflow: async (userMessage: string): Promise<PipelineResult | null> => {
      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const result = await sohoaasApi.testCompletePipeline(userMessage, token.access_token)
        
        if (result) {
          // Refresh workflows list after creation
          await get().loadWorkflows()
          set({ loading: false })
          return result
        } else {
          throw new Error('Failed to create workflow')
        }
      } catch (error) {
        console.error('Failed to create workflow:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to create workflow'
        })
        return null
      }
    },

    startDiscovery: async (userMessage: string) => {
      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const discovery = await sohoaasApi.startWorkflowDiscovery(userMessage, token.access_token)
        set({ 
          discovery,
          loading: false 
        })
      } catch (error) {
        console.error('Failed to start workflow discovery:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to start workflow discovery'
        })
      }
    },

    continueDiscovery: async (userResponse: string) => {
      const { discovery } = get()
      if (!discovery?.session_id) {
        set({ error: 'No active discovery session' })
        return
      }

      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const updatedDiscovery = await sohoaasApi.continueWorkflowDiscovery(
          discovery.session_id, 
          userResponse, 
          token.access_token
        )
        
        set({ 
          discovery: updatedDiscovery,
          loading: false 
        })
      } catch (error) {
        console.error('Failed to continue workflow discovery:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to continue workflow discovery'
        })
      }
    },

    executeWorkflow: async (workflowId: string, parameters?: Record<string, any>) => {
      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const result = await sohoaasApi.executeWorkflow(workflowId, parameters, token.access_token)
        set({ loading: false })
        return result
      } catch (error) {
        console.error('Failed to execute workflow:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to execute workflow'
        })
        return null
      }
    },

    getWorkflow: async (workflowId: string) => {
      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const workflow = await sohoaasApi.getWorkflow(workflowId, token.access_token)
        set({ 
          currentWorkflow: workflow,
          loading: false 
        })
      } catch (error) {
        console.error('Failed to get workflow:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to get workflow'
        })
      }
    },

    analyzeIntent: async (userMessage: string): Promise<IntentAnalysis | null> => {
      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const analysis = await sohoaasApi.analyzeIntent(userMessage, token.access_token)
        set({ loading: false })
        return analysis
      } catch (error) {
        console.error('Failed to analyze intent:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to analyze intent'
        })
        return null
      }
    },

    generateWorkflow: async (userMessage: string, intentAnalysis?: IntentAnalysis) => {
      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const result = await sohoaasApi.generateWorkflow(userMessage, intentAnalysis, token.access_token)
        set({ loading: false })
        return result
      } catch (error) {
        console.error('Failed to generate workflow:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to generate workflow'
        })
        return null
      }
    },

    testPipeline: async (userMessage: string): Promise<PipelineResult | null> => {
      set({ loading: true, error: null })
      
      try {
        const { token } = useAuthStore.getState()
        if (!token?.access_token) {
          throw new Error('No authentication token available')
        }

        const result = await sohoaasApi.testCompletePipeline(userMessage, token.access_token)
        set({ loading: false })
        return result
      } catch (error) {
        console.error('Failed to test pipeline:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to test pipeline'
        })
        return null
      }
    },

    setCurrentWorkflow: (workflow: Workflow | null) => {
      set({ currentWorkflow: workflow })
    },

    clearError: () => {
      set({ error: null })
    },

    reset: () => {
      set({
        workflows: [],
        currentWorkflow: null,
        discovery: null,
        loading: false,
        error: null
      })
    }
  }))
)

// Subscribe to workflow changes for logging
useWorkflowStore.subscribe(
  (state) => state.workflows.length,
  (workflowCount) => {
    console.log('Workflow count changed:', workflowCount)
  }
)
