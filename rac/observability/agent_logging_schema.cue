package observability

// SOHOAAS Agent Logging and Observability Schema
// Comprehensive input/output tracking and LLM interaction logging

AgentLoggingSchema: {
    version: "1.0.0"
    description: "Complete logging schema for all SOHOAAS agents with LLM interaction tracking"
    
    // Base logging structure for all agents
    BaseAgentLog: {
        agent_id: string
        agent_version: string
        execution_id: string
        timestamp: string
        user_id: string
        session_id: string
        
        // Input event logging
        input: {
            event_name: string
            event_data: {...}
            event_source: string
            event_timestamp: string
        }
        
        // LLM interaction logging
        llm_interaction: {
            provider: "openai"
            model: string
            temperature: number
            max_tokens: number
            
            // What we send to LLM
            prompt_sent: {
                system_prompt: string
                user_prompt: string
                context_data: {...}
                input_schema: {...}
                available_services?: {...}  // For agents that use MCP catalog
            }
            
            // What we get from LLM
            llm_response: {
                raw_response: string
                parsed_response: {...}
                token_usage: {
                    prompt_tokens: number
                    completion_tokens: number
                    total_tokens: number
                }
                response_time_ms: number
            }
            
            // LLM interaction metadata
            interaction_metadata: {
                prompt_hash: string
                response_hash: string
                success: bool
                error?: string
                retry_count: number
            }
        }
        
        // Agent processing steps
        processing_steps: [
            {
                step_name: string
                step_start: string
                step_end: string
                step_duration_ms: number
                step_input: {...}
                step_output: {...}
                step_status: "success" | "error" | "skipped"
                step_error?: string
            }
        ]
        
        // Output event logging
        output: {
            event_name: string
            event_data: {...}
            event_target: string
            event_timestamp: string
            validation_status: "valid" | "invalid" | "warning"
            validation_errors?: [string]
        }
        
        // Performance metrics
        performance: {
            total_execution_time_ms: number
            llm_interaction_time_ms: number
            processing_time_ms: number
            memory_usage_mb: number
        }
        
        // Error tracking
        errors: [
            {
                error_type: string
                error_message: string
                error_stack?: string
                error_timestamp: string
                error_context: {...}
            }
        ]
    }
}

// Specific logging schemas for each agent

PersonalCapabilitiesAgentLog: BaseAgentLog & {
    agent_id: "personal_capabilities"
    
    // Specific input structure
    input: {
        event_name: "user_authenticated"
        event_data: {
            user_id: string
            connected_services: [string]
            oauth_tokens: {...}
        }
    }
    
    // MCP discovery specific logging
    mcp_discovery: {
        mcp_endpoint: string
        discovery_steps: [
            {
                step: "query_providers"
                endpoint: string
                response_time_ms: number
                providers_found: [string]
                status: "success" | "error"
            },
            {
                step: "discover_schemas"
                services_queried: number
                schemas_discovered: number
                parallel_requests: number
                total_discovery_time_ms: number
                status: "success" | "error"
            }
        ]
        enhanced_catalog: {
            providers_count: number
            services_count: number
            functions_count: number
            parameters_discovered: number
            validation_rules_count: number
        }
    }
    
    // LLM interaction (minimal for this agent - mostly HTTP calls)
    llm_interaction: null  // This agent primarily does HTTP calls to MCP server
    
    // Output structure
    output: {
        event_name: "capabilities_discovered"
        event_data: {
            user_id: string
            available_services: {...}  // Complete MCP catalog with schemas
            available_actions: [string]
            examples: [string]
        }
    }
}

IntentGathererAgentLog: BaseAgentLog & {
    agent_id: "intent_gatherer"
    
    // Input structure
    input: {
        event_name: "user_message_received"
        event_data: {
            user_id: string
            message: string
            conversation_history: [...]
            user_context: {...}
        }
    }
    
    // Multi-turn conversation tracking
    conversation_flow: {
        turn_number: number
        discovery_phase: "pattern_discovery" | "trigger_identification" | "action_sequence" | "data_requirements" | "validation"
        conversation_state: {...}
        discovered_elements: {
            workflow_pattern?: string
            triggers?: [...]
            actions?: [...]
            data_requirements?: [...]
        }
    }
    
    // LLM interaction for workflow discovery
    llm_interaction: {
        provider: "openai"
        model: "gpt-4"
        temperature: 0.3
        
        prompt_sent: {
            system_prompt: string  // Intent gatherer system prompt
            user_prompt: string    // User's message
            context_data: {
                conversation_history: [...]
                discovery_phase: string
                previous_discoveries: {...}
            }
        }
        
        llm_response: {
            raw_response: string
            parsed_response: {
                next_question?: string
                discovered_pattern?: string
                workflow_intent?: {...}
                conversation_complete: bool
            }
        }
    }
    
    // Output structure
    output: {
        event_name: "workflow_intent_discovered"
        event_data: {
            user_id: string
            workflow_pattern: string
            validated_triggers: {...}
            workflow_parameters: {...}
            conversation_complete: bool
        }
    }
}

IntentAnalystAgentLog: BaseAgentLog & {
    agent_id: "intent_analyst"
    
    // Input structure
    input: {
        event_name: "workflow_intent_discovered"
        event_data: {
            user_id: string
            workflow_pattern: string
            validated_triggers: {...}
            workflow_parameters: {...}
        }
    }
    
    // LLM interaction for parameter extraction and validation
    llm_interaction: {
        provider: "openai"
        model: "gpt-4"
        temperature: 0.2
        
        prompt_sent: {
            system_prompt: string  // Intent analyst system prompt
            user_prompt: string    // Workflow pattern to analyze
            context_data: {
                workflow_pattern: string
                existing_parameters: {...}
            }
        }
        
        llm_response: {
            raw_response: string
            parsed_response: {
                validated_parameters: {...}
                missing_parameters: [string]
                parameter_validation_rules: {...}
                analysis_complete: bool
            }
        }
    }
    
    // Parameter analysis tracking
    parameter_analysis: {
        parameters_identified: number
        validation_rules_applied: number
        missing_parameters_count: number
        parameter_types_detected: [string]
    }
    
    // Output structure
    output: {
        event_name: "intent_analyzed"
        event_data: {
            user_id: string
            validated_workflow_pattern: string
            complete_parameters: {...}
            validation_rules: {...}
            analysis_status: "complete" | "needs_clarification"
        }
    }
}

WorkflowGeneratorAgentLog: BaseAgentLog & {
    agent_id: "workflow_generator"
    
    // Input structure (orchestrated by Agent Manager)
    input: {
        event_name: "workflow_generation_requested"
        event_data: {
            user_id: string
            workflow_pattern: string
            workflow_parameters: {...}
            available_services: {...}  // From Personal Capabilities Agent
        }
    }
    
    // LLM interaction for CUE file generation
    llm_interaction: {
        provider: "openai"
        model: "gpt-4"
        temperature: 0.1  // Low temperature for deterministic generation
        
        prompt_sent: {
            system_prompt: string  // Workflow generator system prompt
            user_prompt: string    // Workflow pattern
            context_data: {
                workflow_pattern: string
                workflow_parameters: {...}
                available_services: {...}  // Complete MCP catalog
            }
            input_schema: {...}
        }
        
        llm_response: {
            raw_response: string
            parsed_response: {
                workflow_cue: string      // Generated CUE file
                workflow_metadata: {...}
                validation_passed: bool
            }
        }
    }
    
    // CUE generation analysis
    cue_generation: {
        steps_generated: number
        services_used: [string]
        parameters_mapped: number
        dependencies_created: number
        validation_rules_applied: number
        mcp_compliance_check: {
            exact_service_paths: bool
            exact_function_names: bool
            parameter_schema_compliance: bool
        }
    }
    
    // Output structure
    output: {
        event_name: "workflow_generated"
        event_data: {
            user_id: string
            workflow_name: string
            workflow_cue: string
            workflow_description: string
            execution_steps: [...]
            required_parameters: {...}
            service_dependencies: [string]
            mcp_services_used: [string]
        }
    }
}

AgentManagerLog: BaseAgentLog & {
    agent_id: "agent_manager"
    
    // Event routing logging
    event_routing: {
        incoming_event: {
            event_name: string
            source_agent: string
            event_data: {...}
            routing_timestamp: string
        }
        
        routing_decision: {
            target_agents: [string]
            routing_rules_applied: [string]
            orchestration_required: bool
            data_aggregation_needed: bool
        }
        
        orchestration_steps: [
            {
                step: string
                agent_called: string
                data_sent: {...}
                response_received: {...}
                step_duration_ms: number
            }
        ]
        
        outgoing_events: [
            {
                event_name: string
                target_agent: string
                event_data: {...}
                dispatch_timestamp: string
            }
        ]
    }
    
    // No direct LLM interaction - pure orchestration
    llm_interaction: null
    
    // Aggregated data from multiple agents
    data_aggregation: {
        capabilities_data?: {...}  // From Personal Capabilities
        intent_data?: {...}        // From Intent Gatherer
        analysis_data?: {...}      // From Intent Analyst
        final_orchestrated_data: {...}
    }
}

// Logging configuration and output formats
LoggingConfig: {
    log_level: "DEBUG" | "INFO" | "WARN" | "ERROR"
    output_format: "JSON" | "STRUCTURED"
    
    destinations: {
        console: bool
        file: {
            enabled: bool
            path: string
            rotation: "daily" | "size"
        }
        cloud_logging: {
            enabled: bool
            project_id: string
            log_name: string
        }
        metrics: {
            enabled: bool
            prometheus_endpoint: string
        }
    }
    
    // What to log for each agent
    agent_logging_config: {
        personal_capabilities: {
            log_mcp_requests: bool
            log_schema_discovery: bool
            log_performance_metrics: bool
        }
        intent_gatherer: {
            log_conversation_flow: bool
            log_llm_interactions: bool
            log_discovery_progress: bool
        }
        intent_analyst: {
            log_parameter_analysis: bool
            log_llm_interactions: bool
            log_validation_rules: bool
        }
        workflow_generator: {
            log_cue_generation: bool
            log_llm_interactions: bool
            log_mcp_compliance: bool
        }
        agent_manager: {
            log_event_routing: bool
            log_orchestration_steps: bool
            log_data_aggregation: bool
        }
    }
}

// Sample log queries for debugging
LogQueries: {
    // Find all LLM interactions for a specific user session
    llm_interactions_by_session: """
    SELECT agent_id, llm_interaction.prompt_sent, llm_interaction.llm_response, timestamp
    FROM agent_logs 
    WHERE session_id = ? AND llm_interaction IS NOT NULL
    ORDER BY timestamp
    """
    
    // Track workflow generation pipeline for specific user
    workflow_pipeline_trace: """
    SELECT agent_id, input.event_name, output.event_name, performance.total_execution_time_ms
    FROM agent_logs 
    WHERE user_id = ? AND execution_id = ?
    ORDER BY timestamp
    """
    
    // Find performance bottlenecks
    performance_analysis: """
    SELECT agent_id, AVG(performance.total_execution_time_ms) as avg_time, 
           AVG(performance.llm_interaction_time_ms) as avg_llm_time
    FROM agent_logs 
    WHERE timestamp > ? 
    GROUP BY agent_id
    """
    
    // Error tracking across agents
    error_analysis: """
    SELECT agent_id, COUNT(*) as error_count, errors.error_type
    FROM agent_logs 
    WHERE errors IS NOT NULL AND timestamp > ?
    GROUP BY agent_id, errors.error_type
    """
}
