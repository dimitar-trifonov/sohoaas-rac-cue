package agents

// =============================================
// 🔹 AGENT MANAGER - PHASE 3 IMPLEMENTATION
// =============================================
// Updated: Centralized service catalog, 4-agent coordination, OAuth2 integration

// Agent Manager is fully deterministic and event-driven
// It coordinates between autonomous agents without making LLM decisions
// NEW: Service catalog management and authentication middleware
AgentManager: {
    version: "1.0.0"
    type: "deterministic_orchestrator"
    
    // Event routing rules - deterministic mapping
    event_routing: {
        // User authentication triggers capability discovery
        "user_authenticated": ["personal_capabilities"]
        
        // User messages route to workflow discovery (multi-turn conversation)
        "user_message_received": {
            if_state: "workflow_discovery": ["intent_gatherer"] // Multi-turn workflow discovery
            if_state: "intent_analysis": ["intent_analyst"]
            if_state: "deterministic_workflow": ["workflow_generator"]
        }
        
        // Agent completion events trigger next agents (4-agent pipeline)
        "capabilities_discovered": [] // Go directly to workflow discovery
        "workflow_intent_discovered": ["intent_analyst"] // Complete workflow patterns to analyst
        "intent_analysis_complete": ["workflow_generator"] // 5 PoC parameters to generator
        "deterministic_workflow_generated": ["cue_generator"] // JSON workflow to CUE conversion
        "cue_conversion_complete": ["workflow_validator"] // CUE to validation
        "validation_complete": ["workflow_executor"] // Validated workflow to execution
        "execution_completed": [] // End of pipeline
        
        // Conditional routing for workflow discovery
        "continue_discovery": ["intent_gatherer"] // Loop back for incomplete discovery
        "workflow_validation_failed": ["intent_gatherer"] // Back to discovery if validation fails
    }
    
    // State transition rules - deterministic
    state_transitions: {
        "user_auth" -> "personal_capabilities": {
            trigger: "user_authenticated"
            condition: "oauth_tokens_valid"
        }
        
        "personal_capabilities" -> "workflow_discovery": {
            trigger: "capabilities_discovered"
            condition: "service_catalog_ready"
        }
        
        "workflow_discovery" -> "intent_analysis": {
            trigger: "workflow_intent_discovered"
            condition: "confidence > 0.8"
        }
        
        "workflow_intent" -> "generated_workflow": {
            trigger: "intent_validated"
            condition: "parameters_complete"
        }
        
        "generated_workflow" -> "workflow_validation": {
            trigger: "workflow_generated"
            condition: "workflow_structure_valid"
        }
        
        "workflow_validation" -> "intent_confirmation": {
            trigger: "workflow_validated"
            condition: "validation_passed"
        }
    }
    
    // Agent health monitoring
    agent_monitoring: {
        health_check_interval: "30s"
        timeout_thresholds: {
            personal_capabilities: "60s"
            intent_gatherer: "30s"
            story_coaching: "45s"
            intent_analyst: "30s"
            workflow_generator: "120s"
            workflow_editor: "60s"
        }
    }
    
    // Event bus configuration
    event_bus: {
        type: "in_memory" // For PoC, can be upgraded to Redis/NATS
        buffer_size: 1000
        retry_policy: {
            max_retries: 3
            backoff: "exponential"
            base_delay: "1s"
        }
    }
}
