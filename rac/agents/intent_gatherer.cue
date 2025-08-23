package agents

import "../schemas.cue"

// =============================================
// 🔹 INTENT GATHERER AGENT - PHASE 3 MULTI-TURN
// =============================================
// Updated: Multi-turn workflow discovery, pattern identification, conversation management

IntentGathererAgent: {
    version: "1.0.0"
    type: "autonomous_agent"
    poc_status: "essential" // Core LLM value: Multi-turn conversation → Complete workflow pattern discovery
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        {
            id: "workflow_discovery"
            type: "object"
            fields: [
                { name: "session_id", type: "string", required: true },
                { name: "user_id", type: "string", required: true },
                { name: "conversation_history", type: "array", required: true }, // Multi-turn conversation
                { name: "discovered_patterns", type: "array" }, // Repetitive processes identified
                { name: "trigger_identification", type: "object" }, // When automation should activate
                { name: "action_sequence", type: "array" }, // Workflow steps mapped
                { name: "data_requirements", type: "object" }, // Services and information needed
                { name: "discovery_phase", type: "string" }, // "pattern", "trigger", "action", "data", "validation"
                { name: "status", type: "string" } // "discovering", "complete", "needs_clarification"
            ]
            metadata: {
                tags: ["discovery", "multi_turn", "workflow_patterns", "conversation"]
            }
        },
        // NOTE: workflow_intent state removed - discovery outputs directly to intent_analysis
    ]
    
    events: [
        {
            id: "user_message_received"
            version: "1.0"
            type: "input"
            description: "User sends message in workflow discovery conversation"
            triggers: ["discover_workflow_intent"]
        },
        {
            id: "workflow_intent_discovered"
            version: "1.0"
            type: "output"
            description: "Complete workflow pattern discovered through conversation"
            data_schema: {
                workflow_pattern: "string"
                trigger_conditions: "object"
                action_sequence: "array"
                data_requirements: "array"
                discovery_complete: "boolean"
            }
        }
    ]
    
    logic: [
        {
            id: "discover_workflow_intent"
            type: "genkit_flow"
            description: "Multi-turn conversation to discover complete workflow automation intent"
            input_schema: {
                user_message: "string"
                conversation_history: ["object"]
                discovery_phase: "string"
                collected_intent: "object"
            }
            steps: [
                {
                    name: "analyze_conversation_phase"
                    description: "Determine current discovery phase and what to ask next"
                    action: "llm.analyze_discovery_phase"
                    llm_config: {
                        model: "gpt-4"
                        temperature: 0.2
                        system_prompt: "Guide workflow discovery conversation through phases: pattern → trigger → actions → data → validation"
                    }
                },
                {
                    name: "extract_workflow_elements"
                    description: "Extract workflow components from user response"
                    action: "llm.extract_workflow_elements"
                    llm_config: {
                        model: "gpt-4"
                        temperature: 0.1
                        system_prompt: "Extract workflow patterns, triggers, actions, and data requirements from user input"
                    }
                },
                {
                    name: "generate_next_questions"
                    description: "Generate contextual questions to continue discovery"
                    action: "llm.generate_discovery_questions"
                    llm_config: {
                        model: "gpt-4"
                        temperature: 0.3
                        system_prompt: "Generate intelligent follow-up questions to complete workflow discovery"
                    }
                },
                {
                    name: "assess_workflow_completeness"
                    description: "Determine if workflow intent is complete enough for analysis"
                    action: "workflow.assess_completeness"
                }
            ]
            output_event: "workflow_intent_discovered"
        }
    ]
    
    // LAYER 2: ARCHITECTURE (How)
    architecture: {
        type: "microservice"
        components: {
            conversation_manager: {
                type: "llm_agent"
                responsibilities: ["discovery_phase_management", "workflow_pattern_recognition", "conversation_guidance"]
            }
            workflow_analyzer: {
                type: "llm_agent"
                responsibilities: ["trigger_identification", "action_sequence_extraction", "data_requirement_analysis"]
            }
        }
        communication: {
            input: ["event_bus"]
            output: ["event_bus"]
            external: ["openai_api"]
        }
        quality_attributes: {
            reliability: "high"
            performance: "high" // Fast response for chat
            accuracy: "high"
        }
    }
    
    // LAYER 3: IMPLEMENTATION (With What)
    bindings: [
        {
            type: "deployment"
            technology: "golang"
            framework: "genkit"
            deployment: {
                service_name: "sohoaas-backend" // Monolithic service for PoC
                deployment_type: "in_process_agent"
                platform: "cloud_run"
                resources: {
                    cpu: "2.0" // Shared across all agents
                    memory: "4Gi" // Shared across all agents
                }
                config: {
                    openai_api_key: "${OPENAI_API_KEY}"
                    firebase_project_id: "${FIREBASE_PROJECT_ID}"
                    max_context_length: "4000"
                }
            }
        }
    ]
    
    tests: [
        {
            id: "email_intent_extraction"
            type: "unit"
            description: "Extract email sending intent"
            input: {
                user_message: "Send an email to john@acme.com about the proposal follow-up"
                available_capabilities: ["send_message", "create_document"]
            }
            expected: {
                intent_type: "email"
                confidence: 0.95
                entities: {
                    recipient: "john@acme.com"
                    topic: "proposal follow-up"
                }
                requires_clarification: false
            }
        },
        {
            id: "complex_workflow_intent"
            type: "unit"
            description: "Detect complex multi-step workflow"
            input: {
                user_message: "Prepare for my meeting with Sarah tomorrow - create agenda, send invite, and follow up on last week's action items"
            }
            expected: {
                intent_type: "complex"
                confidence: 0.85
                requires_clarification: true
            }
        }
    ]
    
    ui: [
        {
            id: "intent_confirmation"
            type: "component"
            description: "Show extracted intent for user confirmation"
            props: {
                intent: "extracted_intent"
                confidence: "number"
            }
            events: ["intent_confirmed", "intent_rejected"]
        }
    ]
}
