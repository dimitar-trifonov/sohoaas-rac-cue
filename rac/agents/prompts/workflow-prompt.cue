package rac

// =============================================
// 🔹 WORKFLOW PROMPT SCHEMA - LLM FOCUSED
// =============================================
// Streamlined RaC schema for workflow generation
// Contains only essential structures without architecture/bindings

// =============================================
// 🔹 PARAMETER REFERENCE SPECIFICATION
// =============================================
// Documents the exact parameter types supported by the backend implementation

ParameterReferenceSpec: {
    version: "1.0.0"
    description: "Specification of parameter reference formats supported by SOHOAAS backend"
    
    supported_formats: {
        user_parameters: {
            format: "${user.parameter_name}"
            description: "References user-provided input parameters"
            validation: "Fully validated - parameter must exist in workflow user_parameters section"
            examples: [
                "${user.email_address}",
                "${user.folder_name}",
                "${user.search_query}"
            ]
            backend_support: "complete"
        }
        
        step_outputs: {
            format: "${steps.step_id.outputs.field_name}"
            description: "References outputs from previous workflow steps"
            validation: "Fully validated - step must exist, circular dependencies detected"
            examples: [
                "${steps.get_email.outputs.subject}",
                "${steps.create_folder.outputs.folder_id}",
                "${steps.list_messages.outputs.message_ids}"
            ]
            backend_support: "complete"
            dependency_tracking: "automatic"
        }
        
        environment_variables: {
            format: "${ENV_VAR_NAME}"
            description: "References system environment variables"
            validation: "Format validated only - assumed available at runtime"
            examples: [
                "${WORKSPACE_NAME}",
                "${USER_EMAIL}",
                "${API_BASE_URL}"
            ]
            backend_support: "basic"
            runtime_availability: "assumed"
        }
        
        computed_values: {
            format: "${computed.expression}"
            description: "References computed expressions (limited backend support)"
            validation: "Format validated only - execution deferred to runtime"
            examples: [
                "${computed.timestamp}",
                "${computed.current_user}"
            ]
            backend_support: "limited"
            execution_engine: "not_implemented"
            recommendation: "avoid_in_poc"
        }
    }
    
    unsupported_formats: {
        system_parameters: {
            format: "${CURRENT_DATE}, ${LAST_WEEK}"
            reason: "No system parameter resolution in backend"
            alternative: "Use environment variables or user parameters"
        }
        
        complex_computed_functions: {
            format: "${computed.function_name(args)}"
            reason: "No function execution engine implemented"
            alternative: "Use simple computed expressions or step outputs"
        }
        
        nested_parameter_sections: {
            format: "system_parameters: {...}, computed_parameters: {...}"
            reason: "JSON schema only supports user_parameters section"
            alternative: "Include all parameters in user_parameters section"
        }
    }
    
    validation_rules: {
        circular_dependencies: "Backend detects and rejects workflows with circular step dependencies"
        parameter_existence: "All user parameters must be defined in user_parameters section"
        step_references: "Referenced steps must exist and cannot reference themselves"
        format_compliance: "All parameter references must match exact format patterns"
    }
    
    best_practices: {
        prefer_user_parameters: "Use ${user.param} for all user-configurable values"
        use_step_outputs: "Chain workflow steps using ${steps.step_id.outputs.field}"
        limit_env_vars: "Only use environment variables for system-level configuration"
        avoid_computed: "Avoid computed expressions in PoC - limited backend support"
        explicit_dependencies: "LLM must set depends_on field when step uses ${steps.step_id.outputs.*}"
    }
}

// =============================================
// 🔹 CORE SCHEMA DEFINITIONS
// =============================================

#State: {
    id:        string & !=""
    type:      "object" | "array" | "primitive"
    fields?: [...{
        name:     string
        type:     string
        required?: bool | *false
        format?:  string
    }]
    metadata?: {
        tags?: [...string]
    }
}

#Event: {
    id:          string & !=""
    type:        "user" | "system"
    description: string
    triggers?: [...{
        source: { type: "user" | "system" | "custom", name: string }
        params?: { threshold?: number, delay?: number }
    }]
    actions?: [...{
        type: "validate" | "update" | "create" | "delete"
        state?: string
        fields?: [...string]
    }]
    metadata?: {
        tags?: [...string]
    }
}

#Logic: {
    id: string & !=""
    appliesTo?: string
    rules?: [...{
        if:   string
        then: [...{ error: string }]
    }]
    guards?:  [...string]
    effects?: [...string]
    metadata?: {
        tags?: [...string]
    }
}

#Test: {
    id: string & !=""
    testCases?: [...{
        id:          string & !=""
        description: string
        input?: {
            event?: string
            data?:  {...}
        }
        expected?:   {...}
        expectError?: string
    }]
    metadata?: {
        tags?: [...string]
    }
}

// =============================================
// 🔹 WORKFLOW GENERATION FOCUSED STATES
// =============================================

WorkflowGenerationStates: [
    {
        id: "user_intent"
        type: "object"
        fields: [
            { name: "user_message", type: "string", required: true },
            { name: "detected_actions", type: "array", required: true },
            { name: "confidence", type: "string", required: true },
            { name: "requires_clarification", type: "boolean" }
        ]
        metadata: {
            tags: ["intent", "analysis", "input"]
        }
    },
    {
        id: "user_capabilities"
        type: "object"
        fields: [
            { name: "available_services", type: "array", required: true },
            { name: "mcp_functions", type: "object", required: true },
            { name: "service_parameters", type: "object", required: true }
        ]
        metadata: {
            tags: ["capabilities", "mcp", "services"]
        }
    },
    {
        id: "workflow_steps"
        type: "object"
        fields: [
            { name: "step_id", type: "string", required: true },
            { name: "action", type: "string", required: true },
            { name: "parameters", type: "object", required: true },
            { name: "capability_status", type: "string", required: true }, // "capable", "not_capable", "needs_resolution"
            { name: "depends_on", type: "array" }
        ]
        metadata: {
            tags: ["steps", "capabilities", "mapping"]
        }
    },
    {
        id: "workflow_json"
        type: "object"
        fields: [
            { name: "name", type: "string", required: true },
            { name: "description", type: "string", required: true },
            { name: "steps", type: "array", required: true },
            { name: "user_parameters", type: "object", required: true },
            { name: "execution_config", type: "object", required: true }
        ]
        metadata: {
            tags: ["generation", "json", "output"]
        }
    }
]

// =============================================
// 🔹 WORKFLOW GENERATION FOCUSED EVENTS
// =============================================

WorkflowGenerationEvents: [
    {
        id: "intent_received"
        type: "input"
        description: "User intent and capabilities received for workflow generation"
        triggers: [{
            source: { type: "system", name: "workflow_generator" }
        }]
        actions: [{
            type: "create"
            state: "user_intent"
            fields: ["user_message", "detected_actions", "confidence"]
        }]
        metadata: {
            tags: ["input", "intent"]
        }
    },
    {
        id: "capabilities_mapped"
        type: "system"
        description: "User capabilities mapped to intent actions"
        triggers: [{
            source: { type: "system", name: "capability_mapper" }
        }]
        actions: [{
            type: "create"
            state: "workflow_steps"
            fields: ["step_id", "action", "capability_status"]
        }]
        metadata: {
            tags: ["mapping", "capabilities"]
        }
    },
    {
        id: "parameters_extracted"
        type: "system"
        description: "Parameters extracted from MCP service schemas"
        triggers: [{
            source: { type: "system", name: "parameter_extractor" }
        }]
        actions: [{
            type: "update"
            state: "workflow_steps"
            fields: ["parameters"]
        }]
        metadata: {
            tags: ["parameters", "mcp"]
        }
    },
    {
        id: "workflow_json_generated"
        type: "output"
        description: "Complete JSON workflow generated"
        triggers: [{
            source: { type: "system", name: "json_generator" }
        }]
        actions: [{
            type: "create"
            state: "workflow_json"
            fields: ["name", "description", "steps", "user_parameters", "execution_config"]
        }]
        metadata: {
            tags: ["output", "json", "workflow"]
        }
    }
]

// =============================================
// 🔹 WORKFLOW GENERATION LOGIC
// =============================================

WorkflowGenerationLogic: [
    {
        id: "intent_analysis"
        appliesTo: "user_intent"
        rules: [
            {
                if: "confidence == 'low'"
                then: [{ error: "Intent unclear, requires clarification" }]
            },
            {
                if: "detected_actions == []"
                then: [{ error: "No actionable steps detected in user message" }]
            }
        ]
        guards: ["user_message != null", "confidence != null"]
        effects: ["trigger_capability_mapping"]
        metadata: {
            tags: ["intent", "analysis", "validation"]
        }
    },
    {
        id: "capability_mapping"
        appliesTo: "workflow_steps"
        rules: [
            {
                if: "capability_status == 'not_capable'"
                then: [{ error: "Step requires unavailable service - needs frontend resolution" }]
            },
            {
                if: "action not in available_services"
                then: [{ error: "Action not supported by user capabilities" }]
            }
        ]
        guards: ["action != null", "capability_status != null"]
        effects: ["trigger_parameter_extraction"]
        metadata: {
            tags: ["capabilities", "mapping", "validation"]
        }
    },
    {
        id: "parameter_extraction"
        appliesTo: "workflow_steps"
        rules: [
            {
                if: "parameters == {}"
                then: [{ error: "No parameters extracted from MCP service schema" }]
            },
            {
                if: "required_parameters missing"
                then: [{ error: "Required parameters not provided by user or system" }]
            }
        ]
        guards: ["mcp_functions != null", "service_parameters != null"]
        effects: ["trigger_json_generation"]
        metadata: {
            tags: ["parameters", "mcp", "extraction"]
        }
    },
    {
        id: "json_generation"
        appliesTo: "workflow_json"
        rules: [
            {
                if: "steps == []"
                then: [{ error: "No valid workflow steps generated" }]
            },
            {
                if: "user_parameters missing for required fields"
                then: [{ error: "Missing user parameter definitions" }]
            }
        ]
        guards: ["name != null", "description != null", "steps != null"]
        effects: ["workflow_generation_complete"]
        metadata: {
            tags: ["json", "generation", "validation"]
        }
    }
]

// =============================================
// 🔹 WORKFLOW GENERATION TESTS
// =============================================

WorkflowGenerationTests: [
    {
        id: "intent_to_steps_mapping"
        testCases: [{
            id: "email_intent_mapping"
            description: "Map email intent to workflow steps with capability checking"
            input: {
                user_intent: {
                    user_message: "Send an email to john@example.com about the meeting"
                    detected_actions: ["send_message"]
                    confidence: "high"
                }
                user_capabilities: {
                    available_services: ["gmail", "docs", "calendar"]
                    mcp_functions: {
                        "gmail.send_message": {
                            required_fields: ["to", "subject", "body"]
                            optional_fields: ["attachments"]
                        }
                    }
                }
            }
            expected: {
                workflow_steps: [{
                    step_id: "send_email"
                    action: "gmail.send_message"
                    capability_status: "capable"
                    parameters: {
                        to: "${user.to}"
                        subject: "${user.subject}"
                        body: "${user.body}"
                    }
                    depends_on: []
                }]
            }
        }]
        metadata: {
            tags: ["mapping", "capabilities", "intent"]
        }
    },
    {
        id: "capability_gap_handling"
        testCases: [{
            id: "unsupported_service"
            description: "Handle steps requiring unavailable services"
            input: {
                user_intent: {
                    user_message: "Book a flight to Paris"
                    detected_actions: ["book_flight"]
                    confidence: "high"
                }
                user_capabilities: {
                    available_services: ["gmail", "docs", "calendar"]
                    mcp_functions: {}
                }
            }
            expected: {
                workflow_steps: [{
                    step_id: "book_flight"
                    action: "travel.book_flight"
                    capability_status: "not_capable"
                    parameters: {}
                    depends_on: []
                }]
            }
        }]
        metadata: {
            tags: ["capability_gaps", "resolution", "frontend"]
        }
    },
    {
        id: "parameter_extraction"
        testCases: [{
            id: "mcp_parameter_mapping"
            description: "Extract parameters from MCP service schemas"
            input: {
                workflow_steps: [{
                    step_id: "send_email"
                    action: "gmail.send_message"
                    capability_status: "capable"
                }]
                user_capabilities: {
                    service_parameters: {
                        "gmail.send_message": {
                            to: { type: "string", required: true }
                            subject: { type: "string", required: true }
                            body: { type: "string", required: true }
                            attachments: { type: "array", required: false }
                        }
                    }
                }
            }
            expected: {
                workflow_steps: [{
                    step_id: "send_email"
                    action: "gmail.send_message"
                    capability_status: "capable"
                    parameters: {
                        to: "${user.to}"
                        subject: "${user.subject}"
                        body: "${user.body}"
                    }
                }]
                user_parameters: {
                    to: { type: "string", required: true }
                    subject: { type: "string", required: true }
                    body: { type: "string", required: true }
                }
            }
        }]
        metadata: {
            tags: ["parameters", "mcp", "extraction"]
        }
    },
    {
        id: "json_workflow_generation"
        testCases: [{
            id: "complete_workflow_json"
            description: "Generate complete JSON workflow from processed steps"
            input: {
                user_intent: {
                    user_message: "Send an email to john@example.com about the meeting"
                }
                workflow_steps: [{
                    step_id: "send_email"
                    action: "gmail.send_message"
                    capability_status: "capable"
                    parameters: {
                        to: "${user.to}"
                        subject: "${user.subject}"
                        body: "${user.body}"
                    }
                }]
            }
            expected: {
                workflow_json: {
                    name: "Send Email Workflow"
                    description: "Send an email message to specified recipient"
                    steps: [{
                        id: "send_email"
                        action: "gmail.send_message"
                        parameters: {
                            to: "${user.to}"
                            subject: "${user.subject}"
                            body: "${user.body}"
                        }
                    }]
                    user_parameters: {
                        to: { type: "string", required: true }
                        subject: { type: "string", required: true }
                        body: { type: "string", required: true }
                    }
                    execution_config: {
                        mode: "sequential"
                        timeout: "5m"
                        environment: "development"
                    }
                }
            }
        }]
        metadata: {
            tags: ["json", "generation", "complete"]
        }
    },
    {
        id: "step_dependent_parameters"
        testCases: [{
            id: "email_with_document_creation"
            description: "Multi-step workflow where later steps depend on outputs from earlier steps"
            input: {
                user_intent: {
                    user_message: "Get the oldest email from john@example.com and create a document with its content"
                    detected_actions: ["get_message", "create_document"]
                    confidence: "high"
                }
                user_capabilities: {
                    available_services: ["gmail", "docs"]
                    mcp_functions: {
                        "gmail.send_message": {
                            required_fields: ["to", "subject", "body"]
                            output_fields: ["message_id", "thread_id", "status", "sent_at", "to", "subject"]
                        }
                        "docs.create_document": {
                            required_fields: ["title", "content"]
                            optional_fields: ["folder_id"]
                        }
                    }
                }
            }
            expected: {
                workflow_json: {
                    name: "Send Email and Document Workflow"
                    description: "Send an email and create a confirmation document with response details"
                    steps: [
                        {
                            id: "send_email"
                            action: "gmail.send_message"
                            parameters: {
                                to: "${user.recipient_email}"
                                subject: "${user.email_subject}"
                                body: "${user.email_body}"
                            }
                        },
                        {
                            id: "create_doc"
                            action: "docs.create_document"
                            parameters: {
                                title: "Email Confirmation: ${steps.send_email.outputs.subject}"
                                content: "Email sent to ${steps.send_email.outputs.to} with message ID: ${steps.send_email.outputs.message_id}"
                            }
                            depends_on: ["send_email"]
                        }
                    ]
                    user_parameters: {
                        recipient_email: { type: "string", required: true }
                        email_subject: { type: "string", required: true }
                        email_body: { type: "string", required: true }
                    }
                    execution_config: {
                        mode: "sequential"
                        timeout: "10m"
                        environment: "development"
                    }
                }
            }
        }]
        metadata: {
            tags: ["step_dependencies", "parameter_chaining", "multi_step"]
        }
    },
    {
        id: "environment_dependent_parameters"
        testCases: [{
            id: "folder_organization_with_env_vars"
            description: "Workflow using environment variables and step dependencies"
            input: {
                user_intent: {
                    user_message: "Create a folder with my workspace name and move emails there"
                    detected_actions: ["create_folder", "move_files"]
                    confidence: "high"
                }
                user_capabilities: {
                    available_services: ["drive", "gmail"]
                    mcp_functions: {
                        "drive.create_folder": {
                            required_fields: ["name", "parent_folder"]
                            output_fields: ["folder_id", "folder_url"]
                        }
                        "gmail.list_messages": {
                            required_fields: ["query"]
                            output_fields: ["message_ids", "count"]
                        }
                        "gmail.move_messages": {
                            required_fields: ["message_ids", "destination_folder"]
                        }
                    }
                }
            }
            expected: {
                workflow_json: {
                    name: "Workspace Email Organization"
                    description: "Create workspace folder and organize emails"
                    steps: [
                        {
                            id: "create_workspace_folder"
                            action: "drive.create_folder"
                            parameters: {
                                name: "${WORKSPACE_NAME} - ${user.folder_suffix}"
                                parent_folder: "${user.base_folder}"
                            }
                        },
                        {
                            id: "find_emails"
                            action: "gmail.list_messages"
                            parameters: {
                                query: "${user.search_query}"
                            }
                        },
                        {
                            id: "move_emails"
                            action: "gmail.move_messages"
                            parameters: {
                                message_ids: "${steps.find_emails.outputs.message_ids}"
                                destination_folder: "${steps.create_workspace_folder.outputs.folder_id}"
                            }
                            depends_on: ["create_workspace_folder", "find_emails"]
                        }
                    ]
                    user_parameters: {
                        base_folder: { type: "string", required: true }
                        folder_suffix: { type: "string", required: true }
                        search_query: { type: "string", required: true }
                    }
                    execution_config: {
                        mode: "sequential"
                        timeout: "15m"
                        environment: "development"
                    }
                }
            }
        }]
        metadata: {
            tags: ["environment_variables", "step_dependencies", "realistic"]
        }
    },
    {
        id: "complex_dependency_chain"
        testCases: [{
            id: "multi_service_workflow"
            description: "Complex workflow with multiple step dependencies using only supported parameter types"
            input: {
                user_intent: {
                    user_message: "Find emails from my manager, create a document with the content, and schedule a follow-up meeting"
                    detected_actions: ["list_messages", "create_document", "create_event"]
                    confidence: "high"
                }
                user_capabilities: {
                    available_services: ["gmail", "docs", "calendar"]
                    mcp_functions: {
                        "gmail.list_messages": {
                            required_fields: ["query"]
                            output_fields: ["message_ids", "messages", "count"]
                        }
                        "docs.create_document": {
                            required_fields: ["title"]
                            output_fields: ["document_id", "title", "url", "status", "revision_id", "created_at"]
                        }
                        "calendar.create_event": {
                            required_fields: ["title", "start_time", "end_time"]
                            optional_fields: ["description", "attendees"]
                            output_fields: ["event_id", "title", "start_time", "end_time", "status", "html_link", "created_at", "updated_at"]
                        }
                    }
                }
            }
            expected: {
                workflow_json: {
                    name: "Manager Email Processing and Follow-up"
                    description: "Process manager emails, create document, and schedule follow-up"
                    steps: [
                        {
                            id: "find_manager_emails"
                            action: "gmail.list_messages"
                            parameters: {
                                query: "from:${user.manager_email}"
                            }
                        },
                        {
                            id: "create_summary_doc"
                            action: "docs.create_document"
                            parameters: {
                                title: "${user.document_title}"
                                content: "${steps.find_manager_emails.outputs.messages}"
                            }
                            depends_on: ["find_manager_emails"]
                        },
                        {
                            id: "schedule_followup"
                            action: "calendar.create_event"
                            parameters: {
                                title: "${user.meeting_title}"
                                start_time: "${user.meeting_time}"
                                end_time: "${user.meeting_end_time}"
                                description: "Review document: ${steps.create_summary_doc.outputs.url}"
                                attendees: ["${user.manager_email}"]
                            }
                            depends_on: ["create_summary_doc"]
                        }
                    ]
                    user_parameters: {
                        manager_email: { type: "string", required: true }
                        document_title: { type: "string", required: true }
                        meeting_title: { type: "string", required: true }
                        meeting_time: { type: "string", required: true }
                        meeting_end_time: { type: "string", required: true }
                    }
                    execution_config: {
                        mode: "sequential"
                        timeout: "20m"
                        environment: "development"
                    }
                }
            }
        }]
        metadata: {
            tags: ["complex_dependencies", "multi_service", "realistic_backend_support"]
        }
    }
]
