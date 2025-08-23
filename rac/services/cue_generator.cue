package services 

import "../schemas.cue"

// =============================================
// 🔹 CUE GENERATOR SERVICE - JSON→CUE CONVERSION
// =============================================
// Purpose: Deterministic conversion of JSON workflow specifications to executable CUE format

CueGeneratorService: {
    version: "1.0.0"
    type: "deterministic_service"
    poc_status: "essential" // Core value: JSON→CUE conversion with schema validation
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        {
            id: "json_to_cue_conversion"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "workflow_id", type: "string", required: true },
                { name: "source_json", type: "object", required: true }, // JSON workflow specification
                { name: "target_cue", type: "string", required: true }, // Generated CUE content
                { name: "conversion_rules", type: "object", required: true }, // Applied transformation rules
                { name: "schema_validation", type: "object" }, // CUE schema compliance results
                { name: "conversion_metadata", type: "object" }, // Transformation details
                { name: "status", type: "string" } // "converting", "validated", "ready", "error"
            ]
            metadata: {
                tags: ["conversion", "deterministic", "json_to_cue", "schema"]
            }
        }
    ]
    
    // LAYER 2: EVENTS (When things happen)
    events: [
        {
            id: "json_workflow_received"
            version: "1.0"
            type: "input"
            description: "JSON workflow specification received from workflow generator"
            triggers: ["convert_to_cue"]
            data_schema: {
                workflow_json: "object"
                user_id: "string"
                workflow_name: "string"
            }
        },
        {
            id: "cue_conversion_complete"
            version: "1.0"
            type: "output"
            description: "JSON successfully converted to valid CUE format"
            triggers: ["workflow_validator"]
            data_schema: {
                cue_content: "string"
                validation_results: "object"
                conversion_metadata: "object"
            }
        },
        {
            id: "conversion_failed"
            version: "1.0"
            type: "error"
            description: "JSON→CUE conversion failed due to invalid structure or schema"
            data_schema: {
                error_details: "object"
                validation_errors: "array"
            }
        }
    ]
    
    // LAYER 3: LOGIC (How it works)
    logic: [
        {
            id: "convert_to_cue"
            type: "deterministic_conversion"
            description: "Convert JSON workflow to CUE format with schema validation"
            steps: [
                {
                    name: "validate_json_structure"
                    description: "Validate JSON against /rac/schemas/workflow_json_schema.json"
                    action: "schema.validate_json_workflow"
                    schema_path: "/rac/schemas/workflow_json_schema.json"
                    validation: "Must pass strict JSON schema validation before conversion"
                },
                {
                    name: "convert_json_steps_to_cue"
                    description: "Transform JSON workflow steps to CUE step format"
                    action: "converter.json_steps_to_cue"
                    rules: [
                        "step.action → CUE action field with MCP dot notation",
                        "step.inputs → CUE inputs with parameter references",
                        "step.outputs → CUE outputs with variable bindings",
                        "${user.param} → CUE user parameter syntax"
                    ]
                },
                {
                    name: "convert_user_parameters_to_cue"
                    description: "Transform JSON user parameters to CUE parameter definitions"
                    action: "converter.json_params_to_cue"
                    rules: [
                        "parameter.type → CUE type constraint",
                        "parameter.required → CUE required field",
                        "parameter.validation → CUE validation rules"
                    ]
                },
                {
                    name: "convert_service_bindings_to_cue"
                    description: "Transform JSON service bindings to CUE service configuration"
                    action: "converter.json_services_to_cue"
                    rules: [
                        "service.oauth_scopes → CUE auth.scopes array",
                        "service.service → CUE service binding name"
                    ]
                },
                {
                    name: "validate_cue_syntax"
                    description: "Verify generated CUE is syntactically correct"
                    action: "cue.validate_syntax"
                },
                {
                    name: "validate_mcp_functions"
                    description: "Validate workflow steps use valid MCP service functions"
                    action: "mcp.validate_functions"
                    validation: "Each step.action must match existing MCP function"
                },
                {
                    name: "validate_mcp_parameters"
                    description: "Validate step parameters against MCP function schemas"
                    action: "mcp.validate_parameters"
                    validation: "Parameters must match MCP function input schema requirements"
                },
                {
                    name: "validate_service_bindings"
                    description: "Verify service bindings and OAuth configurations"
                    action: "service.verify_bindings"
                    validation: "Service bindings must have correct OAuth scopes and endpoints"
                },
                {
                    name: "validate_schema_compliance"
                    description: "Ensure CUE matches #DeterministicWorkflow schema"
                    action: "schema.validate_cue_workflow"
                    schema_path: "/rac/schemas/deterministic_workflow.cue"
                }
            ]
            output: {
                state: "json_to_cue_conversion"
                fields: ["target_cue", "conversion_rules", "schema_validation", "status"]
            }
        }
    ]
    
    // LAYER 4: TESTS (Validation)
    tests: [
        {
            id: "json_to_cue_conversion_test"
            description: "Test JSON workflow conversion to valid CUE"
            input: {
                workflow_json: {
                    workflow_name: "Test Email Workflow"
                    description: "Send test email"
                    trigger: { type: "manual" }
                    steps: [
                        {
                            id: "send_message"
                            name: "Send Test Email"
                            service: "gmail"
                            action: "send_message"
                            inputs: {
                                to: "${user.recipient}"
                                subject: "Test Subject"
                                body: "Test message"
                            }
                            outputs: { message_id: "sent_id" }
                            depends_on: []
                        }
                    ]
                    user_parameters: {
                        recipient: {
                            name: "recipient"
                            type: "string"
                            required: true
                            description: "Email recipient"
                            prompt: "Enter recipient email:"
                        }
                    }
                    services: {
                        gmail: {
                            service: "gmail"
                            oauth_scopes: ["https://www.googleapis.com/auth/gmail.send"]
                        }
                    }
                }
            }
            expected: {
                target_cue: "package workflow\n\nworkflow: #DeterministicWorkflow & {\n\tname: \"Test Email Workflow\"\n\t..."
                schema_validation: { valid: true, errors: [] }
                status: "ready"
            }
        }
    ]
}
