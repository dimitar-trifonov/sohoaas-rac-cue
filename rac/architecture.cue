package rac

// =============================================
// 🔹 SOHOAAS ARCHITECTURE BINDINGS
// =============================================

// LAYER 2: ARCHITECTURE (How the system should be structured)
SOHOAASArchitecture: #Binding & {
    id: "sohoaas_microservices_architecture"
    type: "architecture"
    architecture: {
        pattern: "intent-to-workflow-automation"
        
        components: [
            {
                id: "intent_capture"
                name: "Intent Capture Service"
                description: "Captures and validates user intent through conversational interface"
                responsibilities: [
                    "natural_language_processing",
                    "conversation_management",
                    "intent_validation",
                    "context_gathering"
                ]
                interfaces: ["chat_ui", "intent_processor"]
                capabilities: [
                    "real_time_communication",
                    "session_management",
                    "input_validation"
                ]
                communicatesWith: [{
                    component: "intent_processor"
                    protocol: "REST"
                    pattern: "sync"
                }]
            },
            {
                id: "intent_processor"
                name: "Intent Analysis Engine"
                description: "LLM-powered service that transforms intent into structured workflows"
                responsibilities: [
                    "llm_integration",
                    "intent_analysis",
                    "workflow_generation",
                    "tool_discovery"
                ]
                interfaces: ["intent_capture", "workflow_engine", "mcp_server"]
                capabilities: [
                    "language_understanding",
                    "workflow_synthesis",
                    "cue_generation"
                ]
                dependsOn: ["intent_capture", "workflow_engine"]
                communicatesWith: [
                    {
                        component: "workflow_engine"
                        protocol: "REST"
                        pattern: "async"
                    },
                    {
                        component: "mcp_server"
                        protocol: "MCP"
                        pattern: "sync"
                    }
                ]
            },
            {
                id: "workflow_engine"
                name: "Dynamic Workflow Engine"
                description: "Executes generated workflows using available tools"
                responsibilities: [
                    "workflow_execution",
                    "tool_orchestration",
                    "progress_tracking",
                    "error_handling"
                ]
                interfaces: ["intent_processor", "mcp_server"]
                capabilities: [
                    "cue_parsing",
                    "parallel_execution",
                    "failure_recovery"
                ]
                communicatesWith: [{
                    component: "mcp_server"
                    protocol: "MCP"
                    pattern: "sync"
                }]
            }
        ]
        
        communicationPatterns: [{
            id: "intent_to_execution_flow"
            description: "End-to-end flow from user intent to workflow execution"
            path: ["intent_capture", "intent_processor", "workflow_engine"]
            protocol: "REST + MCP"
            pattern: "request-response"
        }]
        
        qualityAttributes: {
            scalability: "horizontal scaling per component"
            availability: "99.9% uptime with graceful degradation"
            performance: "< 5 minutes from intent to execution"
            security: "OAuth2 + API key authentication"
            maintainability: "loosely coupled, independently deployable"
        }
        
        topology: {
            distribution: "multi-node"
            networking: "service mesh with load balancing"
            storage: "distributed state with workflow persistence"
        }
    }
    metadata: {
        layer: "architecture"
        tags: ["microservices", "intent-driven", "llm-powered"]
    }
}
