package agents

import "../schemas.cue"

// =============================================
// 🔹 WORKFLOW EDITOR AGENT
// =============================================

WorkflowEditorAgent: {
    version: "1.0.0"
    type: "autonomous_agent"
    poc_status: "optional" // Nice-to-have for PoC - can defer complex editing features
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        {
            id: "workflow_validation"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "workflow_name", type: "string", required: true },
                { name: "validation_results", type: "array" },
                { name: "syntax_errors", type: "array" },
                { name: "missing_bindings", type: "array" },
                { name: "workflow_ready", type: "boolean" }
            ]
            metadata: {
                tags: ["validation", "format", "completeness", "mcp"]
            }
        },
        {
            id: "workflow_editing"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "workflow_name", type: "string", required: true },
                { name: "edit_requests", type: "array" },
                { name: "modified_workflow", type: "object" },
                { name: "edit_history", type: "array" },
                { name: "status", type: "string" } // "editing", "validating", "complete"
            ]
            metadata: {
                tags: ["editing", "modification", "refinement"]
            }
        }
    ]
    
    events: [
        {
            id: "workflow_generated"
            version: "1.0"
            type: "input"
            description: "New workflow received for validation"
            triggers: ["validate_workflow"]
        },
        {
            id: "edit_requested"
            version: "1.0"
            type: "input"
            description: "User requests workflow modifications"
            triggers: ["edit_workflow"]
        },
        {
            id: "workflow_validated"
            version: "1.0"
            type: "output"
            description: "Workflow validated and ready for execution"
            data_schema: {
                workflow_name: "string"
                validation_status: "string"
                final_workflow: "object"
            }
        }
    ]
    
    logic: [
        {
            id: "validate_workflow"
            type: "genkit_flow"
            description: "Validate generated workflow for correctness and completeness"
            input_schema: {
                workflow_name: "string"
                rac_definition: "object"
                mcp_bindings: "object"
            }
            steps: [
                {
                    name: "validate_rac_syntax"
                    description: "Check RaC definition syntax and structure"
                    action: "rac.validate_syntax"
                },
                {
                    name: "validate_mcp_bindings"
                    description: "Verify MCP service bindings are correct"
                    action: "mcp.validate_bindings"
                },
                {
                    name: "check_workflow_completeness"
                    description: "Ensure all required parameters are present"
                    action: "workflow.check_completeness"
                },
                {
                    name: "test_workflow_logic"
                    description: "Run basic logic tests on workflow"
                    action: "workflow.test_logic"
                }
            ]
            output_event: "workflow_validated"
        },
        {
            id: "edit_workflow"
            type: "genkit_flow"
            description: "Apply user-requested edits to workflow"
            input_schema: {
                workflow: "object"
                edit_requests: ["string"]
                user_feedback: "string"
            }
            steps: [
                {
                    name: "parse_edit_requests"
                    description: "Understand what changes user wants"
                    action: "llm.parse_edits"
                    llm_config: {
                        model: "gpt-4"
                        temperature: 0.2
                        system_prompt: "Parse user edit requests and translate to specific workflow modifications"
                    }
                },
                {
                    name: "apply_modifications"
                    description: "Apply changes to workflow structure"
                    action: "workflow.apply_edits"
                },
                {
                    name: "revalidate_workflow"
                    description: "Validate modified workflow"
                    action: "workflow.validate"
                },
                {
                    name: "track_changes"
                    description: "Record edit history for user review"
                    action: "workflow.track_changes"
                }
            ]
            output_event: "workflow_validated"
        }
    ]
    
    // LAYER 2: ARCHITECTURE (How)
    architecture: {
        type: "microservice"
        components: {
            rac_validator: {
                type: "rule_engine"
                responsibilities: ["syntax_validation", "structure_checking"]
            }
            workflow_editor: {
                type: "llm_agent"
                responsibilities: ["edit_interpretation", "modification_application"]
            }
            mcp_validator: {
                type: "integration_service"
                responsibilities: ["binding_validation", "service_testing"]
            }
        }
        communication: {
            input: ["event_bus"]
            output: ["event_bus"]
            external: ["openai_api", "mcp_services", "rac_validator"]
        }
        quality_attributes: {
            reliability: "high"
            accuracy: "high"
            usability: "high"
        }
    }
    
    // LAYER 3: IMPLEMENTATION (With What)
    bindings: [
        {
            type: "deployment"
            technology: "golang"
            framework: "genkit"
            deployment: {
                service_name: "workflow-editor-agent"
                port: 8086
                resources: {
                    cpu: "1.0"
                    memory: "1Gi"
                }
                config: {
                    openai_api_key: "${OPENAI_API_KEY}"
                    rac_validator_url: "${RAC_VALIDATOR_URL}"
                    mcp_registry_url: "${MCP_REGISTRY_URL}"
                    max_edit_iterations: "5"
                }
            }
        }
    ]
    
    tests: [
        {
            id: "workflow_syntax_validation"
            type: "unit"
            description: "Validate RaC workflow syntax"
            input: {
                rac_definition: {
                    states: [
                        { id: "email_draft", type: "object" }
                    ]
                    events: [
                        { id: "draft_email", triggers: ["create_draft"] }
                    ]
                }
            }
            expected: {
                validation_status: "valid"
                syntax_errors: []
            }
        },
        {
            id: "workflow_edit_application"
            type: "integration"
            description: "Apply user edit requests to workflow"
            input: {
                workflow: { /* existing workflow */ }
                edit_requests: ["Change email subject to 'Urgent: Proposal Follow-up'"]
                user_feedback: "The subject line should be more urgent"
            }
            expected: {
                modified_workflow: {
                    /* workflow with updated subject */
                }
                edit_history: [
                    {
                        change: "Updated email subject"
                        timestamp: "2024-01-01T10:00:00Z"
                    }
                ]
            }
        }
    ]
    
    ui: [
        {
            id: "workflow_editor"
            type: "component"
            description: "Interactive workflow editing interface"
            props: {
                workflow: "workflow_validation"
                validation_errors: "array"
                edit_suggestions: "array"
            }
            events: ["edit_applied", "workflow_approved"]
        },
        {
            id: "validation_results"
            type: "component"
            description: "Show workflow validation results"
            props: {
                validation: "workflow_validation"
                errors: "array"
                warnings: "array"
            }
            events: ["fix_requested", "validation_accepted"]
        }
    ]
}
