// SOHOAAS API Service Layer
// Handles all communication with the SOHOAAS backend and MCP service
// Following the 5-agent PoC system architecture

import type {
  AuthToken,
  ServiceCatalog,
  Workflow,
  PipelineResult,
  Agent,
  WorkflowDiscovery,
  IntentAnalysis
} from '../types'

class SOHOAASApiService {
  // Use environment variables for Docker deployment, fallback to localhost for development
  private readonly PROXY_BASE_URL = import.meta.env.VITE_PROXY_URL || 'http://localhost:3000'
  private readonly BACKEND_BASE_URL = import.meta.env.VITE_BACKEND_URL || 'http://localhost:8081'
  
  // Authentication & Token Management
  async getAuthToken(): Promise<AuthToken | null> {
    try {
      // Get Firebase ID token for backend authentication
      const { useAuthStore } = await import('../stores/authStore')
      const authState = useAuthStore.getState()
      
      if (authState.isAuthenticated && authState.token) {
        return authState.token
      }
      
      // If not authenticated or no token, try to refresh
      if (authState.firebaseUser) {
        const idToken = await authState.firebaseUser.getIdToken(true)
        return {
          access_token: idToken,
          token_type: 'Bearer',
          expires_in: 3600,
          valid: true
        }
      }
      
      return null
    } catch (error) {
      console.error('Failed to get Firebase auth token:', error)
      return null
    }
  }

  // Get Google API access token for MCP service calls
  async getGoogleAccessToken(): Promise<string | null> {
    try {
      const { useAuthStore } = await import('../stores/authStore')
      const authState = useAuthStore.getState()
      
      if (authState.isAuthenticated && authState.googleAccessToken) {
        return authState.googleAccessToken
      }
      
      return null
    } catch (error) {
      console.error('Failed to get Google access token:', error)
      return null
    }
  }

  async checkAuthStatus(): Promise<boolean> {
    const token = await this.getAuthToken()
    return !!(token?.access_token && token.valid)
  }

  async initiateGoogleAuth(): Promise<void> {
    // Firebase Auth handles Google authentication directly
    // This method is kept for compatibility but delegates to the auth store
    const { useAuthStore } = await import('../stores/authStore')
    const { login } = useAuthStore.getState()
    await login()
  }

  // Service Catalog & Capabilities
  async getServiceCatalog(authToken?: string): Promise<ServiceCatalog | null> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      // Try nginx proxy first, fallback to direct backend
      const urls = [`${this.PROXY_BASE_URL}/api/v1/services`, `${this.BACKEND_BASE_URL}/api/v1/services`]
      
      for (const url of urls) {
        try {
          const response = await fetch(url, {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json'
            }
          })

          if (response.ok) {
            return await response.json()
          }
        } catch (error) {
          console.warn(`Failed to get service catalog from ${url}:`, error)
          continue
        }
      }
      
      throw new Error('All service catalog endpoints failed')
    } catch (error) {
      console.error('Failed to get service catalog:', error)
      return null
    }
  }

  async getPersonalCapabilities(authToken?: string): Promise<any> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      // Try nginx proxy first, fallback to direct backend
      const urls = [`${this.PROXY_BASE_URL}/api/v1/capabilities`, `${this.BACKEND_BASE_URL}/api/v1/capabilities`]
      
      for (const url of urls) {
        try {
          const response = await fetch(url, {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json'
            }
          })

          if (response.ok) {
            return await response.json()
          }
        } catch (error) {
          console.warn(`Failed to get capabilities from ${url}:`, error)
          continue
        }
      }
      
      throw new Error('All capabilities endpoints failed')
    } catch (error) {
      console.error('Failed to get personal capabilities:', error)
      return null
    }
  }

  // Agent Management
  async getAgents(authToken?: string): Promise<Agent[]> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      const response = await fetch(`${this.BACKEND_BASE_URL}/api/v1/agents`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      })

      if (!response.ok) {
        throw new Error(`Agents request failed: ${response.status}`)
      }

      const data = await response.json()
      return data.agents || []
    } catch (error) {
      console.error('Failed to get agents:', error)
      return []
    }
  }

  // Intent Analysis
  async analyzeIntent(userMessage: string, authToken?: string): Promise<IntentAnalysis | null> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      // Try nginx proxy first, fallback to direct backend
      const urls = [`${this.PROXY_BASE_URL}/api/v1/intent/analyze`, `${this.BACKEND_BASE_URL}/api/v1/intent/analyze`]
      
      for (const url of urls) {
        try {
          const response = await fetch(url, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json'
            },
            body: JSON.stringify({ user_message: userMessage })
          })

          if (response.ok) {
            return await response.json()
          }
        } catch (error) {
          console.warn(`Failed to analyze intent from ${url}:`, error)
          continue
        }
      }
      
      throw new Error('All intent analysis endpoints failed')
    } catch (error) {
      console.error('Failed to analyze intent:', error)
      return null
    }
  }

  // Workflow Generation (Two-step process: Intent Analysis → Workflow Generation)
  async generateWorkflow(
    userMessage: string, 
    intentAnalysis?: IntentAnalysis,
    authToken?: string
  ): Promise<any> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      console.log('[API] Starting workflow generation process for:', userMessage)

      // Step 1: Intent Analysis (if not provided)
      let validatedIntent = intentAnalysis
      if (!validatedIntent) {
        console.log('[API] Step 1: Analyzing user intent...')
        const intentResponse = await fetch(`${this.BACKEND_BASE_URL}/api/v1/intent/analyze`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            workflow_intent: {
              user_message: userMessage,
              workflow_pattern: '',
              trigger_conditions: {},
              action_sequence: [],
              data_requirements: [],
              user_parameters: []
            }
          })
        })

        if (!intentResponse.ok) {
          throw new Error(`Intent analysis failed: ${intentResponse.status}`)
        }

        const intentResult = await intentResponse.json()
        console.log('[API] Intent analysis result:', intentResult)
        
        if (!intentResult.agent_response?.output) {
          throw new Error('Invalid intent analysis response')
        }
        
        validatedIntent = intentResult.agent_response.output
      }

      // Step 2: Workflow Generation using validated intent and original user input
      console.log('[API] Step 2: Generating workflow with validated intent and user input...')
      const workflowResponse = await fetch(`${this.BACKEND_BASE_URL}/api/v1/workflow/generate`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          user_intent: userMessage,
          validated_intent: validatedIntent
        })
      })

      if (!workflowResponse.ok) {
        throw new Error(`Workflow generation failed: ${workflowResponse.status}`)
      }

      const result = await workflowResponse.json()
      console.log('[API] Workflow generation completed:', result)
      return result

    } catch (error) {
      console.error('Failed to generate workflow:', error)
      return null
    }
  }

  // Workflow Management
  async getWorkflows(authToken?: string): Promise<Workflow[]> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      const response = await fetch(`${this.BACKEND_BASE_URL}/api/v1/workflows`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      })

      if (!response.ok) {
        throw new Error(`Get workflows failed: ${response.status}`)
      }

      const data = await response.json()
      return data.workflows || []
    } catch (error) {
      console.error('Failed to get workflows:', error)
      return []
    }
  }

  async getWorkflow(workflowId: string, authToken?: string): Promise<Workflow | null> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      const response = await fetch(`${this.BACKEND_BASE_URL}/api/v1/workflows/${workflowId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      })

      if (!response.ok) {
        throw new Error(`Get workflow failed: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Failed to get workflow:', error)
      return null
    }
  }

  async executeWorkflow(workflowId: string, parameters?: Record<string, any>, authToken?: string): Promise<any> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      const response = await fetch(`${this.BACKEND_BASE_URL}/api/v1/workflow/execute`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          workflow_id: workflowId,
          parameters: parameters || {},
          user_id: `frontend_user_${Date.now()}`
        })
      })

      if (!response.ok) {
        throw new Error(`Workflow execution failed: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Failed to execute workflow:', error)
      return null
    }
  }

  // Workflow Discovery & Creation
  async startWorkflowDiscovery(userMessage: string, authToken?: string): Promise<WorkflowDiscovery | null> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      const response = await fetch(`${this.BACKEND_BASE_URL}/api/v1/workflow/discover`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          user_message: userMessage,
          user_id: `frontend_user_${Date.now()}`
        })
      })

      if (!response.ok) {
        throw new Error(`Workflow discovery failed: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Failed to start workflow discovery:', error)
      return null
    }
  }

  async continueWorkflowDiscovery(
    sessionId: string, 
    userResponse: string, 
    authToken?: string
  ): Promise<WorkflowDiscovery | null> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      const response = await fetch(`${this.BACKEND_BASE_URL}/api/v1/workflow/continue`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          session_id: sessionId,
          user_response: userResponse
        })
      })

      if (!response.ok) {
        throw new Error(`Continue workflow discovery failed: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Failed to continue workflow discovery:', error)
      return null
    }
  }

  // Complete Pipeline Testing
  async testCompletePipeline(userMessage: string, authToken?: string): Promise<PipelineResult | null> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      const response = await fetch(`${this.BACKEND_BASE_URL}/api/v1/test/pipeline`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          user_id: `frontend_user_${Date.now()}`,
          user_message: userMessage,
          conversation_history: [],
          user: {
            user_id: `frontend_user_${Date.now()}`,
            oauth_tokens: {
              google: {
                access_token: token,
                token_type: 'Bearer'
              }
            },
            connected_services: ['gmail', 'calendar', 'docs', 'drive']
          }
        })
      })

      if (!response.ok) {
        throw new Error(`Pipeline test failed: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Failed to test complete pipeline:', error)
      return null
    }
  }

  // Service Validation
  async validateServiceCatalog(authToken?: string): Promise<any> {
    try {
      const token = authToken || (await this.getAuthToken())?.access_token
      if (!token) {
        throw new Error('No authentication token available')
      }

      const response = await fetch(`${this.BACKEND_BASE_URL}/api/v1/validate/catalog`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      })

      if (!response.ok) {
        throw new Error(`Service validation failed: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Failed to validate service catalog:', error)
      return null
    }
  }
}

// Export singleton instance
export const sohoaasApi = new SOHOAASApiService()
export default sohoaasApi