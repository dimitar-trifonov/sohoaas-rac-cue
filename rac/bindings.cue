package rac

// =============================================
// 🔹 SOHOAAS PHASE 3 IMPLEMENTATION BINDINGS
// =============================================
// Updated to reflect: In-process deployment, 4-agent architecture, OAuth2 authentication

#Binding: {
    id:   string & !=""
    type: "technology" | "architecture" | "deployment" | "integration"
    
    // Technology bindings (traditional)
    tech?: string
    mappings?: [...{
        source:   string
        target:   string
        strategy: string
    }]
    
    // Architecture bindings (NEW)
    architecture?: {
        pattern: string  // e.g., "microservices", "monolith", "serverless"
        
        components?: [...{
            id:              string & !=""
            name:            string
            description?:    string
            responsibilities: [...string]
            interfaces:      [...string]
            
            // Deployment model (NEW)
            deployment_model?: "in_process" | "microservice" | "serverless"
            
            // Agent type (NEW for SOHOAAS)
            agent_type?: "llm_agent" | "orchestrator" | "service_proxy"
            
            // Capability requirements (abstract)
            capabilities?: [...string]
            
            // Dependencies on other components
            dependsOn?: [...string]
            
            // Communication patterns
            communicatesWith?: [...{
                component: string
                protocol:  string
                pattern:   "sync" | "async" | "event-driven" | "in_process_call"
            }]
        }]
        
        // System-wide patterns
        communicationPatterns?: [...{
            id:          string
            description: string
            path:        [...string]
            protocol:    string
            pattern:     "request-response" | "event-driven" | "streaming"
        }]
        
        // Quality attributes
        qualityAttributes?: {
            scalability?:   string
            availability?:  string
            performance?:   string
            security?:      string
            maintainability?: string
        }
        
        // Deployment topology
        topology?: {
            distribution: "single-node" | "multi-node" | "cloud-native" | "edge"
            networking?:  string
            storage?:     string
        }
    }
    
    // Deployment bindings (concrete implementation)
    deployment?: {
        technology?: string
        framework?:  string
        port?:       int
        environment?: {
            containerization?: string
            orchestration?:    string
            monitoring?:       string
            logging?:          string
        }
        
        // Resource requirements
        resources?: {
            cpu?:    string
            memory?: string
            storage?: string
        }
        
        // Configuration
        config?: {
            [key=string]: string | int | bool
        }
    }
    
    version?: string
    metadata?: {
        createdBy?: string
    }
    
    metadata?: {
        layer: "requirements" | "architecture" | "implementation"
        tags?: [...string]
    }
}

// =============================================
// 🔹 SOHOAAS PHASE 3 ARCHITECTURE BINDING
// =============================================

SOHOAASArchitectureBinding: #Binding & {
    id: "sohoaas_phase3_architecture"
    type: "architecture"
    
    architecture: {
        pattern: "monolith_with_agents"
        
        components: [
            {
                id: "agent_manager"
                name: "Agent Manager"
                description: "Centralized orchestrator with deterministic routing"
                deployment_model: "in_process"
                agent_type: "orchestrator"
                responsibilities: [
                    "Event-driven agent coordination",
                    "Service catalog management",
                    "Deterministic workflow routing",
                    "Authentication middleware"
                ]
                interfaces: ["REST API", "Internal Agent Communication"]
                capabilities: ["service_discovery", "agent_orchestration", "oauth2_validation"]
                communicatesWith: [
                    {component: "personal_capabilities_agent", protocol: "function_call", pattern: "in_process_call"},
                    {component: "intent_gatherer_agent", protocol: "function_call", pattern: "in_process_call"},
                    {component: "intent_analyst_agent", protocol: "function_call", pattern: "in_process_call"},
                    {component: "workflow_generator_agent", protocol: "function_call", pattern: "in_process_call"},
                    {component: "mcp_service", protocol: "HTTP", pattern: "sync"}
                ]
            },
            {
                id: "personal_capabilities_agent"
                name: "Personal Capabilities Agent"
                description: "Service discovery with static service-to-capability mappings"
                deployment_model: "in_process"
                agent_type: "service_proxy"
                responsibilities: [
                    "Google Workspace service catalog discovery",
                    "User capability mapping",
                    "MCP service integration"
                ]
                interfaces: ["Agent Manager API"]
                capabilities: ["service_catalog", "capability_mapping"]
                dependsOn: ["mcp_service"]
            },
            {
                id: "intent_gatherer_agent"
                name: "Intent Gatherer Agent"
                description: "Multi-turn workflow discovery through conversation"
                deployment_model: "in_process"
                agent_type: "llm_agent"
                responsibilities: [
                    "Multi-turn conversation management",
                    "Workflow pattern discovery",
                    "Trigger identification",
                    "Action sequence mapping"
                ]
                interfaces: ["Agent Manager API", "Genkit LLM Integration"]
                capabilities: ["conversation_management", "pattern_discovery", "llm_integration"]
                dependsOn: ["genkit_service"]
            },
            {
                id: "intent_analyst_agent"
                name: "Intent Analyst Agent"
                description: "Simplified intent analysis with 5 PoC parameters"
                deployment_model: "in_process"
                agent_type: "llm_agent"
                responsibilities: [
                    "Intent validation with 5 parameters",
                    "Service requirement analysis",
                    "Google Workspace service validation",
                    "Missing information identification"
                ]
                interfaces: ["Agent Manager API", "Genkit LLM Integration"]
                capabilities: ["intent_analysis", "service_validation", "parameter_extraction"]
                dependsOn: ["genkit_service", "service_catalog"]
            },
            {
                id: "workflow_generator_agent"
                name: "Workflow Generator Agent"
                description: "Deterministic CUE workflow generation with steps-based execution"
                deployment_model: "in_process"
                agent_type: "llm_agent"
                responsibilities: [
                    "Deterministic workflow generation",
                    "Sequential step creation with dependencies",
                    "User parameter design with validation",
                    "Complete CUE file generation"
                ]
                interfaces: ["Agent Manager API", "Genkit LLM Integration"]
                capabilities: ["workflow_generation", "cue_generation", "step_orchestration"]
                dependsOn: ["genkit_service", "service_catalog"]
            },
            {
                id: "genkit_service"
                name: "Google Genkit Service"
                description: "LLM integration with Google GenAI plugin"
                deployment_model: "in_process"
                responsibilities: [
                    "LLM model management",
                    "Prompt loading and execution",
                    "Reflection server for debugging"
                ]
                interfaces: ["Genkit API", "Reflection Server"]
                capabilities: ["llm_execution", "prompt_management"]
            },
            {
                id: "mcp_service"
                name: "MCP Authentication Service"
                description: "OAuth2 Google Workspace API proxy"
                deployment_model: "microservice"
                responsibilities: [
                    "OAuth2 authentication flow",
                    "Google Workspace API proxy",
                    "Token validation and refresh"
                ]
                interfaces: ["REST API", "OAuth2 Endpoints"]
                capabilities: ["oauth2_flow", "api_proxy", "token_management"]
            }
        ]
        
        communicationPatterns: [
            {
                id: "agent_orchestration"
                description: "Agent Manager coordinates all agents via in-process calls"
                path: ["agent_manager", "personal_capabilities_agent", "intent_gatherer_agent", "intent_analyst_agent", "workflow_generator_agent"]
                protocol: "function_call"
                pattern: "event-driven"
            },
            {
                id: "mcp_integration"
                description: "External OAuth2 service integration"
                path: ["agent_manager", "mcp_service"]
                protocol: "HTTP"
                pattern: "request-response"
            },
            {
                id: "llm_integration"
                description: "Genkit LLM service integration"
                path: ["intent_gatherer_agent", "genkit_service"]
                protocol: "function_call"
                pattern: "request-response"
            }
        ]
        
        qualityAttributes: {
            scalability: "Single instance with in-process scaling"
            availability: "Single point of failure acceptable for PoC"
            performance: "Optimized for development speed over performance"
            security: "OAuth2 authentication with Google Workspace"
            maintainability: "Monolithic for simplified debugging and deployment"
        }
        
        topology: {
            distribution: "single-node"
            networking: "HTTP REST API + in-process function calls"
            storage: "Stateless with external OAuth2 token management"
        }
    }
    
    metadata: {
        layer: "architecture"
        tags: ["sohoaas", "phase3", "in_process", "4_agents", "oauth2", "monolith"]
    }
}
