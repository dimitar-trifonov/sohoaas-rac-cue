package agents

import "../schemas.cue"

// =============================================
// 🔹 PERSONAL CAPABILITIES AGENT - PHASE 3
// =============================================
// Updated: Static service-to-capability mappings, centralized service catalog

PersonalCapabilitiesAgent: {
    version: "1.0.0"
    type: "autonomous_agent"
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        {
            id: "personal_capabilities"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "service_catalog", type: "object", required: true }, // Google Workspace service catalog
                { name: "user_capabilities", type: "object", required: true }, // Structured capability mapping
                { name: "available_actions", type: "array", required: true },
                { name: "examples", type: "array" },
                { name: "status", type: "string" } // "discovering", "ready", "error"
            ]
            metadata: {
                tags: ["capabilities", "discovery", "personal", "google_workspace", "static_mapping"]
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
            description: "Personal capabilities with Google Workspace service catalog ready"
            data_schema: {
                user_id: "string"
                service_catalog: {
                    gmail: {
                        functions: ["send_email", "search_emails", "create_draft"]
                                "[function_name]": {
                                    parameters: {
                                        "[param_name]": {
                                            type: "string"
                                            required: "boolean"
                                            description: "string"
                                            validation: "string"
                                            default: "any"
                                        }
                                    }
                                    outputs: {
                                        "[output_name]": {
                                            type: "string"
                                            description: "string"
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
                available_actions: ["string"]
                examples: ["string"]
            }
        }
    ]
    
    logic: [
        {
            id: "discover_capabilities"
            type: "genkit_flow"
            description: "Discover complete MCP service schemas with parameter validation and function signatures"
            input_schema: {
                user_id: "string"
                connected_services: ["string"]
                mcp_endpoint: "string"
            }
            steps: [
                {
                    name: "query_mcp_providers"
                    description: "Query MCP server for available providers and services"
                    action: "http.get"
                    endpoint: "${input.mcp_endpoint}/api/services"
                    headers: {
                        "Content-Type": "application/json"
                        "Authorization": "Bearer ${user.oauth_token}"
                    }
                },
                {
                    name: "discover_function_schemas"
                    description: "Query each MCP function for complete parameter and output schemas"
                    action: "mcp.discover_schemas"
                    parallel_requests: true
                    schema_discovery: {
                        for_each_provider: "${step.query_mcp_providers.providers}"
                        for_each_service: "${provider.services}"
                        for_each_function: "${service.functions}"
                        query_endpoint: "${input.mcp_endpoint}/api/services/${provider}/${service}/${function}/schema"
                        extract_schema: {
                            parameters: {
                                required_fields: "array"
                                optional_fields: "array"
                                field_types: "object"
                                validation_rules: "object"
                                default_values: "object"
                            }
                            outputs: {
                                return_fields: "object"
                                field_descriptions: "object"
                            }
                        }
                    }
                },
                {
                    name: "build_enhanced_catalog"
                    description: "Build complete service catalog with parameter schemas"
                    action: "catalog.build_enhanced"
                    catalog_structure: {
                        providers: {
                            workspace: {
                                drive: {
                                    create_folder: {
                                        parameters: {
                                            name: {type: "string", required: true, description: "Folder name"}
                                            parent_id: {type: "string", required: false, default: "root", description: "Parent folder ID"}
                                        }
                                        outputs: {
                                            folder_id: {type: "string", description: "Created folder ID"}
                                            folder_url: {type: "string", description: "Folder sharing URL"}
                                        }
                                    }
                                    share_file: {
                                        parameters: {
                                            file_id: {type: "string", required: true, description: "File ID to share"}
                                            email: {type: "string", required: true, validation: "email", description: "Email to share with"}
                                            role: {type: "string", required: true, enum: ["reader", "writer", "owner"], description: "Permission level"}
                                        }
                                        outputs: {
                                            permission_id: {type: "string", description: "Permission ID"}
                                        }
                                    }
                                }
                                docs: {
                                    create_document: {
                                        parameters: {
                                            title: {type: "string", required: true, description: "Document title"}
                                            folder_id: {type: "string", required: false, description: "Folder to create document in"}
                                        }
                                        outputs: {
                                            document_id: {type: "string", description: "Created document ID"}
                                            document_url: {type: "string", description: "Document editing URL"}
                                        }
                                    }
                                    insert_text: {
                                        parameters: {
                                            document_id: {type: "string", required: true, description: "Document ID"}
                                            text: {type: "string", required: true, description: "Text to insert"}
                                            index: {type: "number", required: false, default: 1, description: "Position to insert text"}
                                        }
                                        outputs: {
                                            success: {type: "boolean", description: "Operation success"}
                                        }
                                    }
                                }
                                gmail: {
                                    send_message: {
                                        parameters: {
                                            to: {type: "string", required: true, validation: "email", description: "Recipient email"}
                                            subject: {type: "string", required: true, description: "Email subject"}
                                            body: {type: "string", required: true, description: "Email body"}
                                            cc: {type: "string", required: false, validation: "email", description: "CC email"}
                                            bcc: {type: "string", required: false, validation: "email", description: "BCC email"}
                                        }
                                        outputs: {
                                            message_id: {type: "string", description: "Sent message ID"}
                                            thread_id: {type: "string", description: "Email thread ID"}
                                        }
                                    }
                                }
                                calendar: {
                                    create_event: {
                                        parameters: {
                                            summary: {type: "string", required: true, description: "Event title"}
                                            start: {type: "object", required: true, description: "Event start time"}
                                            end: {type: "object", required: false, description: "Event end time"}
                                            attendees: {type: "array", required: false, description: "Event attendees"}
                                            description: {type: "string", required: false, description: "Event description"}
                                        }
                                        outputs: {
                                            event_id: {type: "string", description: "Created event ID"}
                                            event_url: {type: "string", description: "Event URL"}
                                        }
                                    }
                                }
                            }
                        }
                    }
                },
                {
                    name: "generate_capability_examples"
                    description: "Generate contextual examples based on discovered schemas"
                    action: "template.generate_schema_examples"
                    example_generation: {
                        based_on_schemas: "${step.build_enhanced_catalog.catalog}"
                        example_templates: {
                            folder_creation: "Create project folders with proper naming and sharing"
                            document_automation: "Generate meeting agendas with calendar integration"
                            email_workflows: "Send follow-up emails with document attachments"
                            calendar_management: "Schedule meetings with automatic invitations"
                        }
                    }
                }
            ]
            output_event: "capabilities_discovered"
        }
    ]
    
    // LAYER 2: ARCHITECTURE (How)
    architecture: {
        type: "microservice"
        components: {
            mcp_schema_discoverer: {
                type: "http_client"
                responsibilities: ["mcp_api_queries", "schema_discovery", "parameter_extraction"]
            }
            catalog_builder: {
                type: "data_processor"
                responsibilities: ["schema_aggregation", "catalog_construction", "validation_rules"]
            }
            capability_mapper: {
                type: "mapping_service"
                responsibilities: ["schema_to_capability_mapping", "example_generation"]
            }
        }
        communication: {
            input: ["event_bus"]
            output: ["event_bus"]
            external: ["mcp_server_api"] // MCP server for schema discovery
        }
        quality_attributes: {
            reliability: "high"
            performance: "medium"
            scalability: "medium"
        }
    }
    
    // LAYER 3: IMPLEMENTATION (With What)
    bindings: [
        {
            type: "deployment"
            technology: "golang"
            framework: "genkit"
            deployment: {
                service_name: "personal-capabilities-agent"
                port: 8081
                resources: {
                    cpu: "0.5"
                    memory: "512Mi"
                }
                config: {
                    mcp_endpoint: "http://localhost:8080"
                    schema_cache_ttl: "1h"
                    parallel_discovery: true
                    max_concurrent_requests: 10
                }
            }
        }
    ]
    
    tests: [
        {
            id: "capability_discovery_flow"
            type: "integration"
            description: "Test full capability discovery for authenticated user"
            input: {
                event: "user_authenticated"
                data: {
                    user_id: "test@example.com"
                    oauth_tokens: { access_token: "mock_token" }
                    connected_services: ["gmail", "docs", "calendar"]
                }
            }
            expected: {
                event: "capabilities_discovered"
                data: {
                    user_id: "test@example.com"
                    available_services: {
                        providers: {
                            workspace: {
                                gmail: {
                                    send_message: {
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
                                docs: {
                                    create_document: {
                                        parameters: {
                                            title: {type: "string", required: true}
                                        }
                                        outputs: {
                                            document_id: {type: "string"}
                                            document_url: {type: "string"}
                                        }
                                    }
                                }
                                calendar: {
                                    create_event: {
                                        parameters: {
                                            summary: {type: "string", required: true}
                                            start: {type: "object", required: true}
                                        }
                                        outputs: {
                                            event_id: {type: "string"}
                                        }
                                    }
                                }
                            }
                        }
                    }
                    available_actions: ["send_email", "create_document", "schedule_meeting"]
                    status: "ready"
                }
            }
        }
    ]
    
    ui: [
        {
            id: "capabilities_display"
            type: "component"
            description: "Show discovered capabilities to user"
            props: {
                capabilities: "personal_capabilities"
                examples: "array"
            }
            events: ["capabilities_ready"]
        }
    ]
}
