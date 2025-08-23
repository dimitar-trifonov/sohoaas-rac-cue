package agents

import "../schemas.cue"

// =============================================
// 🔹 INTENT ANALYST AGENT - PHASE 3 SIMPLIFIED
// =============================================
// Updated: 5 PoC parameters instead of 9, Google Workspace validation

IntentAnalystAgent: {
    version: "1.0.0"
    type: "autonomous_agent"
    poc_status: "essential" // Core LLM value: Simplified 5-parameter analysis for PoC
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        {
            id: "intent_analysis"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "workflow_intent", type: "object", required: true }, // From workflow discovery
                { name: "is_automation_request", type: "boolean", required: true }, // PoC Parameter 1
                { name: "required_services", type: "array", required: true }, // PoC Parameter 2 - Google Workspace only
                { name: "can_fulfill", type: "boolean", required: true }, // PoC Parameter 3
                { name: "missing_info", type: "array", required: true }, // PoC Parameter 4
                { name: "next_action", type: "string", required: true }, // PoC Parameter 5 - enum
                { name: "service_validation", type: "object" }, // Validated against service catalog
                { name: "confidence", type: "string" } // "high", "medium", "low"
            ]
            metadata: {
                tags: ["intent", "analysis", "poc_simplified", "5_parameters", "google_workspace"]
            }
        }
    ]
    
    events: [
        {
            id: "workflow_intent_discovered"
            version: "1.0"
            type: "input"
            description: "Complete workflow pattern received from intent gatherer"
            triggers: ["analyze_workflow_intent"]
        },
        {
            id: "intent_analysis_complete"
            version: "1.0"
            type: "output"
            description: "Intent analysis complete with 5 PoC parameters for workflow generation"
            data_schema: {
                is_automation_request: "boolean"
                required_services: ["gmail", "google_docs", "google_drive", "google_calendar", "google_sheets"]
                can_fulfill: "boolean"
                missing_info: ["string"]
                next_action: "generate_workflow" | "request_clarification" | "reject_request"
                explanation: "string"
                service_validation: "object"
            }
        }
    ]
    
    logic: [
        {
            id: "analyze_workflow_intent"
            type: "genkit_flow"
            description: "Validate and analyze discovered workflow patterns for feasibility and completeness"
            input_schema: {
                workflow_pattern: "string"
                trigger_conditions: "object"
                action_sequence: ["string"]
                data_requirements: ["string"]
                user_capabilities: ["string"]
            }
            steps: [
                {
                    name: "validate_and_extract"
                    description: "Single-step workflow validation and parameter extraction for PoC simplicity"
                    action: "llm.validate_workflow"
                    llm_config: {
                        model: "gpt-4"
                        temperature: 0.1
                        system_prompt: "Validate workflow feasibility and extract parameters needed for RaC generation. Focus on: Can we build this? What services/parameters do we need?"
                    }
                }
            ]
            output_event: "workflow_intent_validated"
        }
    ]
    
    // LAYER 2: ARCHITECTURE (How)
    architecture: {
        type: "microservice"
        components: {
            workflow_validator: {
                type: "llm_agent"
                responsibilities: ["feasibility_check", "parameter_extraction", "simple_validation"]
            }
        }
        communication: {
            input: ["event_bus"]
            output: ["event_bus"]
            external: ["openai_api", "mcp_registry"]
        }
        quality_attributes: {
            accuracy: "high"
            completeness: "high"
            reliability: "high"
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
                }
            }
        }
    ]
    
    tests: [
        {
            id: "email_workflow_analysis"
            type: "unit"
            description: "Analyze simple email sending workflow"
            input: {
                raw_intent: "Send follow-up email to john@acme.com about proposal"
                intent_type: "email"
                entities: {
                    recipient: "john@acme.com"
                    topic: "proposal follow-up"
                }
            }
            expected: {
                workflow_type: "email_followup"
                action_sequence: [
                    {
                        action: "send_message"
                        parameters: {
                            to: "john@acme.com"
                            subject: "Follow-up on proposal"
                            template: "proposal_followup"
                        }
                    }
                ]
                complexity: "simple"
                validation_status: "complete"
            }
        },
        {
            id: "complex_meeting_prep_analysis"
            type: "unit"
            description: "Analyze complex multi-step meeting preparation"
            input: {
                raw_intent: "Prepare for client meeting tomorrow - create agenda, send invite, gather last quarter's reports"
                intent_type: "complex"
            }
            expected: {
                workflow_type: "meeting_preparation"
                complexity: "complex"
                action_sequence: [
                    { action: "create_document", type: "agenda" },
                    { action: "schedule_meeting", type: "calendar_invite" },
                    { action: "gather_documents", type: "report_collection" }
                ]
            }
        }
    ]
    
    ui: [
        {
            id: "intent_breakdown"
            type: "component"
            description: "Show analyzed workflow steps to user"
            props: {
                workflow_intent: "workflow_intent"
                action_sequence: "array"
                missing_parameters: "array"
            }
            events: ["parameters_provided", "workflow_approved"]
        }
    ]
}
