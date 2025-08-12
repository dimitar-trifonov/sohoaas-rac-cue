package tests

import (
    "github.com/sohoaas/rac/schemas"
)

// Test for Workflow Generation with Real MCP Services
WorkflowGenerationTest: {
    id: "workflow_generation_with_mcp_services"
    version: "1.0"
    
    // Setup: Real MCP services from curl -s http://localhost:8080/api/services | jq '.providers'
    setup: {
        available_services: {
            providers: {
                workspace: {
                    calendar: {
                        description: "Manage calendar events using Google Calendar API"
                        functions: ["create_event", "update_event", "delete_event", "get_event", "list_events"]
                    }
                    docs: {
                        description: "Create, edit, and manage documents using Google Docs API"
                        functions: ["create_document", "get_document", "insert_text", "update_document", "batch_update"]
                    }
                    drive: {
                        description: "Store, share, and manage files using Google Drive API"
                        functions: ["create_folder", "upload_file", "list_files", "get_file", "move_file", "share_file"]
                    }
                    gmail: {
                        description: "Send, receive, and manage emails using Gmail API"
                        functions: ["send_message", "get_message", "list_messages", "search_messages"]
                    }
                }
            }
        }
        user_context: {
            user_id: "test_user_123"
            connected_services: ["workspace"]
        }
    }
    
    testCases: [
        {
            id: "test_document_creation_workflow"
            description: "Test generation of document creation workflow with team notification"
            setup: {
                workflow_pattern: "I'd like to automate the workflow for our weekly team meeting. Each Monday morning, a Google Calendar event should be created, a shared Google Doc with the agenda should be prepared, and the team should receive an email notification with a link to the document. After the meeting, we want to auto-generate a summary and save it to Google Drive."
                validated_triggers: {
                    trigger_type: "manual"
                    description: "User initiates document creation"
                }
                validated_actions: [
                    "create_calendar_event",
                    "create_agenda_document", 
                    "send_email_notification",
                    "create_meeting_folder"
                ]
                required_services: ["workspace"]
                workflow_parameters: {
                    meeting_date: "string"
                    meeting_time: "string"
                    team_email: "string"
                }
            }
            input: {
                event: "generate_workflow"
                data: {
                    workflow_pattern: "I'd like to automate the workflow for our weekly team meeting. Each Monday morning, a Google Calendar event should be created, a shared Google Doc with the agenda should be prepared, and the team should receive an email notification with a link to the document. After the meeting, we want to auto-generate a summary and save it to Google Drive."
                    validated_triggers: {
                        trigger_type: "manual"
                        description: "User initiates weekly meeting automation"
                    }
                    validated_actions: [
                        "create_calendar_event",
                        "create_agenda_document", 
                        "send_email_notification",
                        "create_meeting_folder"
                    ]
                    required_services: ["workspace"]
                    workflow_parameters: {
                        meeting_date: "string"
                        meeting_time: "string"
                        team_email: "string"
                    }
                }
            }
            expected: {
                // Simplified expectations - LLM is not deterministic, so we check structure not exact content
                workflow_structure_requirements: {
                    must_have_version: true
                    must_have_name: true
                    must_have_description: true
                    must_have_steps: true
                    must_have_user_parameters: true
                    must_have_services: true
                }
                step_requirements: {
                    minimum_steps: 4
                    required_step_fields: ["id", "name", "service", "action", "parameters"]
                    must_use_workspace_services: true
                    must_have_dependencies: true // At least one step should depend on another
                }
                parameter_requirements: {
                    minimum_parameters: 3
                    required_parameter_fields: ["type", "required", "prompt"]
                    must_include_meeting_date: true
                    must_include_meeting_time: true
                    must_include_team_email: true
                }
                service_requirements: {
                    must_use_workspace_provider: true
                    required_workspace_services: ["drive", "docs", "calendar", "gmail"]
                    no_unavailable_services: true // Should not reference slack or other unavailable services
                }
                data_flow_requirements: {
                    must_use_user_parameters: true // ${user.param_name}
                    should_use_step_outputs: true // ${steps.step_id.outputs.field}
                }
            }
            notes: "Tests complex weekly meeting automation workflow with calendar, docs, drive, and email integration"
        },
        {
            id: "test_meeting_preparation_workflow"
            description: "Test generation of meeting preparation workflow with calendar and document creation"
            setup: {
                workflow_pattern: "Prepare for weekly team meeting: create agenda document and calendar invite"
                validated_actions: ["create_agenda_document", "create_calendar_event", "send_email_reminder"]
                required_services: ["workspace"]
            }
            input: {
                event: "generate_workflow"
                data: {
                    workflow_pattern: "Prepare for weekly team meeting: create agenda document and calendar invite"
                    validated_actions: ["create_agenda_document", "create_calendar_event", "send_email_reminder"]
                    required_services: ["workspace"]
                    workflow_parameters: {
                        meeting_title: "string"
                        meeting_date: "string"
                        attendees: "array"
                    }
                }
            }
            expected: {
                step_requirements: {
                    minimum_steps: 3
                    must_use_workspace_services: ["docs", "calendar", "gmail"]
                    must_have_sequential_dependencies: true
                }
                parameter_requirements: {
                    must_include_meeting_title: true
                    must_include_meeting_date: true
                }
            }
            notes: "Tests sequential workflow with workspace services (docs → calendar → gmail)"
        },
        {
            id: "test_client_onboarding_workflow"
            description: "Test generation of comprehensive client onboarding workflow with multiple services"
            setup: {
                workflow_pattern: "When a new client signs a contract, I want to automatically: (1) create a dedicated Google Drive folder, (2) generate a welcome document from a template, (3) share the folder with the client, and (4) send them a personalized welcome email. Internally, I also want to notify the account manager with separate email."
                validated_triggers: {
                    trigger_type: "manual"
                    description: "User initiates client onboarding process"
                }
                validated_actions: [
                    "create_client_folder",
                    "generate_welcome_document",
                    "share_folder_with_client",
                    "send_client_welcome_email",
                    "notify_account_manager"
                ]
                required_services: ["workspace"]
                workflow_parameters: {
                    client_name: "string"
                    client_email: "string"
                    account_manager_email: "string"
                    contract_type: "string"
                }
            }
            input: {
                event: "generate_workflow"
                data: {
                    workflow_pattern: "When a new client signs a contract, I want to automatically: (1) create a dedicated Google Drive folder, (2) generate a welcome document from a template, (3) share the folder with the client, and (4) send them a personalized welcome email. Internally, I also want to notify the account manager with separate email."
                    validated_triggers: {
                        trigger_type: "manual"
                        description: "User initiates client onboarding process"
                    }
                    validated_actions: [
                        "create_client_folder",
                        "generate_welcome_document",
                        "share_folder_with_client",
                        "send_client_welcome_email",
                        "notify_account_manager"
                    ]
                    required_services: ["workspace"]
                    workflow_parameters: {
                        client_name: "string"
                        client_email: "string"
                        account_manager_email: "string"
                        contract_type: "string"
                    }
                }
            }
            expected: {
                workflow_structure_requirements: {
                    must_have_version: true
                    must_have_name: true
                    must_have_description: true
                    must_have_steps: true
                    must_have_user_parameters: true
                    must_have_services: true
                }
                step_requirements: {
                    minimum_steps: 5
                    required_step_fields: ["id", "name", "service", "action", "parameters"]
                    must_use_workspace_services: true
                    must_have_dependencies: true
                }
                parameter_requirements: {
                    minimum_parameters: 4
                    required_parameter_fields: ["type", "required", "prompt"]
                    must_include_client_name: true
                    must_include_client_email: true
                    must_include_account_manager_email: true
                    must_include_contract_type: true
                }
                service_requirements: {
                    must_use_workspace_provider: true
                    required_workspace_services: ["drive", "docs", "gmail"]
                    no_unavailable_services: true
                }
                data_flow_requirements: {
                    must_use_user_parameters: true
                    should_use_step_outputs: true
                    must_share_folder_with_client: true
                    must_send_separate_emails: true
                }
            }
            notes: "Tests complex client onboarding workflow with folder creation, document generation, sharing, and dual email notifications"
        },
        {
            id: "test_exact_mcp_service_catalog_validation"
            description: "Test that Workflow Generator uses EXACT MCP service catalog from Agent Manager"
            setup: {
                workflow_pattern: "Create project folder, generate document, and notify team"
                validated_triggers: {
                    trigger_type: "manual"
                    description: "User initiates project setup workflow"
                }
                workflow_parameters: {
                    project_name: "Alpha Project"
                    document_title: "Alpha Project Requirements"
                    team_email: "team@company.com"
                }
                // Simulated Agent Manager event with MCP catalog
                agent_manager_event: {
                    event_name: "workflow_generation_requested"
                    available_services: {
                        providers: {
                            workspace: {
                                drive: {
                                    functions: ["create_folder", "upload_file", "share_file", "move_file"]
                                }
                                docs: {
                                    functions: ["create_document", "get_document", "insert_text"]
                                }
                                gmail: {
                                    functions: ["send_message", "list_messages", "search_messages"]
                                }
                            }
                        }
                    }
                }
            }
            expectations: {
                should_succeed: true
                mcp_catalog_compliance: {
                    uses_exact_service_paths: ["workspace.drive", "workspace.docs", "workspace.gmail"]
                    uses_exact_function_names: ["create_folder", "create_document", "send_message"]
                    no_guessed_services: true
                    no_invented_functions: true
                    service_path_format: "workspace.service_name"
                }
                service_bindings_validation: {
                    exact_provider_references: ["workspace"]
                    correct_service_structure: true
                    matches_catalog_exactly: true
                }
                workflow_structure: {
                    steps_use_catalog_services_only: true
                    function_names_from_catalog: true
                    no_generic_service_names: true
                }
            }
        },
        {
            id: "test_weekly_review_reminder_workflow"
            description: "Test enhanced MCP schema discovery with complex parameter validation and data flow"
            setup: {
                workflow_pattern: "Every Friday at 4 PM, I want to send a review reminder to the team. The reminder should include a link to a shared Google Doc (weekly report draft). The document should be copied from a template into a dated folder in Drive. The Calendar should show a 30-minute 'Review & Comment' session. The Gmail notification should go to all reviewers."
                validated_triggers: {
                    trigger_type: "scheduled"
                    description: "Weekly Friday 4 PM review reminder automation"
                    schedule: "0 16 * * 5" // Every Friday at 4 PM
                }
                workflow_parameters: {
                    template_document_id: "string"
                    reviewers_emails: "array"
                    team_lead_email: "string"
                    review_duration: "number"
                }
                // Enhanced MCP catalog with complete parameter schemas
                available_services: {
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
                                copy_file: {
                                    parameters: {
                                        file_id: {type: "string", required: true, description: "Source file ID to copy"}
                                        name: {type: "string", required: true, description: "Name for copied file"}
                                        parent_id: {type: "string", required: false, description: "Destination folder ID"}
                                    }
                                    outputs: {
                                        file_id: {type: "string", description: "Copied file ID"}
                                        file_url: {type: "string", description: "File sharing URL"}
                                    }
                                }
                                share_file: {
                                    parameters: {
                                        file_id: {type: "string", required: true, description: "File ID to share"}
                                        email: {type: "string", required: true, validation: "email", description: "Email to share with"}
                                        role: {type: "string", required: true, enum: ["reader", "writer", "commenter"], description: "Permission level"}
                                    }
                                    outputs: {
                                        permission_id: {type: "string", description: "Permission ID"}
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
                        }
                    }
                }
            }
            expectations: {
                should_succeed: true
                enhanced_schema_validation: {
                    uses_exact_parameter_schemas: true
                    validates_required_parameters: ["name", "file_id", "summary", "to", "subject", "body"]
                    validates_optional_parameters: ["parent_id", "end", "attendees", "cc", "bcc"]
                    applies_validation_rules: {
                        email_validation: ["email", "cc", "bcc"]
                        enum_validation: ["role"]
                        type_validation: ["string", "object", "array"]
                    }
                    uses_default_values: {
                        parent_id: "root"
                        role: "commenter"
                    }
                }
                workflow_structure: {
                    steps_count: {min: 6}
                    sequential_dependencies: true
                    data_flow_validation: {
                        folder_id_propagation: true
                        file_url_in_email: true
                        event_url_in_email: true
                    }
                }
                parameter_collection: {
                    template_document_id: {type: "string", required: true, prompt: "Template document ID?"}
                    reviewers_emails: {type: "array", required: true, validation: "email_array", prompt: "Reviewer email addresses?"}
                    team_lead_email: {type: "string", required: true, validation: "email", prompt: "Team lead email?"}
                    review_duration: {type: "number", required: false, default: 30, prompt: "Review session duration (minutes)?"}
                }
                service_bindings: {
                    exact_mcp_paths: ["workspace.drive", "workspace.calendar", "workspace.gmail"]
                    complete_parameter_schemas: true
                    validation_rules_applied: true
                }
            }
            notes: "Tests enhanced MCP schema discovery with complex parameter validation, scheduled triggers, and multi-service data flow"
        },
        {
            id: "test_invalid_service_workflow"
            description: "Test error handling when required service is not available"
            input: {
                event: "generate_workflow"
                data: {
                    workflow_pattern: "Send message to Slack channel"
                    required_services: ["slack"] // slack not available in workspace
                }
            }
            expectError: "Service 'slack' is not available in workspace provider"
            notes: "Tests error handling for services not available in workspace provider"
        },
        {
            id: "test_circular_dependency_detection"
            description: "Test detection of circular dependencies in workflow steps"
            input: {
                event: "generate_workflow"
                data: {
                    workflow_pattern: "Complex workflow with potential circular dependencies"
                    validated_actions: ["step_a", "step_b", "step_c"]
                    // This would create: step_a depends on step_c, step_c depends on step_b, step_b depends on step_a
                }
            }
            expectError: "Circular dependency detected in workflow steps"
            notes: "Tests validation of step dependencies for circular references"
        }
    ]
    
    metadata: {
        createdBy: "workflow_generator_test"
        tags: ["workflow_generation", "mcp_services", "deterministic_workflow", "validation", "client_onboarding"]
        description: "Comprehensive test suite for workflow generation with real MCP service integration"
    }
}

// LLM Test Interpreter Instructions
TestInstructions: {
    description: "Instructions for LLM to execute this test as an interpreter"
    execution_steps: [
        "1. Load available MCP services from setup.available_services",
        "2. For each test case, simulate the workflow generation process",
        "3. Validate that generated workflows conform to #DeterministicWorkflow schema",
        "4. Check that all required services are available in the setup",
        "5. Verify step dependencies create a valid execution order (no circular deps)",
        "6. Validate user parameters have proper types and validation rules",
        "7. Ensure service bindings match available MCP services",
        "8. For error test cases, verify expected errors are properly detected",
        "9. Report test results with pass/fail status and detailed feedback"
    ]
    validation_rules: [
        "Generated CUE must be syntactically valid",
        "All referenced services must exist in available_services",
        "Step dependencies must form a directed acyclic graph (DAG)",
        "User parameters must have required fields: type, required, prompt",
        "Service bindings must include proper oauth_scopes",
        "Parameter references (${user.param}, ${steps.id.outputs.field}) must be valid"
    ]
}

// Expected Test Results Schema
TestResults: {
    test_id: string
    status: "PASS" | "FAIL" | "ERROR"
    execution_time: string
    details: {
        generated_workflow?: string
        validation_errors?: [...string]
        dependency_graph?: [...string]
        service_validation?: {
            available: [...string]
            missing: [...string]
        }
    }
    notes?: string
}
