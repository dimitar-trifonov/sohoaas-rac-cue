package agents

// =============================================
// 🔹 SOHOAAS MULTI-AGENT SYSTEM
// =============================================

// Import all agent definitions
import (
    "agent_manager.cue"
    "personal_capabilities.cue"
    "intent_gatherer.cue"
    "story_coaching.cue"
    "intent_analyst.cue"
    "workflow_generator.cue"
    "workflow_editor.cue"
)

// Complete SOHOAAS Multi-Agent System
SOHOAASAgentSystem: {
    version: "2.0.0"
    architecture: "event_driven_microservices"
    
    // System-wide configuration
    system_config: {
        event_bus_type: "in_memory" // Can be upgraded to Redis/NATS
        agent_communication: "async_events"
        orchestration: "deterministic"
        llm_provider: "openai"
        framework: "genkit"
    }
    
    // Agent Manager - Deterministic Orchestrator
    orchestrator: AgentManager
    
    // Autonomous Agents
    agents: {
        personal_capabilities: PersonalCapabilitiesAgent
        intent_gatherer: IntentGathererAgent
        story_coaching: StoryCoachingAgent
        intent_analyst: IntentAnalystAgent
        workflow_generator: WorkflowGeneratorAgent
        workflow_editor: WorkflowEditorAgent
    }
    
    // System-wide event definitions
    system_events: [
        // User lifecycle events
        {
            id: "user_authenticated"
            source: "auth_service"
            targets: ["personal_capabilities"]
            priority: "high"
        },
        {
            id: "user_message_received"
            source: "frontend"
            targets: ["intent_gatherer", "story_coaching"]
            routing_logic: "state_based"
        },
        
        // Agent coordination events
        {
            id: "capabilities_discovered"
            source: "personal_capabilities"
            targets: ["story_coaching"]
        },
        {
            id: "story_analyzed"
            source: "story_coaching"
            targets: ["intent_gatherer"]
        },
        {
            id: "intent_extracted"
            source: "intent_gatherer"
            targets: ["intent_analyst"]
        },
        {
            id: "intent_validated"
            source: "intent_analyst"
            targets: ["workflow_generator"]
        },
        {
            id: "workflow_generated"
            source: "workflow_generator"
            targets: ["workflow_editor"]
        },
        {
            id: "workflow_validated"
            source: "workflow_editor"
            targets: ["execution_engine"]
        }
    ]
    
    // Global state management
    shared_states: {
        user_session: {
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "session_id", type: "string", required: true },
                { name: "current_state", type: "string" },
                { name: "agent_context", type: "object" }
            ]
        }
        
        workflow_context: {
            fields: [
                { name: "workflow_id", type: "string", required: true },
                { name: "user_id", type: "string", required: true },
                { name: "current_agent", type: "string" },
                { name: "progress", type: "object" }
            ]
        }
    }
    
    // System-wide deployment configuration
    deployment: {
        platform: "kubernetes"
        namespace: "sohoaas-agents"
        
        services: {
            agent_manager: {
                replicas: 1
                port: 8080
                resources: { cpu: "0.5", memory: "512Mi" }
            }
            personal_capabilities: {
                replicas: 2
                port: 8081
                resources: { cpu: "0.5", memory: "512Mi" }
            }
            intent_gatherer: {
                replicas: 3
                port: 8082
                resources: { cpu: "1.0", memory: "1Gi" }
            }
            story_coaching: {
                replicas: 2
                port: 8083
                resources: { cpu: "0.8", memory: "1Gi" }
            }
            intent_analyst: {
                replicas: 2
                port: 8084
                resources: { cpu: "1.0", memory: "1Gi" }
            }
            workflow_generator: {
                replicas: 1
                port: 8085
                resources: { cpu: "1.5", memory: "2Gi" }
            }
            workflow_editor: {
                replicas: 2
                port: 8086
                resources: { cpu: "1.0", memory: "1Gi" }
            }
        }
        
        networking: {
            service_mesh: "istio"
            load_balancer: "nginx"
            api_gateway: "kong"
        }
    }
    
    // System-wide monitoring and observability
    observability: {
        metrics: {
            agent_performance: ["response_time", "success_rate", "error_rate"]
            system_health: ["event_throughput", "queue_depth", "resource_usage"]
        }
        
        logging: {
            level: "info"
            structured: true
            correlation_id: "trace_id"
        }
        
        tracing: {
            enabled: true
            sampler: "probabilistic"
            sample_rate: 0.1
        }
    }
    
    // Integration tests for the complete system
    system_tests: [
        {
            id: "end_to_end_workflow"
            description: "Complete user journey from authentication to workflow execution"
            scenario: "oauth_to_execution"
            steps: [
                { agent: "auth_service", action: "authenticate_user" },
                { agent: "personal_capabilities", action: "discover_capabilities" },
                { agent: "story_coaching", action: "analyze_daily_story" },
                { agent: "intent_gatherer", action: "extract_intent" },
                { agent: "intent_analyst", action: "validate_intent" },
                { agent: "workflow_generator", action: "generate_workflow" },
                { agent: "workflow_editor", action: "validate_workflow" }
            ]
            expected_duration: "< 30s"
        }
    ]
}
