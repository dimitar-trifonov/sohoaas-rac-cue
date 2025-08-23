package services

import "../schemas.cue"

// =============================================
// 🔹 WORKFLOW EXECUTOR SERVICE - STEP-BY-STEP EXECUTION
// =============================================
// Purpose: Execute validated CUE workflows with MCP service calls and result aggregation

WorkflowExecutorService: {
    version: "1.0.0"
    type: "deterministic_service"
    poc_status: "essential" // Core value: Reliable step-by-step workflow execution
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        {
            id: "workflow_execution"
            type: "object"
            fields: [
                { name: "user_id", type: "string", required: true },
                { name: "execution_id", type: "string", required: true },
                { name: "workflow_specification", type: "string", required: true }, // Validated CUE workflow
                { name: "execution_parameters", type: "object", required: true }, // Collected user parameters
                { name: "current_step", type: "string" }, // Currently executing step ID
                { name: "completed_steps", type: "array" }, // Successfully completed steps
                { name: "step_outputs", type: "object" }, // Outputs from completed steps
                { name: "execution_log", type: "array", required: true }, // Detailed execution trace
                { name: "execution_start_time", type: "string" },
                { name: "execution_end_time", type: "string" },
                { name: "status", type: "string", required: true }, // "preparing", "running", "completed", "failed", "paused"
                { name: "result", type: "object" } // Final execution results
            ]
            metadata: {
                tags: ["execution", "mcp_calls", "step_by_step", "results"]
            }
        }
    ]
    
    // LAYER 2: EVENTS (When things happen)
    events: [
        {
            id: "validated_workflow_received"
            version: "1.0"
            type: "input"
            description: "Validated CUE workflow received from validator"
            triggers: ["prepare_execution"]
            data_schema: {
                cue_workflow: "string"
                validation_results: "object"
                user_id: "string"
            }
        },
        {
            id: "execution_started"
            version: "1.0"
            type: "internal"
            description: "Workflow execution has begun"
            triggers: ["execute_next_step"]
            data_schema: {
                execution_id: "string"
                total_steps: "number"
                execution_parameters: "object"
            }
        },
        {
            id: "step_completed"
            version: "1.0"
            type: "internal"
            description: "Individual workflow step completed successfully"
            triggers: ["execute_next_step", "complete_execution"]
            data_schema: {
                step_id: "string"
                step_outputs: "object"
                execution_time_ms: "number"
            }
        },
        {
            id: "execution_completed"
            version: "1.0"
            type: "output"
            description: "Workflow execution completed successfully"
            data_schema: {
                execution_id: "string"
                final_results: "object"
                execution_summary: "object"
                total_execution_time_ms: "number"
            }
        },
        {
            id: "execution_failed"
            version: "1.0"
            type: "error"
            description: "Workflow execution failed at a specific step"
            data_schema: {
                failed_step: "string"
                error_details: "object"
                completed_steps: "array"
                recovery_options: "array"
            }
        }
    ]
    
    // LAYER 3: LOGIC (How it works)
    logic: [
        {
            id: "prepare_execution"
            type: "execution_preparation"
            description: "Prepare workflow for execution with parameter collection"
            steps: [
                {
                    name: "generate_execution_id"
                    description: "Create unique execution ID for tracking"
                    action: "execution.generate_id"
                },
                {
                    name: "collect_user_parameters"
                    description: "Prompt user for required parameters"
                    action: "execution.collect_parameters"
                    validation: "All required parameters must be provided"
                },
                {
                    name: "resolve_step_dependencies"
                    description: "Calculate optimal step execution order"
                    action: "execution.resolve_dependencies"
                },
                {
                    name: "initialize_execution_context"
                    description: "Set up execution environment and logging"
                    action: "execution.initialize_context"
                }
            ]
            output: {
                state: "workflow_execution"
                fields: ["execution_id", "execution_parameters", "status"]
            }
        },
        {
            id: "execute_next_step"
            type: "step_execution"
            description: "Execute the next workflow step with MCP service calls"
            steps: [
                {
                    name: "prepare_step_inputs"
                    description: "Resolve all step inputs from parameters and previous outputs"
                    action: "execution.prepare_inputs"
                    input_resolution: [
                        "${user.param} → execution_parameters[param]",
                        "$(step.outputs.field) → step_outputs[step][field]"
                    ]
                },
                {
                    name: "execute_mcp_service_call"
                    description: "Make authenticated call to MCP service"
                    action: "mcp.execute_service_call"
                    requirements: [
                        "Valid OAuth2 token for service",
                        "Proper service endpoint and action",
                        "Correctly formatted input parameters"
                    ]
                },
                {
                    name: "process_step_outputs"
                    description: "Extract and store step outputs for dependent steps"
                    action: "execution.process_outputs"
                },
                {
                    name: "log_step_execution"
                    description: "Record detailed execution information"
                    action: "execution.log_step"
                    logged_data: [
                        "step_id", "execution_time", "inputs", "outputs", "service_response"
                    ]
                }
            ]
            output: {
                state: "workflow_execution"
                fields: ["current_step", "completed_steps", "step_outputs", "execution_log"]
            }
        },
        {
            id: "complete_execution"
            type: "execution_completion"
            description: "Finalize workflow execution and aggregate results"
            steps: [
                {
                    name: "aggregate_final_results"
                    description: "Combine all step outputs into final result"
                    action: "execution.aggregate_results"
                },
                {
                    name: "generate_execution_summary"
                    description: "Create summary of execution performance and results"
                    action: "execution.generate_summary"
                    summary_includes: [
                        "total_execution_time", "steps_completed", "services_used", "final_outputs"
                    ]
                },
                {
                    name: "store_execution_history"
                    description: "Save execution details for future reference and reuse"
                    action: "execution.store_history"
                }
            ]
            output: {
                state: "workflow_execution"
                fields: ["status", "result", "execution_end_time"]
            }
        }
    ]
    
    // LAYER 4: TESTS (Validation)
    tests: [
        {
            id: "workflow_execution_test"
            description: "Test complete workflow execution with MCP service calls"
            input: {
                cue_workflow: """
                package workflow
                
                workflow: #DeterministicWorkflow & {
                    name: "Test Email Workflow"
                    steps: [
                        {
                            id: "send_message"
                            name: "Send Test Email"
                            service: "gmail"
                            action: "send_message"
                            inputs: {
                                to: "${user.recipient}"
                                subject: "Test Subject"
                                body: "Test message body"
                            }
                            outputs: { message_id: "sent_message_id" }
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
                    service_bindings: {
                        gmail: {
                            service: "gmail"
                            auth: {
                                oauth2: {
                                    scopes: ["https://www.googleapis.com/auth/gmail.send"]
                                }
                            }
                        }
                    }
                }
                """
                user_parameters: {
                    recipient: "test@example.com"
                }
            }
            expected: {
                status: "completed"
                completed_steps: ["send_message"]
                step_outputs: {
                    send_message: {
                        message_id: "msg_12345"
                    }
                }
                result: {
                    success: true
                    final_outputs: {
                        email_sent: true
                        message_id: "msg_12345"
                    }
                }
            }
        }
    ]
}
