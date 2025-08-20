// SOHOAAS Frontend Types
// Following the 5-agent PoC system architecture

export interface AuthToken {
  access_token: string
  expires_in: number
  token_type: string
  valid: boolean
}

export interface User {
  user_id: string
  email: string
  name: string
  oauth_tokens: {
    google: {
      access_token: string
      token_type: string
    }
  }
  connected_services: string[]
}

export interface GoogleWorkspaceService {
  service: string
  functions: string[]
  status: 'connected' | 'disconnected' | 'error'
  metadata: {
    enabled: boolean
  }
}

export interface ServiceCatalog {
  count: number
  services: GoogleWorkspaceService[]
}

export interface WorkflowStep {
  id: string
  name: string
  service: string
  action: string
  inputs: Record<string, any>
  outputs: Record<string, any>
  depends_on?: string[]
}

export interface UserParameter {
  name: string
  type: 'string' | 'number' | 'boolean' | 'array' | 'object'
  required: boolean
  description: string
  prompt?: string
  default?: any
  validation?: Record<string, any>
}

export interface ServiceBinding {
  service: string
  oauth_scopes: string[]
  endpoint?: string
}

export interface Workflow {
  id: string
  name: string
  description: string
  status: 'draft' | 'active' | 'completed' | 'error'
  created_at: string
  updated_at?: string
  user_message: string
  steps: WorkflowStep[]
  user_parameters: UserParameter[]
  service_bindings: ServiceBinding[]
  trigger?: {
    type: 'schedule' | 'manual' | 'event'
    schedule?: string
    event?: string
  }
}

// Extended type to handle both parsed Workflow and WorkflowFile from backend
export type WorkflowOrFile = Workflow | {
  id: string
  name: string
  description: string
  status: string
  filename: string
  user_id: string
  content: string
  created_at: string
  updated_at: string
  user_message?: string
  steps?: WorkflowStep[]
  user_parameters?: UserParameter[]
  service_bindings?: ServiceBinding[]
  parsed_data?: {
    name?: string
    description?: string
    version?: string
    original_intent?: string
    steps?: Array<{
      id: string
      name: string
      action: string
      parameters: Record<string, any>
      depends_on?: string[]
    }>
    user_parameters?: Record<string, {
      type: string
      prompt?: string
      required?: boolean
      description?: string
      validation?: any
    }>
    service_bindings?: Record<string, {
      endpoint?: string
      oauth_scopes?: string[]
    }>
    execution_config?: {
      mode?: string
      timeout?: string
      environment?: string
    }
  }
}

export interface WorkflowFile {
  filename: string
  id: string
  path: string
  saved_at: string
}

export interface IntentAnalysis {
  can_fulfill: boolean
  is_automation_request: boolean
  missing_info: string[]
  next_action: string
  required_services: string[]
  confidence_score?: number
}

export interface WorkflowGeneration {
  workflow_cue: string
  workflow_file: WorkflowFile
  generated_at: string
}

export interface PipelineResult {
  phase: 'intent_analysis' | 'workflow_generation' | 'execution_preparation' | 'completed'
  intent_analysis?: IntentAnalysis
  workflow_generation?: WorkflowGeneration
  error?: string
  details?: string
  workflow_id?: string
}

export interface Agent {
  name: string
  type: 'intent_gatherer' | 'intent_analyst' | 'workflow_generator' | 'personal_capabilities' | 'agent_manager'
  status: 'active' | 'inactive' | 'error'
  description: string
  last_execution?: string
}

export interface ConversationMessage {
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
}

export interface WorkflowDiscovery {
  session_id: string
  messages: ConversationMessage[]
  current_step: string
  collected_parameters: Record<string, any>
  status: 'in_progress' | 'completed' | 'error'
}

// API Response Types
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

// Store State Types
export interface AuthState {
  isAuthenticated: boolean
  token: AuthToken | null
  user: User | null
  loading: boolean
  error: string | null
}

export interface WorkflowState {
  workflows: Workflow[]
  currentWorkflow: Workflow | null
  discovery: WorkflowDiscovery | null
  loading: boolean
  error: string | null
}

export interface ServiceState {
  catalog: ServiceCatalog | null
  agents: Agent[]
  loading: boolean
  error: string | null
}

export interface UIState {
  activeTab: 'create' | 'workflows' | 'agents'
  sidebarOpen: boolean
  notifications: Notification[]
}

export interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message: string
  timestamp: string
  read: boolean
}
