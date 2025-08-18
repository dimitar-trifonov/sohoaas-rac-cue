package agents

import "../schemas.cue"

// =============================================
// 🔹 WORKFLOW VALIDATOR SERVICE - PARAMETER & SERVICE VERIFICATION
// =============================================
// Purpose: Validate all parameters, service bindings, and dependencies before execution

WorkflowValidatorService: {
    version: "1.0.0"
    type: "validation_service"
    poc_status: "essential" // Core value: Comprehensive validation before execution
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        {
            id: "workflow_validation"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "workflow_id", type: "string", required: true },
                { name: "cue_specification", type: "string", required: true }, // CUE workflow from generator
                { name: "parameter_validation", type: "object", required: true }, // User parameter checks
                { name: "service_validation", type: "object", required: true }, // MCP service availability
                { name: "dependency_validation", type: "object", required: true }, // Step dependency checks
                { name: "oauth_validation", type: "object", required: true }, // OAuth scope verification
                { name: "validation_errors", type: "array" }, // All validation issues found
                { name: "validation_warnings", type: "array" }, // Non-blocking issues
                { name: "execution_ready", type: "boolean", required: true }, // Ready for execution
                { name: "status", type: "string" } // "validating", "ready", "blocked", "error"
            ]
            metadata: {
                tags: ["validation", "parameters", "services", "dependencies", "oauth"]
            }
        }
    ]
    
    // LAYER 2: EVENTS (When things happen)
    events: [
        {
            id: "cue_workflow_received"
            version: "1.0"
            type: "input"
            description: "CUE workflow specification received from CUE generator"
            triggers: ["validate_workflow"]
            data_schema: {
                cue_content: "string"
                user_id: "string"
                workflow_id: "string"
            }
        },
        {
            id: "validation_complete"
            version: "1.0"
            type: "output"
            description: "Workflow validation completed - ready for execution"
            triggers: ["workflow_executor"]
            data_schema: {
                validation_results: "object"
                execution_ready: "boolean"
                required_parameters: "array"
            }
        },
        {
            id: "validation_blocked"
            version: "1.0"
            type: "error"
            description: "Workflow validation failed - cannot proceed to execution"
            data_schema: {
                validation_errors: "array"
                blocking_issues: "array"
                suggested_fixes: "array"
            }
        }
    ]
    
    // LAYER 3: LOGIC (How it works)
    logic: [
        {
            id: "validate_workflow"
            type: "comprehensive_validation"
            description: "Validate all aspects of workflow before execution"
            steps: [
                {
                    name: "validate_user_parameters"
                    description: "Check all user parameters are properly defined and collectable"
                    action: "validator.check_user_parameters"
                    validations: [
                        "All ${user.param} references have corresponding parameter definitions",
                        "Required parameters have prompts for user collection",
                        "Parameter types are valid (string, number, boolean, array, object)",
                        "Parameter validation rules are syntactically correct"
                    ]
                },
                {
                    name: "validate_service_bindings"
                    description: "Verify all required services are available and properly configured"
                    action: "validator.check_service_availability"
                    validations: [
                        "All referenced services exist in MCP service catalog",
                        "Service actions are available and properly named",
                        "OAuth scopes match service requirements",
                        "Service endpoints are reachable"
                    ]
                },
                {
                    name: "validate_step_dependencies"
                    description: "Check step execution order and data flow dependencies"
                    action: "validator.check_step_dependencies"
                    validations: [
                        "All depends_on references point to valid step IDs",
                        "No circular dependencies in step execution order",
                        "All $(step.outputs.field) references are valid",
                        "Step execution order is deterministic"
                    ]
                },
                {
                    name: "validate_oauth_permissions"
                    description: "Verify user has required OAuth permissions for all services"
                    action: "validator.check_oauth_scopes"
                    validations: [
                        "User has granted all required OAuth scopes",
                        "OAuth tokens are valid and not expired",
                        "Service-specific permissions are sufficient"
                    ]
                },
                {
                    name: "validate_data_flow"
                    description: "Check data flow between steps is consistent"
                    action: "validator.check_data_flow"
                    validations: [
                        "Step outputs match expected input types for dependent steps",
                        "All required inputs have valid sources (parameters or step outputs)",
                        "No missing data dependencies"
                    ]
                }
            ]
            output: {
                state: "workflow_validation"
                fields: ["parameter_validation", "service_validation", "dependency_validation", "oauth_validation", "execution_ready"]
            }
        }
    ]
    
    // LAYER 4: TESTS (Validation)
    tests: [
        {
            id: "workflow_validation_test"
            description: "Test comprehensive workflow validation"
            input: {
                cue_content: """
                package workflow
                
                workflow: #DeterministicWorkflow & {
                    name: "Test Email Workflow"
                    steps: [
                        {
                            id: "send_email"
                            service: "gmail"
                            action: "send_message"
                            inputs: {
                                to: "${user.recipient}"
                                subject: "Test"
                            }
                            outputs: { message_id: "sent_id" }
                        }
                    ]
                    user_parameters: [
                        {
                            name: "recipient"
                            type: "string"
                            required: true
                            prompt: "Enter email:"
                        }
                    ]
                    service_bindings: [
                        {
                            service: "gmail"
                            oauth_scopes: ["https://www.googleapis.com/auth/gmail.send"]
                        }
                    ]
                }
                """
            }
            expected: {
                parameter_validation: { valid: true, missing_parameters: [] }
                service_validation: { valid: true, unavailable_services: [] }
                dependency_validation: { valid: true, circular_dependencies: [] }
                oauth_validation: { valid: true, missing_scopes: [] }
                execution_ready: true
                status: "ready"
            }
        }
    ]
}
