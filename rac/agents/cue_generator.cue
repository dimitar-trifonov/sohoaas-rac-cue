package agents

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
                    name: "apply_conversion_rules"
                    description: "Transform JSON structure to CUE syntax"
                    action: "converter.json_to_cue_transform"
                    rules: [
                        "workflow_name → name field",
                        "steps array → CUE steps structure",
                        "user_parameters → #UserParameter definitions",
                        "service_bindings → #ServiceBinding definitions",
                        "${USER_INPUT:param} → CUE variable syntax",
                        "$(step.outputs.field) → CUE reference syntax"
                    ]
                },
                {
                    name: "validate_cue_syntax"
                    description: "Verify generated CUE is syntactically correct"
                    action: "cue.validate_syntax"
                },
                {
                    name: "validate_schema_compliance"
                    description: "Ensure CUE matches #DeterministicWorkflow schema"
                    action: "schema.validate_cue_workflow"
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
                            id: "send_email"
                            name: "Send Test Email"
                            service: "gmail"
                            action: "send_message"
                            inputs: {
                                to: "${USER_INPUT:recipient}"
                                subject: "Test Subject"
                                body: "Test message"
                            }
                            outputs: { message_id: "sent_id" }
                            depends_on: []
                        }
                    ]
                    user_parameters: [
                        {
                            name: "recipient"
                            type: "string"
                            required: true
                            description: "Email recipient"
                            prompt: "Enter recipient email:"
                        }
                    ]
                    service_bindings: [
                        {
                            service: "gmail"
                            oauth_scopes: ["https://www.googleapis.com/auth/gmail.send"]
                        }
                    ]
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
