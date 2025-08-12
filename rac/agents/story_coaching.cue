package agents

import "../schemas.cue"

// =============================================
// 🔹 STORY COACHING AGENT
// =============================================

StoryCoachingAgent: {
    version: "1.0.0"
    type: "autonomous_agent"
    poc_status: "future_enhancement" // Skip for PoC - valuable for non-technical users
    
    // LAYER 1: REQUIREMENTS (What)
    states: [
        {
            id: "daily_story_coaching"
            type: "object"
            fields: [
                { name: "session_id", type: "string", required: true },
                { name: "user_id", type: "string", required: true },
                { name: "user_story", type: "string" },
                { name: "story_highlights", type: "array" },
                { name: "automation_suggestions", type: "array" },
                { name: "conversation_flow", type: "array" },
                { name: "status", type: "string" } // "listening", "analyzing", "suggesting", "completed"
            ]
            metadata: {
                tags: ["coaching", "story", "daily_routine", "natural"]
            }
        }
    ]
    
    events: [
        {
            id: "start_story_coaching"
            version: "1.0"
            type: "input"
            description: "Begin daily story coaching session"
            triggers: ["initiate_coaching"]
        },
        {
            id: "story_received"
            version: "1.0"
            type: "input"
            description: "User shares their daily story"
            triggers: ["analyze_story"]
        },
        {
            id: "story_analyzed"
            version: "1.0"
            type: "output"
            description: "Story analysis complete with automation suggestions"
            data_schema: {
                story_highlights: ["string"]
                automation_opportunities: ["string"]
                suggested_workflows: ["object"]
            }
        }
    ]
    
    logic: [
        {
            id: "initiate_coaching"
            type: "genkit_flow"
            description: "Start natural coaching conversation"
            input_schema: {
                user_id: "string"
                user_capabilities: ["string"]
            }
            steps: [
                {
                    name: "generate_opening"
                    description: "Create personalized coaching opening"
                    action: "llm.generate_coaching_prompt"
                    llm_config: {
                        model: "gpt-4"
                        temperature: 0.8
                        system_prompt: "You are a friendly automation coach helping users discover automation opportunities in their daily routine. Be conversational and encouraging."
                    }
                }
            ]
        },
        {
            id: "analyze_story"
            type: "genkit_flow"
            description: "Analyze user's daily story for automation opportunities"
            input_schema: {
                user_story: "string"
                user_capabilities: ["string"]
            }
            steps: [
                {
                    name: "extract_routine_patterns"
                    description: "Identify repetitive tasks and pain points"
                    action: "llm.extract_patterns"
                    llm_config: {
                        model: "gpt-4"
                        temperature: 0.3
                        system_prompt: "Analyze daily routine story and identify repetitive tasks that could be automated"
                    }
                },
                {
                    name: "match_automation_opportunities"
                    description: "Match patterns to available capabilities"
                    action: "capability.match_opportunities"
                },
                {
                    name: "generate_suggestions"
                    description: "Create specific automation suggestions"
                    action: "llm.generate_suggestions"
                    llm_config: {
                        model: "gpt-4"
                        temperature: 0.7
                        system_prompt: "Generate specific, actionable automation suggestions based on user's story and available capabilities"
                    }
                }
            ]
            output_event: "story_analyzed"
        }
    ]
    
    // LAYER 2: ARCHITECTURE (How)
    architecture: {
        type: "microservice"
        components: {
            conversation_manager: {
                type: "llm_agent"
                responsibilities: ["natural_conversation", "coaching_guidance"]
            }
            pattern_analyzer: {
                type: "llm_agent"
                responsibilities: ["routine_analysis", "opportunity_detection"]
            }
            suggestion_engine: {
                type: "rule_engine"
                responsibilities: ["capability_matching", "suggestion_ranking"]
            }
        }
        communication: {
            input: ["event_bus"]
            output: ["event_bus"]
            external: ["openai_api"]
        }
        quality_attributes: {
            usability: "high" // Natural conversation experience
            personalization: "high"
            engagement: "high"
        }
    }
    
    // LAYER 3: IMPLEMENTATION (With What)
    bindings: [
        {
            type: "deployment"
            technology: "golang"
            framework: "genkit"
            deployment: {
                service_name: "story-coaching-agent"
                port: 8083
                resources: {
                    cpu: "0.8"
                    memory: "1Gi"
                }
                config: {
                    openai_api_key: "${OPENAI_API_KEY}"
                    coaching_style: "friendly"
                    max_story_length: "2000"
                }
            }
        }
    ]
    
    tests: [
        {
            id: "monday_morning_routine_analysis"
            type: "integration"
            description: "Analyze typical Monday morning routine story"
            input: {
                user_story: "Every Monday morning I get to the office, grab my coffee, and spend 2 hours going through weekend emails. I pull out project updates and try to remember what everyone was working on. Then I write a status report for my boss."
                user_capabilities: ["send_email", "create_document", "read_emails"]
            }
            expected: {
                story_highlights: ["Monday morning email review", "Status report creation"]
                automation_suggestions: [
                    "I can automatically scan your emails and create your Monday status report",
                    "I can organize your weekend emails by project"
                ]
            }
        }
    ]
    
    ui: [
        {
            id: "coaching_chat"
            type: "component"
            description: "Natural conversation interface for story coaching"
            props: {
                coaching_session: "daily_story_coaching"
                suggestions: "array"
            }
            events: ["story_shared", "suggestion_selected"]
        }
    ]
}
