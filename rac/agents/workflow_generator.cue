package agents

import "../schemas.cue"

// =============================================
// 🔹 WORKFLOW GENERATOR AGENT - PHASE 3 DETERMINISTIC
// =============================================
// Updated: Deterministic CUE workflow generation, steps-based execution, simplified output

WorkflowGeneratorAgent: {
    version: "1.0.0"
    type: "autonomous_agent"
    poc_status: "essential" // Core value: Intent → Deterministic CUE workflow with steps-based execution
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        {
            id: "deterministic_workflow"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "workflow_name", type: "string", required: true },
                { name: "workflow_steps", type: "array", required: true }, // Sequential steps with dependencies
                { name: "user_parameters", type: "object", required: true }, // Required parameters with validation
                { name: "service_bindings", type: "object", required: true }, // MCP service connections
                { name: "data_flow", type: "object" }, // Step outputs feeding into inputs
                { name: "source_intent", type: "object", required: true }, // Reference to user_intent state
                { name: "json_specification", type: "object", required: true }, // Structured JSON workflow (LLM output)
                // Note: cue_specification generated deterministically from JSON by system
                { name: "validation_results", type: "object" }, // Schema and service validation
                { name: "execution_ready", type: "boolean" }, // Ready for immediate execution
                { name: "status", type: "string" } // "generated", "validated", "ready", "error"
            ]
            metadata: {
                tags: ["workflow", "deterministic", "cue_file", "executable", "steps_based"]
            }
        }
    ]
    
    events: [
        {
            id: "intent_analysis_complete"
            version: "1.0"
            type: "input"
            description: "Intent analysis with 5 PoC parameters ready for workflow generation"
            triggers: ["generate_deterministic_workflow"]
        },
        {
            id: "deterministic_workflow_generated"
            version: "1.0"
            type: "output"
            description: "Deterministic CUE workflow generated with steps-based execution"
            data_schema: {
                workflow_name: "string"
                cue_specification: "string" // Complete CUE file content
                workflow_steps: [{
                    id: "string"
                    service: "string"
                    action: "string"
                    dependencies: ["string"]
                }]
                user_parameters: {
                    "[param_name]": {
                        type: "string"
                        required: "boolean"
                        validation: "string"
                        prompt: "string"
                    }
                }
                service_bindings: {
                    "[service_name]": {
                        oauth_scopes: ["string"]
                        mcp_connection: "string"
                    }
                }
                execution_ready: "boolean"
                execution_steps: "array"
                required_parameters: "object"
                service_dependencies: "array"
                validation_schema: "object"
            }
        }
    ]
    
    logic: [
        {
            id: "generate_workflow"
            type: "genkit_flow"
            description: "Generate complete deterministic workflow CUE file from validated workflow pattern"
            input_schema: {
                workflow_pattern: "string"
                validated_triggers: "object"
                validated_actions: "array"
                required_services: "array"
                workflow_parameters: "object"
            }
            steps: [
                {
                    name: "generate_deterministic_workflow"
                    description: "Generate complete deterministic workflow CUE file with steps, parameters, and dependencies"
                    action: "llm.generate_workflow_cue"
                    llm_config: {
                        model: "gpt-4"
                        temperature: 0.1
                        system_prompt: "Generate a complete deterministic workflow CUE file with: 1) Sequential steps with service calls, 2) User parameters with validation, 3) Step dependencies and data flow. Focus on executable, deterministic automation."
                    }
                },
                {
                    name: "validate_workflow_schema"
                    description: "Validate generated workflow against CUE schema"
                    action: "cue.validate_schema"
                },
                {
                    name: "verify_service_bindings"
                    description: "Verify all required services are properly bound"
                    action: "service.verify_bindings"
                }
            ]
            output_event: "workflow_generated"
        }
    ]
    
    // LAYER 2: ARCHITECTURE (How)
    architecture: {
        type: "microservice"
        components: {
            workflow_generator: {
                type: "llm_agent"
                responsibilities: ["deterministic_workflow_generation", "step_sequencing", "parameter_extraction"]
            }
            schema_validator: {
                type: "cue_engine"
                responsibilities: ["workflow_validation", "schema_compliance", "dependency_verification"]
            }
            service_binder: {
                type: "mcp_mapper"
                responsibilities: ["service_binding_validation", "oauth_scope_verification"]
            }
        }
        communication: {
            input: ["event_bus"]
            output: ["event_bus"]
            external: ["openai_api", "mcp_registry", "rac_templates"]
        }
        quality_attributes: {
            correctness: "high"
            completeness: "high"
            maintainability: "high"
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
            id: "email_workflow_generation"
            type: "integration"
            description: "Generate complete email workflow from validated intent"
            input: {
                workflow_type: "email_followup"
                action_sequence: [
                    {
                        action: "send_email"
                        parameters: {
                            to: "john@acme.com"
                            subject: "Follow-up on proposal"
                            template: "proposal_followup"
                        }
                    }
                ]
                parameters: {
                    recipient: "john@acme.com"
                    topic: "proposal follow-up"
                }
            }
            expected: {
                workflow_name: "Proposal Follow-up Email"
                rac_definition: {
                    states: [
                        { id: "email_draft", type: "object" },
                        { id: "email_sent", type: "object" }
                    ]
                    events: [
                        { id: "draft_email", triggers: ["create_draft"] },
                        { id: "send_email", triggers: ["send_message"] }
                    ]
                }
                mcp_bindings: {
                    gmail: {
                        send_email: {
                            to: "john@acme.com"
                            subject: "Follow-up on proposal"
                        }
                    }
                }
            }
        }
    ]
    
    ui: [
        {
            id: "workflow_preview"
            type: "component"
            description: "Show generated workflow structure to user"
            props: {
                workflow: "generated_workflow"
                rac_definition: "object"
                execution_plan: "object"
            }
            events: ["workflow_approved", "workflow_needs_editing"]
        }
    ]
}
