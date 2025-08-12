package rac

// =============================================
// 🔹 SOHOAAS PHASE 3 DEPLOYMENT BINDINGS
// =============================================
// Updated: In-process deployment, Golang + Genkit, OAuth2 authentication

// LAYER 3: IMPLEMENTATION (With what technologies)
SOHOAASDeployments: [...#Binding] & [
    // NOTE: Frontend deployment deferred for PoC - focus on backend pipeline
    {
        id: "sohoaas_backend_binding"
        type: "deployment"
        tech: "Go"
        mappings: [
            {
                source: "agent_manager"
                target: "Go Gin Server with Agent Manager"
                strategy: "in_process_deployment"
            },
            {
                source: "personal_capabilities_agent"
                target: "In-process Agent Component"
                strategy: "function_integration"
            },
            {
                source: "intent_gatherer_agent"
                target: "Genkit LLM Flow"
                strategy: "genkit_integration"
            },
            {
                source: "intent_analyst_agent"
                target: "Genkit LLM Flow"
                strategy: "genkit_integration"
            },
            {
                source: "workflow_generator_agent"
                target: "Genkit LLM Flow"
                strategy: "genkit_integration"
            }
        ]
        deployment: {
            technology: "Go"
            framework: "Gin + Google Genkit"
            port: 8081
            environment: {
                containerization: "Docker"
                orchestration: "Single binary deployment"
                monitoring: "Genkit reflection server on port 3101"
                logging: "Structured JSON logging with go-kit"
            }
            resources: {
                cpu: "2.0 cores"
                memory: "4Gi"
                storage: "Stateless"
            }
            config: {
                "GOOGLE_API_KEY": "${GOOGLE_API_KEY}"
                "MCP_SERVER_URL": "http://localhost:8080"
                "GENKIT_REFLECTION_PORT": "3101"
                "BACKEND_PORT": "8081"
                "LOG_LEVEL": "info"
                "OAUTH_REDIRECT_URL": "http://localhost:3002/api/auth/callback"
            }
        }
        metadata: {
            layer: "implementation"
            tags: ["backend", "api", "go", "genkit", "4_agents", "in_process", "oauth2"]
        }
    },
    {
        id: "mcp_service_binding"
        type: "deployment"
        tech: "Go"
        mappings: [{
            source: "oauth2_authentication"
            target: "MCP OAuth2 Service"
            strategy: "external_service"
        }]
        deployment: {
            technology: "Go"
            framework: "Gin + OAuth2"
            port: 8080
            environment: {
                containerization: "Docker"
                monitoring: "Health checks"
                logging: "OAuth2 audit logging"
            }
            resources: {
                cpu: "0.5 cores"
                memory: "512MB"
            }
            config: {
                "GOOGLE_CLIENT_ID": "${GOOGLE_CLIENT_ID}"
                "GOOGLE_CLIENT_SECRET": "${GOOGLE_CLIENT_SECRET}"
                "OAUTH_REDIRECT_URL": "http://localhost:3002/api/auth/callback"
                "MCP_PORT": "8080"
            }
        }
        metadata: {
            layer: "implementation"
            tags: ["oauth2", "mcp", "google_workspace", "authentication"]
        }
    }
]
