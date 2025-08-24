package agents

import "../schemas.cue"

// =============================================
// 🔹 WORKFLOW GENERATOR AGENT - PHASE 3 DETERMINISTIC
// =============================================
// Updated: JSON workflow generation from validated intents, orchestrated by agent manager

WorkflowGeneratorAgent: {
    version: "1.0.0"
    type: "autonomous_agent"
    poc_status: "essential" // Core value: Intent → JSON workflow specification
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        // Generator Object States
        {
            id: "generator_input_state"
            type: "object"
            description: "Generator receives and validates input data"
            owner: "generator"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "user_intent", type: "string", required: true }, // Original user intent
                { name: "validated_intent", type: "object", required: true }, // From intent analyst
                { name: "available_services", type: "string", required: true }, // MCP service catalog
                { name: "input_validation_status", type: "string" }, // "valid", "invalid", "pending"
                { name: "rac_context", type: "string" } // Loaded RaC specification
            ]
            metadata: {
                tags: ["generator", "input", "validation"]
            }
        },
        {
            id: "generator_output_state"
            type: "object"
            description: "Generator validates and manages final workflow output"
            owner: "generator"
            fields: [
                { name: "workflow_name", type: "string", required: true },
                { name: "json_specification", type: "object", required: true }, // Final validated JSON

                { name: "workflow_steps", type: "array", required: true }, // Extracted steps
                { name: "user_parameters", type: "object" }, // Extracted parameters
                { name: "service_bindings", type: "array" }, // Extracted service requirements
                { name: "output_validation_status", type: "string" }, // "valid", "invalid", "pending"
                { name: "execution_ready", type: "boolean" }, // Ready for workflow execution
                { name: "artifacts_saved", type: "boolean" }, // Files saved to disk
                { name: "artifacts_location", type: "string" }, // Storage backend: "local" | "gcs"
                { name: "status", type: "string" } // "generated", "validated", "ready", "error"
            ]
            metadata: {
                tags: ["generator", "output", "validation", "workflow", "json"]
            }
        },
        // LLM Object States
        {
            id: "llm_input_state"
            type: "object"
            description: "LLM receives formatted prompt and service schemas"
            owner: "llm"
            fields: [
                { name: "formatted_prompt", type: "string", required: true }, // Prepared prompt text
                { name: "user_intent", type: "string", required: true }, // User's natural language intent
                { name: "service_schemas", type: "string", required: true }, // Available services with actions
                { name: "model_config", type: "object" }, // LLM configuration (model, temperature)
                { name: "prompt_template", type: "string" } // Loaded prompt template
            ]
            metadata: {
                tags: ["llm", "input", "prompt"]
            }
        },
        {
            id: "llm_output_state"
            type: "object"
            description: "LLM generates and returns JSON workflow specification"
            owner: "llm"
            fields: [
                { name: "raw_json_response", type: "string", required: true }, // Raw LLM output
                { name: "parsed_workflow", type: "object", required: true }, // Parsed JSON workflow
                { name: "generation_metadata", type: "object" }, // Model info, tokens, etc.
                { name: "generation_status", type: "string" } // "success", "error", "invalid_json"
            ]
            metadata: {
                tags: ["llm", "output", "json_workflow"]
            }
        }
    ]
    
    events: [
        // External Events (Agent Interface)
        {
            id: "intent_analysis_complete"
            version: "1.0"
            type: "input"
            description: "Intent analysis complete, ready for workflow generation"
            target: "generator"
            data_schema: {
                user_intent: "string"
                validated_intent: "object"
                available_services: "string" // MCP service catalog with schemas
            }
            triggers: ["generator_validate_input"]
        },
        {
            id: "workflow_generation_complete"
            version: "1.0"
            type: "output"
            description: "Complete workflow generated and ready for execution"
            source: "generator"
            data_schema: {
                workflow_name: "string"
                json_specification: "object" // Final validated JSON workflow
                execution_ready: "boolean"
                artifacts_saved: "boolean"
            }
        },
        // Internal Events (Object Communication)
        {
            id: "generator_input_validated"
            version: "1.0"
            type: "internal"
            description: "Generator validated input data successfully"
            source: "generator"
            target: "generator"
            triggers: ["generator_prepare_llm_input"]
        },
        {
            id: "llm_input_prepared"
            version: "1.0"
            type: "internal"
            description: "Generator prepared LLM input data"
            source: "generator"
            target: "llm"
            triggers: ["llm_generate_workflow"]
        },
        {
            id: "llm_workflow_generated"
            version: "1.0"
            type: "internal"
            description: "LLM generated JSON workflow successfully"
            source: "llm"
            target: "generator"
            triggers: ["generator_validate_output"]
        },
        {
            id: "generator_output_validated"
            version: "1.0"
            type: "internal"
            description: "Generator validated LLM output successfully"
            source: "generator"
            target: "generator"
            triggers: ["generator_save_artifacts"]
        },

    ]
    
    logic: [
        // Generator Object Logic
        {
            id: "generator_validate_input"
            type: "validation"
            description: "Validate input data and load RaC context"
            owner: "generator"
            input_event: "intent_analysis_complete"
            output_event: "generator_input_validated"
        },
        {
            id: "generator_prepare_llm_input"
            type: "preparation"
            description: "Format prompt and service schemas for LLM"
            owner: "generator"
            input_event: "generator_input_validated"
            output_event: "llm_input_prepared"
        },
        {
            id: "generator_validate_output"
            type: "validation"
            description: "Validate LLM JSON output against schema"
            owner: "generator"
            input_event: "llm_workflow_generated"
            output_event: "generator_output_validated"
        },

        {
            id: "generator_save_artifacts"
            type: "persistence"
            description: "Save JSON workflow specification"
            owner: "generator"
            input_event: "generator_output_validated"
            output_event: "workflow_generation_complete"
        },
        // LLM Object Logic
        {
            id: "llm_generate_workflow"
            type: "ai_generation"
            description: "AI generates JSON workflow from formatted input"
            owner: "llm"
            input_event: "llm_input_prepared"
            output_event: "llm_workflow_generated"
            // Expected JSON schema for AI output
            output_schema: {
                version: "string"
                name: "string" 
                description: "string"
                steps: [{
                    id: "string"
                    name: "string"
                    action: "string"        // Service dot notation (e.g., "gmail.send_message")
                    parameters: "object"    // Step-specific parameters
                    depends_on: ["string"]  // Optional step dependencies
                }]
                user_parameters: {
                    _: {
                        type: "string"      // "string", "number", "boolean", "array", "object"
                        prompt: "string"    // User-friendly prompt text
                        required: "boolean"
                        description: "string"
                    }
                }
                service_bindings: {
                    _: {
                        type: "mcp_service"
                        auth: {
                            method: "oauth2"
                            oauth2: {
                                scopes: ["string"]
                                token_source: "user"
                            }
                        }
                    }
                }
            }
        }
    ]
    
    // LAYER 2: ARCHITECTURE (How)
    architecture: {
        type: "monolithic_service"
        description: "PoC implementation as single service with direct function calls"
        components: {
            genkit_service: {
                type: "go_service"
                location: "app/backend/internal/services/genkit.go"
                responsibilities: [
                    "workflow_generation_orchestration",
                    "ai_prompt_execution", 
                    "json_workflow_generation",
                    "artifact_file_management"
                ]
                methods: [
                    "ExecuteWorkflowGeneratorAgent()",
                    "ExecutePromptWithInputData()",
                    "SaveWorkflowArtifacts()"
                ]
            }
            workflow_generator_flow: {
                type: "genkit_flow"
                location: "app/backend/genkit/flows/"
                responsibilities: [
                    "structured_json_generation",
                    "workflow_schema_validation",
                    "service_binding_verification"
                ]
            }
        }
        communication: {
            input: ["http_api_handlers"]
            output: ["json_response", "file_artifacts"]
            external: ["google_genai_api", "openai_api", "local_file_system"]
        }
        data_flow: {
            input_processing: "Raw input → Structured input validation"
            ai_execution: "Structured input → JSON workflow generation"
            post_processing: "JSON validation → Artifact persistence"
            response: "Validated workflow → External interface"
        }
        quality_attributes: {
            correctness: "high"
            completeness: "high" 
            maintainability: "high"
            poc_simplicity: "high"
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
            id: "object_based_workflow_generation"
            type: "integration"
            description: "Test Generator and LLM object interaction for workflow generation"
            scenario: "Gmail to Document workflow with object separation validation"
            
            // Test Input (External Event)
            input_event: {
                id: "intent_analysis_complete"
                data: {
                    user_intent: "Fetch the oldest Gmail message from bojidar@investclub.bg, create a Google Doc from it in a Drive folder Email-Automation/Bojidar/{{YYYY‑MM‑DD}} if it does not exist yet."
                    validated_intent: {
                        is_automation_request: true
                        required_services: ["gmail", "docs", "drive"]
                        can_fulfill: true
                        missing_info: []
                        next_action: "generate_workflow"
                    }
                    available_services: "gmail: send_message, get_messages (OAuth: gmail.compose, gmail.readonly)\ndocs: create_document, get_document (OAuth: documents)\ndrive: upload_file, create_folder (OAuth: drive.file)"
                }
            }
            
            // Expected State Transitions
            expected_states: {
                generator_input_state: {
                    user_id: "test_user_123"
                    user_intent: "Fetch the oldest Gmail message from bojidar@investclub.bg, create a Google Doc from it in a Drive folder Email-Automation/Bojidar/{{YYYY‑MM‑DD}} if it does not exist yet."
                    validated_intent: {
                        is_automation_request: true
                        required_services: ["gmail", "docs", "drive"]
                        can_fulfill: true
                    }
                    available_services: "gmail: send_message, get_messages (OAuth: gmail.compose, gmail.readonly)\ndocs: create_document, get_document (OAuth: documents)\ndrive: upload_file, create_folder (OAuth: drive.file)"
                    input_validation_status: "valid"
                    rac_context: "loaded_rac_specification"
                }
                llm_input_state: {
                    formatted_prompt: "Generate workflow for: Fetch oldest Gmail message..."
                    user_intent: "Fetch the oldest Gmail message from bojidar@investclub.bg, create a Google Doc from it in a Drive folder Email-Automation/Bojidar/{{YYYY‑MM‑DD}} if it does not exist yet."
                    service_schemas: "gmail: send_message, get_messages (OAuth: gmail.compose, gmail.readonly)\ndocs: create_document, get_document (OAuth: documents)\ndrive: upload_file, create_folder (OAuth: drive.file)"
                    model_config: {
                        temperature: 0.1
                        max_tokens: 2000
                    }
                }
                llm_output_state: {
                    raw_json_response: "{\"version\":\"1.0\",\"name\":\"Gmail to Document Workflow\"...}"
                    parsed_workflow: {
                        version: "1.0"
                        name: "Gmail to Document Workflow"
                        description: "Process Gmail messages and create documents"
                        steps: [
                            {
                                id: "fetch_gmail"
                                name: "Fetch Gmail Message"
                                action: "gmail.get_messages"
                                parameters: {
                                    sender: "bojidar@investclub.bg"
                                    limit: 1
                                    order: "oldest"
                                }
                            },
                            {
                                id: "create_folder"
                                name: "Create Drive Folder"
                                action: "drive.create_folder"
                                depends_on: ["fetch_gmail"]
                            },
                            {
                                id: "create_document"
                                name: "Create Google Doc"
                                action: "docs.create_document"
                                depends_on: ["create_folder"]
                            }
                        ]
                        user_parameters: {
                            sender_email: {
                                type: "string"
                                prompt: "Email sender to fetch from"
                                required: true
                            }
                            folder_pattern: {
                                type: "string"
                                prompt: "Drive folder pattern"
                                required: true
                            }
                        }
                        service_bindings: {
                            gmail: {
                                type: "mcp_service"
                                auth: {
                                    method: "oauth2"
                                    oauth2: {
                                        scopes: ["https://www.googleapis.com/auth/gmail.readonly"]
                                        token_source: "user"
                                    }
                                }
                            }
                            docs: {
                                type: "mcp_service"
                                auth: {
                                    method: "oauth2"
                                    oauth2: {
                                        scopes: ["https://www.googleapis.com/auth/documents"]
                                        token_source: "user"
                                    }
                                }
                            }
                            drive: {
                                type: "mcp_service"
                                auth: {
                                    method: "oauth2"
                                    oauth2: {
                                        scopes: ["https://www.googleapis.com/auth/drive.file"]
                                        token_source: "user"
                                    }
                                }
                            }
                        }
                    }
                    generation_status: "success"
                }
                generator_output_state: {
                    workflow_name: "string"
                    json_specification: {
                        version: "string"
                        name: "string"
                        steps: [{
                            id: "string"
                            name: "string"
                            action: "string"
                            parameters: "object"
                            depends_on: ["string"]
                        }]
                    }
                    workflow_steps: [
                        { id: "fetch_gmail", action: "gmail.get_messages" },
                        { id: "create_folder", action: "drive.create_folder" },
                        { id: "create_document", action: "docs.create_document" }
                    ]
                    output_validation_status: "valid"
                    execution_ready: true
                    artifacts_saved: true
                    status: "ready"
                }
            }
            
            // Expected Event Flow
            expected_events: [
                { id: "intent_analysis_complete", type: "input", target: "generator" },
                { id: "generator_input_validated", type: "internal", source: "generator", target: "generator" },
                { id: "llm_input_prepared", type: "internal", source: "generator", target: "llm" },
                { id: "llm_workflow_generated", type: "internal", source: "llm", target: "generator" },
                { id: "generator_output_validated", type: "internal", source: "generator", target: "generator" },

                { id: "workflow_generation_complete", type: "output", source: "generator" }
            ]
            
            // Expected Logic Execution
            expected_logic_flow: [
                { id: "generator_validate_input", owner: "generator", type: "validation" },
                { id: "generator_prepare_llm_input", owner: "generator", type: "preparation" },
                { id: "llm_generate_workflow", owner: "llm", type: "ai_generation" },
                { id: "generator_validate_output", owner: "generator", type: "validation" },

                { id: "generator_save_artifacts", owner: "generator", type: "persistence" }
            ]
            
            // Final Output Event
            expected_output: {
                event_id: "workflow_generation_complete"
                data: {
                    workflow_name: "Gmail to Document Workflow"
                    json_specification: {
                        version: "1.0"
                        name: "Gmail to Document Workflow"
                        steps: [
                            { id: "fetch_gmail", action: "gmail.get_messages" },
                            { id: "create_folder", action: "drive.create_folder" },
                            { id: "create_document", action: "docs.create_document" }
                        ]
                    }
                    cue_content: "package workflows\n\nworkflow: #DeterministicWorkflow & {...}"
                    workflow_file: {
                        id: "test_user_123_gmail_to_document_workflow"
                        filename: "gmail_to_document_workflow.cue"
                        path: "/workflows/test_user_123/gmail_to_document_workflow.cue"
                        saved_at: "2025-08-13T18:47:00Z"
                    }
                    artifacts_location: "local"
                    cue_conversion_status: "success"
                    execution_ready: true
                    artifacts_saved: true
                }
            }
            
            // Validation Rules
            validation: {
                state_ownership: "Each state must have correct owner (generator or llm)"
                event_flow: "Events must follow: input → internal transitions → output"
                logic_separation: "Generator handles orchestration, LLM handles generation"
                data_consistency: "State data must be consistent across transitions"
                technology_agnostic: "No implementation details in events/logic/states"
            }
        },
        {
            id: "simple_email_workflow_generation"
            type: "unit"
            description: "Test simple single-service workflow generation"
            scenario: "Email workflow with minimal object interaction"
            
            input_event: {
                id: "intent_analysis_complete"
                data: {
                    user_intent: "Send a weekly report email to my team"
                    validated_intent: {
                        is_automation_request: true
                        required_services: ["gmail"]
                        can_fulfill: true
                        missing_info: []
                        next_action: "generate_workflow"
                    }
                    available_services: "gmail: send_message, create_draft (OAuth: gmail.compose)"
                }
            }
            expected: {
                json_specification: {
                    version: "1.0"
                    name: "Gmail to Document Workflow"
                    description: "Process Gmail messages and create documents"
                    steps: [
                        {
                            id: "fetch_gmail"
                            name: "Fetch Gmail Message"
                            action: "gmail.get_messages"
                            parameters: {
                                sender: "bojidar@investclub.bg"
                                limit: 1
                                order: "oldest"
                            }
                        },
                        {
                            id: "create_folder"
                            name: "Create Drive Folder"
                            action: "drive.create_folder"
                            depends_on: ["fetch_gmail"]
                        },
                        {
                            id: "create_document"
                            name: "Create Google Doc"
                            action: "docs.create_document"
                            depends_on: ["create_folder"]
                        }
                    ]
                    user_parameters: {
                        sender_email: {
                            type: "string"
                            prompt: "Email sender to fetch from"
                            required: true
                            placeholder: "sender@example.com"
                        }
                        folder_name: {
                            type: "string"
                            prompt: "Drive folder name"
                            required: true
                            placeholder: "Email-Automation/{{YYYY-MM-DD}}"
                        }
                    }
                    service_bindings: {
                        gmail: {
                            type: "mcp_service"
                            auth: {
                                method: "oauth2"
                                oauth2: {
                                    scopes: ["https://www.googleapis.com/auth/gmail.readonly"]
                                    token_source: "user"
                                }
                            }
                        }
                        docs: {
                            type: "mcp_service"
                            auth: {
                                method: "oauth2"
                                oauth2: {
                                    scopes: ["https://www.googleapis.com/auth/documents"]
                                    token_source: "user"
                                }
                            }
                        }
                        drive: {
                            type: "mcp_service"
                            auth: {
                                method: "oauth2"
                                oauth2: {
                                    scopes: ["https://www.googleapis.com/auth/drive.file"]
                                    token_source: "user"
                                }
                            }
                        }
                    }
                }
                validation_results: {
                    schema_valid: true
                    services_available: true
                    oauth_scopes_valid: true
                }
                execution_ready: true
                status: "ready"
            }
            validation: {
                required_fields: ["json_specification", "validation_results"]
                json_schema_compliance: true
                service_binding_verification: true
                step_dependency_validation: true
            }
        },
        {
            id: "simple_email_workflow_generation"
            type: "unit"
            description: "Generate simple email sending workflow"
            input: {
                user_intent: "Send a weekly report email to my team"
                validated_intent: {
                    is_automation_request: true
                    required_services: ["gmail"]
                    can_fulfill: true
                    missing_info: []
                    next_action: "generate_workflow"
                }
                available_services: {
                    gmail: {
                        actions: ["gmail.send_message", "gmail.get_message"]
                        oauth_scopes: ["https://www.googleapis.com/auth/gmail.compose"]
                    }
                }
                rac_context: "workflow_generator.cue specification"
            }
            expected: {
                json_specification: {
                    version: "1.0"
                    name: "Weekly Report Email"
                    steps: [
                        {
                            id: "send_report"
                            name: "Send Weekly Report"
                            action: "gmail.send_message"
                        }
                    ]
                    user_parameters: {
                        recipient_email: {
                            type: "string"
                            required: true
                        }
                        subject: {
                            type: "string"
                            required: true
                        }
                        body: {
                            type: "string"
                            required: true
                        }
                    }
                    service_bindings: {
                        gmail: {
                            type: "mcp_service"
                            auth: {
                                method: "oauth2"
                                oauth2: {
                                    scopes: ["https://www.googleapis.com/auth/gmail.compose"]
                                    token_source: "user"
                                }
                            }
                        }
                    }
                }
                execution_ready: true
                status: "ready"
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
