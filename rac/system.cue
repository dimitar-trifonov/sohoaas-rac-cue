package rac

// =============================================
// 🔹 RAC SYSTEM DEFINITION
// =============================================

// Base RaC System structure
RacSystem: {
    version?: string | *"1.0"
    states:   [...#State]
    events:   [...#Event]
    logic:    [...#Logic]
    ui:       [...#UI]
    tests:    [...#Test]
    bindings: [...#Binding]
}

// =============================================
// 🔹 SOHOAAS SYSTEM SPECIFICATION
// =============================================

// SOHOAAS PoC System - Phase 3 Complete Implementation
// Focus: Multi-agent workflow automation with personal OAuth2 authentication
// Architecture: In-process deployment with 4 core agents + Agent Manager
SOHOAASSystem: RacSystem & {
    version: "3.0.0"
    
    // DATA FORMAT AUTHORITY CHAIN - CRITICAL ARCHITECTURAL DECISION
    // Since we must execute against actual MCP services, MCP metadata is authoritative
    data_format_authority: {
        source_of_truth: "mcp_service_metadata"
        description: "MCP service metadata defines actual callable functions and parameters"
        
        authority_hierarchy: {
            level_1_authoritative: {
                component: "MCP Server Response"
                format: "Actual MCP service catalog response (PoC minimalistic)"
                role: "Defines available services and functions from live MCP server endpoint"
                example_structure: {
                    providers: {
                        workspace: {
                            description: "Google Workspace Provider"
                            display_name: "Google Workspace"
                            services: {
                                gmail: {
                                    description: "Send, receive, and manage emails using Gmail API"
                                    display_name: "Gmail"
                                    functions: {
                                        send_message: {
                                            name: "send_message"
                                            display_name: "Send Email"
                                            description: "Send an email message via Gmail"
                                            example_payload: {
                                                to: "recipient@example.com"
                                                subject: "Test Email"
                                                body: "This is a test email from SOHOaaS"
                                            }
                                            required_fields: ["to", "subject", "body"]
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
            
            level_2_derived: {
                component: "CUE Workflow Schema"
                format: "deterministic_workflow.cue"
                role: "Must align with MCP tool names and parameter structures"
                alignment_rule: "CUE step.action must match MCP tool name exactly"
            }
            
            level_3_derived: {
                component: "JSON Schema for LLM"
                format: "workflow_json_schema.json"
                role: "Must generate JSON that converts to MCP-compatible CUE"
                alignment_rule: "JSON schema must enforce MCP tool name compliance"
            }
        }
        
        conversion_rules: {
            mcp_to_cue: "MCP tool names become CUE step actions (direct mapping)"
            mcp_to_json: "MCP inputSchema becomes JSON schema parameter validation"
            json_to_cue: "Deterministic conversion preserving MCP compatibility"
        }
        
        validation_requirements: {
            mcp_alignment: "All workflow steps must reference valid MCP tools"
            parameter_validation: "Step parameters must match MCP inputSchema exactly"
            service_discovery: "Workflow generation must query MCP for available tools"
            provider_metadata_validation: "New MCP providers must validate against mcp_service_metadata_schema.json"
        }
        
        provider_onboarding: {
            metadata_schema: "rac/schemas/mcp_service_metadata_schema.json"
            validation_process: [
                "1. New provider submits metadata JSON",
                "2. Validate against mcp_service_metadata_schema.json",
                "3. Verify tool naming conventions (service_action format)",
                "4. Validate inputSchema compliance with JSON Schema v7",
                "5. Test OAuth scopes and authentication flow",
                "6. Register provider in SOHOAAS system"
            ]
            required_fields: [
                "provider_info.name",
                "provider_info.service_types", 
                "tools[].name",
                "tools[].inputSchema",
                "capabilities"
            ]
        }
    }
    
    // SYSTEM ARCHITECTURE - MINIMALISTIC EVENT-DRIVEN DESIGN
    // Critical architectural clarification for implementation alignment
    system_architecture: {
        design_philosophy: "minimalistic_event_driven"
        description: "Agent Manager serves as central orchestrator without complex event infrastructure"
        
        event_system: {
            type: "orchestrated_method_calls"
            rationale: "PoC-appropriate implementation avoiding over-engineering"
            implementation: "Direct method calls through Agent Manager ARE the event mechanism"
            benefits: ["simple", "maintainable", "debuggable", "predictable_execution"]
        }
        
        agent_coordination: {
            orchestrator: "AgentManager"
            communication_pattern: "hub_and_spoke"
            isolation_principle: "agents communicate ONLY through Agent Manager"
            state_management: "centralized in Agent Manager"
            
            event_flow: {
                oauth_login: "GetPersonalCapabilities() → ExecutePersonalCapabilitiesAgent()"
                message_received: "ProcessUserMessage() → ExecuteIntentGathererAgent()"
                intent_analysis: "AnalyzeIntent() → ExecuteIntentAnalystAgent()"
                workflow_generation: "GenerateWorkflow() → ExecuteWorkflowGeneratorAgent()"
                workflow_validation: "ValidateWorkflow() → RaC-compliant validation methods"
            }
        }
        
        execution_model: {
            mode: "sequential_only"  // PoC strategic decision
            coordination: "consecutive_app_steps"
            complexity: "minimal_for_poc_validation"
            infrastructure: "no_event_bus_or_queue_needed"
        }
        
        data_authority_implementation: {
            mcp_catalog: "loadServiceCatalogFromMCP() maintains single source of truth"
            service_isolation: "thread-safe catalog management with sync.RWMutex"
            agent_initialization: "initializeAgents() sets up 4 core agents with proper isolation"
        }
    }
    
    // LAYER 1: REQUIREMENTS (What the system should do)
    // Core PoC States: OAuth2 + Personal Automations
    states: [
        {
            id: "user_auth"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "email", type: "string", required: true },
                { name: "oauth_tokens", type: "object", required: true }, // Google OAuth2 tokens
                { name: "connected_services", type: "array" }, // ["gmail", "docs", "calendar"]
                { name: "session_expires", type: "string" }
            ]
            metadata: {
                tags: ["auth", "oauth2", "personal"]
            }
        },
        {
            id: "personal_capabilities"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "mcp_servers", type: "array", required: true }, // User's connected services
                { name: "available_actions", type: "array", required: true },
                { name: "examples", type: "array" },
                { name: "status", type: "string" } // "discovering", "ready", "error"
            ]
            metadata: {
                tags: ["capabilities", "discovery", "personal", "mcp"]
            }
        },
        {
            id: "workflow_discovery"
            type: "object"
            fields: [
                { name: "session_id", type: "string", required: true },
                { name: "user_id", type: "string", required: true },
                { name: "conversation_history", type: "array", required: true }, // Multi-turn conversation
                { name: "discovered_patterns", type: "array" }, // Repetitive processes identified
                { name: "trigger_identification", type: "object" }, // When automation should activate
                { name: "action_sequence", type: "array" }, // Workflow steps mapped
                { name: "data_requirements", type: "object" }, // Services and information needed
                { name: "discovery_phase", type: "string" }, // "pattern", "trigger", "action", "data", "validation"
                { name: "status", type: "string" } // "discovering", "complete", "needs_clarification"
            ]
            metadata: {
                tags: ["discovery", "multi_turn", "workflow_patterns", "conversation"]
            }
        },
        {
            id: "user_intent"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "original_request", type: "string", required: true },
                { name: "extracted_parameters", type: "object", required: true },
                { name: "intent_type", type: "string", required: true }, // "automation", "query", "help"
                { name: "confidence", type: "string" }, // "high", "medium", "low"
                { name: "timestamp", type: "string", required: true }
            ]
            metadata: {
                tags: ["intent", "user_input", "nlp", "parameters"]
            }
        },
        {
            id: "mcp_service_mapping"
            type: "object"
            fields: [
                { name: "available_services", type: "array", required: true }, // ["gmail", "docs", "calendar"]
                { name: "service_capabilities", type: "object", required: true }, // Actions per service
                { name: "oauth_scopes", type: "object", required: true }, // Required scopes per service
                { name: "intent_to_service_mapping", type: "object", required: true }, // Intent → Service mappings
                { name: "validation_rules", type: "object" }, // Service-specific validation
                { name: "last_updated", type: "string" }
            ]
            metadata: {
                tags: ["mcp", "services", "oauth", "capabilities", "mapping"]
            }
        },
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
        },
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
                { name: "cue_specification", type: "string", required: true }, // Complete CUE file
                { name: "validation_results", type: "object" }, // Schema and service validation
                { name: "execution_ready", type: "boolean" }, // Ready for immediate execution
                { name: "status", type: "string" } // "generated", "validated", "ready", "error"
            ]
            metadata: {
                tags: ["workflow", "deterministic", "cue_file", "executable", "steps_based"]
            }
        },
        {
            id: "workflow_validation"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "workflow_name", type: "string", required: true },
                { name: "parameter_schema", type: "object" }, // Required parameter formats and types
                { name: "mcp_requirements", type: "object" }, // Required MCP services and bindings
                { name: "completeness_check", type: "array" }, // Missing parameters or MCP bindings
                { name: "format_validation", type: "array" }, // Parameter format validation results
                { name: "validation_status", type: "string" }, // "checking", "incomplete", "invalid_format", "approved"
                { name: "workflow_ready", type: "boolean" } // True if workflow can be executed
            ]
            metadata: {
                tags: ["validation", "format", "completeness", "mcp"]
            }
        },
        {
            id: "intent_confirmation"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "original_intent", type: "string" }, // User's original message
                { name: "understood_workflow", type: "object" }, // What LLM thinks user wants
                { name: "missing_data", type: "array" }, // Data that needs clarification
                { name: "suggested_steps", type: "array" }, // Proposed workflow steps
                { name: "user_approval", type: "string" }, // "approved", "modified", "rejected"
                { name: "modifications", type: "array" }, // User requested changes
                { name: "status", type: "string" } // "pending", "confirmed", "revised"
            ]
            metadata: {
                tags: ["confirmation", "validation", "user_approval"]
            }
        },
        {
            id: "rac_execution"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "workflow_name", type: "string", required: true },
                { name: "rac_specification", type: "object", required: true }, // Approved RaC CUE file (reusable)
                { name: "execution_id", type: "string", required: true },
                { name: "execution_parameters", type: "object", required: true }, // User-editable parameters for this run
                { name: "parameter_history", type: "array" }, // Previous parameter values for reuse
                { name: "current_event", type: "string" }, // Which event is executing
                { name: "state_values", type: "object" }, // Current state data
                { name: "execution_log", type: "array" }, // Deterministic execution trace
                { name: "status", type: "string" }, // "ready", "running", "completed", "failed"
                { name: "result", type: "object" } // Final execution results
            ]
            metadata: {
                tags: ["execution", "deterministic", "rac", "reusable", "editable"]
            }
        },
        {
            id: "personal_workflow"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true }, // Owner of the workflow
                { name: "workflow_type", type: "string" }, // "proposal_followup", "meeting_prep", "team_update"
                { name: "steps", type: "array", required: true }, // Multi-step automation sequence
                { name: "current_step", type: "number" }, // Track progress through steps
                { name: "context_data", type: "object" }, // Extracted data between steps
                { name: "oauth_context", type: "object" }, // User's OAuth tokens for MCP calls
                { name: "status", type: "string" }, // "pending", "running", "complete", "error", "waiting_approval"
                { name: "result", type: "string" },
                { name: "artifacts", type: "array" } // Created documents, sent emails, scheduled meetings
            ]
            metadata: {
                tags: ["workflow", "execution", "personal", "multi-step"]
            }
        }
    ]
    
    // Core PoC Events: OAuth2 + Personal Automations
    events: [
        {
            id: "user_authentication"
            version: "1.0"
            type: "user"
            description: "User authenticates via OAuth2 and system discovers personal capabilities"
            events: [{
                id: "oauth_login"
                type: "user"
                description: "User logs in with Google OAuth2"
                triggers: [{
                    source: { type: "user", name: "login_button" }
                }]
                actions: [{
                    type: "create"
                    state: "user_auth"
                    fields: ["user_id", "email", "oauth_tokens", "connected_services"]
                }]
            }]
        },
        {
            id: "personal_discovery"
            version: "1.0"
            type: "system"
            description: "Discover user's personal MCP capabilities after authentication"
            events: [{
                id: "discover_personal_capabilities"
                type: "system"
                description: "Query user's connected services to discover available personal automations"
                triggers: [{
                    source: { type: "system", name: "post_oauth" }
                }]
                actions: [{
                    type: "create"
                    state: "personal_capabilities"
                    fields: ["user_id", "mcp_servers", "available_actions", "examples"]
                }]
            }]
        },
        // NOTE: Story Coaching Agent removed from PoC - deferred for future enhancement
        {
            id: "workflow_discovery_flow"
            version: "1.0"
            type: "user"
            description: "Multi-turn conversation for workflow pattern discovery"
            events: [{
                id: "start_workflow_discovery"
                type: "user"
                description: "User begins multi-turn conversation to discover workflow patterns"
                triggers: [{
                    source: { type: "user", name: "chat_input" }
                }]
                actions: [{
                    type: "create"
                    state: "workflow_discovery"
                    fields: ["session_id", "user_id", "conversation_history", "discovery_phase", "status"]
                }]
            }]
        },
        {
            id: "intent_analysis_flow"
            version: "1.0"
            type: "system"
            description: "Simplified intent analysis with 5 PoC parameters"
            events: [{
                id: "analyze_workflow_intent"
                type: "system"
                description: "Intent Analyst Agent processes workflow intent with simplified parameters"
                triggers: [{
                    source: { type: "system", name: "intent_analyst_agent" }
                }]
                actions: [{
                    type: "create"
                    state: "intent_analysis"
                    fields: ["user_id", "workflow_intent", "is_automation_request", "required_services", "can_fulfill", "missing_info", "next_action", "service_validation"]
                }]
            }]
        },
        {
            id: "deterministic_workflow_generation"
            version: "1.0"
            type: "system"
            description: "Generate deterministic CUE workflow with steps-based execution"
            events: [{
                id: "generate_deterministic_workflow"
                type: "system"
                description: "Workflow Generator Agent creates executable CUE workflow with sequential steps"
                triggers: [{
                    source: { type: "system", name: "workflow_generator_agent" }
                }]
                actions: [{
                    type: "create"
                    state: "deterministic_workflow"
                    fields: ["user_id", "workflow_name", "workflow_steps", "user_parameters", "service_bindings", "cue_specification", "execution_ready"]
                }]
            }]
        },
        {
            id: "intent_confirmation_flow"
            version: "1.0"
            type: "system"
            description: "System confirms understanding of user intent before generating workflow"
            events: [{
                id: "confirm_workflow_intent"
                type: "system"
                description: "Present workflow plan to user for approval before execution"
                triggers: [{
                    source: { type: "system", name: "intent_analyzer" }
                }]
                actions: [{
                    type: "create"
                    state: "intent_confirmation"
                    fields: ["user_id", "original_intent", "understood_workflow", "suggested_steps", "missing_data", "status"]
                }]
            }]
        },
        {
            id: "workflow_execution"
            version: "1.0"
            type: "system"
            description: "System executes approved RaC workflow specification deterministically"
            events: [{
                id: "execute_rac_workflow"
                type: "system"
                description: "Execute approved RaC specification using deterministic workflow engine"
                triggers: [{
                    source: { type: "system", name: "rac_executor" }
                }]
                actions: [{
                    type: "create"
                    state: "rac_execution"
                    fields: ["user_id", "workflow_name", "rac_specification", "execution_id", "status"]
                }]
            }]
        }
    ]
    
    // Core PoC Logic: OAuth2 + Personal Validation
    logic: [
        {
            id: "oauth_validation"
            appliesTo: "user_auth"
            rules: [
                {
                    if: "oauth_tokens != null && len(connected_services) > 0"
                    then: [{ error: "Invalid OAuth tokens or no connected services" }]
                },
                {
                    if: "session_expires != null"
                    then: [{ error: "Session expiration not set" }]
                }
            ]
            metadata: {
                tags: ["validation", "oauth2", "auth"]
            }
        },
        {
            id: "personal_capability_validation"
            appliesTo: "personal_capabilities"
            rules: [{
                if: "len(mcp_servers) > 0 && len(available_actions) > 0 && user_id != null"
                then: [{ error: "No personal capabilities discovered or missing user context" }]
            }]
            metadata: {
                tags: ["validation", "capabilities", "personal"]
            }
        },
        {
            id: "message_validation"
            appliesTo: "chat_session"
            rules: [{
                if: "user_message != null && user_message != ''"
                then: [{ error: "User message cannot be empty" }]
            }]
            metadata: {
                tags: ["validation", "input"]
            }
        },
        {
            id: "intent_confidence_check"
            appliesTo: "simple_intent"
            rules: [{
                if: "confidence == 'low'"
                then: [{ error: "Intent unclear - request clarification" }]
            }]
            metadata: {
                tags: ["validation", "confidence"]
            }
        },
        {
            id: "workflow_complexity_validation"
            appliesTo: "workflow_intent"
            rules: [
                {
                    if: "workflow_type in ['proposal_followup', 'meeting_prep', 'team_update', 'project_summary', 'prospect_followup']"
                    then: [{ error: "Workflow type not supported in PoC" }]
                },
                {
                    if: "confidence in ['high', 'medium'] && complexity != 'complex'"
                    then: [{ error: "Workflow too complex or confidence too low for PoC" }]
                }
            ]
            metadata: {
                tags: ["validation", "workflow", "complexity"]
            }
        }
    ]
    
    // PoC UI: OAuth2 + Personal Capabilities Interface
    ui: [
        {
            id: "oauth_login"
            description: "OAuth2 authentication interface"
            components: [{
                type: "oauth_button"
                fields: [
                    { name: "provider", type: "text", description: "OAuth2 provider (Google)" },
                    { name: "scopes", type: "list", description: "Requested permissions (gmail, docs, calendar)" }
                ]
                submitEvent: "oauth_login"
            }]
            metadata: {
                tags: ["auth", "oauth2"]
            }
        },
        {
            id: "personal_capability_showcase"
            description: "Display user's personal automation capabilities"
            components: [{
                type: "personal_capability_grid"
                fields: [
                    { name: "user_email", type: "text", description: "Authenticated user's email" },
                    { name: "connected_services", type: "list", description: "User's connected Google services" },
                    { name: "available_actions", type: "list", description: "Personal automations available" },
                    { name: "examples", type: "list", description: "Example phrases for user's context" }
                ]
            }]
            metadata: {
                tags: ["capabilities", "personal", "discovery"]
            }
        },
        {
            id: "guided_chat"
            description: "Chat interface with capability context"
            components: [{
                type: "chat_input"
                fields: [
                    { name: "message", type: "text", description: "User intent (guided by capabilities)" }
                ]
                submitEvent: "message_received"
            }, {
                type: "chat_display"
                fields: [
                    { name: "conversation", type: "list", description: "Chat history" },
                    { name: "status", type: "text", description: "Current processing status" },
                    { name: "suggested_actions", type: "list", description: "Quick action buttons" }
                ]
            }]
            metadata: {
                tags: ["poc", "guided"]
            }
        }
    ]
    
    // PoC Tests: OAuth2 + Personal Automations
    tests: [{
        id: "poc_oauth_personal"
        testCases: [
            {
                id: "oauth_login_flow"
                description: "User authenticates with Google OAuth2"
                input: {
                    event: "oauth_login"
                    data: {
                        provider: "google"
                        scopes: ["gmail", "docs", "calendar"]
                        user_email: "john@example.com"
                    }
                }
                expected: {
                    user_auth: {
                        user_id: "john@example.com"
                        email: "john@example.com"
                        connected_services: ["gmail", "docs", "calendar"]
                        oauth_tokens: {} // OAuth2 tokens
                    }
                }
            },
            {
                id: "personal_capability_discovery"
                description: "System discovers user's personal automation capabilities"
                input: {
                    event: "discover_personal_capabilities"
                    data: {
                        user_id: "john@example.com"
                        oauth_tokens: {} // Valid tokens
                    }
                }
                expected: {
                    personal_capabilities: {
                        user_id: "john@example.com"
                        mcp_servers: ["gmail", "docs", "calendar"]
                        available_actions: ["send_message", "create_document", "schedule_meeting"]
                        examples: [
                            "Follow up on the proposal I sent to ACME Corp last week",
                            "Prepare for my client meeting with Sarah tomorrow - create agenda and send invite", 
                            "Generate my weekly team update and send to stakeholders",
                            "Find all emails about the Q4 project and create a status summary",
                            "Schedule a follow-up call with the prospect who downloaded our whitepaper"
                        ]
                        status: "ready"
                    }
                }
            },
            {
                id: "guided_email_intent"
                description: "User selects email action after seeing capabilities"
                input: {
                    event: "message_received"
                    data: { user_message: "Send an email to john@example.com about the meeting" }
                }
                expected: {
                    simple_intent: {
                        detected_action: "send_message"
                        confidence: "high"
                    }
                }
            },
            {
                id: "unsupported_action"
                description: "User requests action not in capabilities"
                input: {
                    event: "message_received"
                    data: { user_message: "Book a flight to Paris" }
                }
                expected: {
                    simple_intent: {
                        confidence: "low"
                    }
                }
                expectError: "Action not supported in PoC"
            }
        ]
        metadata: {
            tags: ["poc", "coaching", "integration"]
        }
    }]
}
