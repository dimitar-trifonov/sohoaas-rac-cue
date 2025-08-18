package services

import "../schemas.cue"

// =============================================
// 🔹 PERSONAL CAPABILITIES AGENT - PHASE 3
// =============================================
// Updated: Static service-to-capability mappings, centralized service catalog

PersonalCapabilitiesService: {
    version: "1.0.0"
    type: "deterministic_service"
    poc_status: "essential" // Core value: MCP service discovery and capability mapping
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        {
            id: "mcp_service_discovery"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "service_catalog", type: "object", required: true }, // MCP service schemas
                { name: "available_actions", type: "array", required: true }, // Available MCP functions
                { name: "status", type: "string" } // "discovering", "ready", "error"
            ]
            metadata: {
                tags: ["mcp", "discovery", "deterministic", "service_catalog"]
            }
        }
    ]
    
    events: [
        {
            id: "user_authenticated"
            version: "1.0"
            type: "input"
            description: "User OAuth2 authentication completed"
            triggers: ["discover_capabilities"]
        },
        {
            id: "capabilities_discovered"
            version: "1.0"
            type: "output"
            description: "MCP service catalog discovered and ready for workflow generation"
            data_schema: {
                user_id: "string"
                service_catalog: "object" // MCP service schemas with parameter definitions
                available_actions: "array" // Available MCP function names
                parameter_schemas: "object" // Function parameter and output schemas for LLM
            }
        }
    ]
    
    logic: [
        {
            id: "discover_capabilities"
            type: "deterministic_discovery"
            description: "Query MCP server for available service schemas and function definitions"
            input_schema: {
                user_id: "string"
                connected_services: ["string"]
                mcp_endpoint: "string"
            }
            steps: [
                {
                    name: "query_mcp_services"
                    description: "Query MCP server for available services and functions"
                    action: "mcp.list_services"
                    endpoint: "${input.mcp_endpoint}/api/services"
                },
                {
                    name: "build_service_catalog"
                    description: "Build service catalog with parameter schemas from MCP response"
                    action: "catalog.build_with_schemas"
                    input: "${step.query_mcp_services.services}"
                    schema_extraction: {
                        include_parameters: true
                        include_outputs: true
                        include_validation_rules: true
                    }
                    output: {
                        state: "mcp_service_discovery"
                        fields: ["service_catalog", "available_actions", "parameter_schemas", "status"]
                    }
                }
            ]
            output_event: "capabilities_discovered"
        }
    ]
    
    // LAYER 2: TESTS (Validation)
    tests: [
        {
            id: "mcp_service_discovery_test"
            description: "Test MCP service discovery for authenticated user"
            input: {
                user_id: "test@example.com"
                connected_services: ["gmail", "docs", "calendar"]
                mcp_endpoint: "http://localhost:8080"
            }
            expected: {
                service_catalog: "object" // MCP service schemas
                available_actions: ["send_message", "create_document", "create_event"]
                parameter_schemas: {
                    "gmail.send_message": {
                        parameters: {
                            to: {type: "string", required: true, validation: "email"}
                            subject: {type: "string", required: true}
                            body: {type: "string", required: true}
                        }
                        outputs: {
                            message_id: {type: "string"}
                        }
                    }
                }
                status: "ready"
            }
        }
    ]
}
