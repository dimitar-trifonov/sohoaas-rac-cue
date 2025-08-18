package services

import "../schemas.cue"

// =============================================
// 🔹 MCP VALIDATOR SERVICE
// =============================================
// Validates workflow steps against actual MCP service functions and parameters

MCPValidatorService: {
    version: "1.0.0"
    type: "validation_service"
    poc_status: "essential" // Critical for workflow execution pipeline
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        {
            id: "mcp_validation_input"
            type: "object"
            description: "Input workflow for MCP validation"
            owner: "validator"
            fields: [
                { name: "workflow_steps", type: "array", required: true },
                { name: "mcp_service_registry", type: "object", required: true },
                { name: "validation_mode", type: "string" } // "strict" | "permissive"
            ]
            metadata: {
                tags: ["validation", "input", "mcp"]
            }
        },
        {
            id: "mcp_validation_results"
            type: "object"
            description: "MCP validation results with detailed error reporting"
            owner: "validator"
            fields: [
                { name: "is_valid", type: "boolean", required: true },
                { name: "function_validation_results", type: "array", required: true },
                { name: "parameter_validation_results", type: "array", required: true },
                { name: "validation_errors", type: "array" },
                { name: "validation_warnings", type: "array" }
            ]
            metadata: {
                tags: ["validation", "results", "mcp"]
            }
        }
    ]
    
    events: [
        {
            id: "mcp_validation_requested"
            version: "1.0"
            type: "input"
            description: "Request MCP function and parameter validation"
            target: "validator"
            data_schema: {
                workflow_steps: "array"
                mcp_service_registry: "object"
                validation_mode: "string"
            }
            triggers: ["validate_mcp_functions", "validate_mcp_parameters"]
        },
        {
            id: "mcp_validation_complete"
            version: "1.0"
            type: "output"
            description: "MCP validation completed with results"
            source: "validator"
            data_schema: {
                is_valid: "boolean"
                function_validation_results: "array"
                parameter_validation_results: "array"
                validation_errors: "array"
            }
        }
    ]
    
    // LAYER 2: LOGIC (How it works)
    logic: [
        {
            id: "validate_mcp_functions"
            type: "function_validation"
            description: "Validate each step uses valid MCP service functions"
            steps: [
                {
                    name: "load_mcp_service_registry"
                    description: "Load available MCP functions from service registry"
                    action: "mcp.load_service_registry"
                    endpoint: "http://localhost:8080/api/services"
                },
                {
                    name: "validate_step_actions"
                    description: "Check each step.action against MCP function registry"
                    action: "mcp.validate_function_exists"
                    validation_rules: [
                        "step.action must match MCP function name exactly",
                        "function must exist in MCP service registry",
                        "service must be available and enabled"
                    ]
                },
                {
                    name: "check_service_availability"
                    description: "Verify MCP services are accessible"
                    action: "mcp.check_service_status"
                }
            ]
            output: {
                state: "mcp_validation_results"
                fields: ["function_validation_results"]
            }
        },
        {
            id: "validate_mcp_parameters"
            type: "parameter_validation"
            description: "Validate step parameters against MCP function schemas"
            steps: [
                {
                    name: "load_function_schemas"
                    description: "Load input schemas for each MCP function"
                    action: "mcp.load_function_schemas"
                },
                {
                    name: "validate_parameter_structure"
                    description: "Check parameters match MCP function input schema"
                    action: "mcp.validate_parameter_schema"
                    validation_rules: [
                        "All required parameters must be provided",
                        "Parameter types must match schema expectations",
                        "No unexpected parameters allowed",
                        "Parameter references (${user.param}) must be resolvable"
                    ]
                },
                {
                    name: "validate_parameter_references"
                    description: "Check parameter reference syntax and targets"
                    action: "mcp.validate_parameter_references"
                    reference_patterns: [
                        "${user.parameter_name}",
                        "${steps.step_id.outputs.output_name}",
                        "${computed.expression}"
                    ]
                }
            ]
            output: {
                state: "mcp_validation_results"
                fields: ["parameter_validation_results"]
            }
        }
    ]
    
    // LAYER 3: ARCHITECTURE (How)
    architecture: {
        type: "validation_service"
        description: "Standalone validation service for MCP function and parameter checking"
        components: {
            mcp_registry_client: {
                type: "http_client"
                location: "app/backend/internal/services/mcp_client.go"
                responsibilities: [
                    "mcp_service_registry_access",
                    "function_schema_retrieval",
                    "service_availability_checking"
                ]
                methods: [
                    "LoadServiceRegistry()",
                    "GetFunctionSchema(functionName)",
                    "CheckServiceStatus(serviceName)"
                ]
            }
            function_validator: {
                type: "validation_engine"
                location: "app/backend/internal/services/mcp_validator.go"
                responsibilities: [
                    "function_existence_validation",
                    "parameter_schema_validation",
                    "parameter_reference_validation"
                ]
                methods: [
                    "ValidateFunctionExists(action)",
                    "ValidateParameterSchema(params, schema)",
                    "ValidateParameterReferences(params)"
                ]
            }
        }
        communication: {
            input: ["json_workflow_data"]
            output: ["validation_results"]
            external: ["mcp_service_registry", "function_schemas"]
        }
        quality_attributes: {
            correctness: "high"
            performance: "medium"
            reliability: "high"
        }
    }
    
    // LAYER 4: IMPLEMENTATION (With What)
    bindings: [
        {
            type: "deployment"
            technology: "golang"
            framework: "genkit"
            deployment: {
                service_name: "sohoaas-backend" // Monolithic service for PoC
                deployment_type: "in_process_service"
                platform: "cloud_run"
                config: {
                    mcp_registry_url: "${MCP_REGISTRY_URL}"
                    validation_timeout: "30s"
                }
            }
        }
    ]
    
    tests: [
        {
            id: "validate_gmail_send_message"
            type: "unit"
            description: "Validate gmail.send_message function and parameters"
            input: {
                workflow_steps: [
                    {
                        id: "send_email"
                        name: "Send Email"
                        action: "gmail.send_message"
                        parameters: {
                            to: "${user.recipient_email}"
                            subject: "${user.email_subject}"
                            body: "${user.message_body}"
                        }
                    }
                ]
                mcp_service_registry: {
                    gmail: {
                        functions: {
                            "gmail.send_message": {
                                required_params: ["to", "subject", "body"]
                                param_types: {
                                    to: "string"
                                    subject: "string"
                                    body: "string"
                                }
                            }
                        }
                    }
                }
            }
            expected: {
                is_valid: true
                function_validation_results: [
                    {
                        step_id: "send_email"
                        action: "gmail.send_message"
                        function_exists: true
                        service_available: true
                    }
                ]
                parameter_validation_results: [
                    {
                        step_id: "send_email"
                        required_params_present: true
                        param_types_valid: true
                        parameter_references_valid: true
                    }
                ]
                validation_errors: []
            }
        },
        {
            id: "validate_invalid_function"
            type: "unit"
            description: "Test validation of non-existent MCP function"
            input: {
                workflow_steps: [
                    {
                        id: "invalid_step"
                        name: "Invalid Step"
                        action: "nonexistent.function"
                        parameters: {}
                    }
                ]
            }
            expected: {
                is_valid: false
                validation_errors: [
                    {
                        type: "function_not_found"
                        step_id: "invalid_step"
                        action: "nonexistent.function"
                        message: "MCP function 'nonexistent.function' not found in service registry"
                    }
                ]
            }
        },
        {
            id: "validate_missing_required_parameter"
            type: "unit"
            description: "Test validation of missing required parameters"
            input: {
                workflow_steps: [
                    {
                        id: "incomplete_step"
                        name: "Incomplete Step"
                        action: "gmail.send_message"
                        parameters: {
                            to: "${user.recipient_email}"
                            // Missing required 'subject' and 'body' parameters
                        }
                    }
                ]
            }
            expected: {
                is_valid: false
                validation_errors: [
                    {
                        type: "missing_required_parameter"
                        step_id: "incomplete_step"
                        missing_params: ["subject", "body"]
                        message: "Required parameters missing: subject, body"
                    }
                ]
            }
        }
    ]
}
