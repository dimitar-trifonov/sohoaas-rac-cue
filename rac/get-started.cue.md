## Here is the get-started.cue structure:

```cue
package rac

// =============================================
// 🔹 SCHEMA DEFINITIONS (Minimal)
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
    access?: {
        read?:  string
        write?: string
    }
    version?: string
    metadata?: {
        createdBy?: string
        tags?: [...string]
    }
}

#Event: {
    id:          string & !=""
    version:     string
    type:        "user" | "system"
    description: string
    events?: [...{
        id:          string
        type:        "user" | "system"
        description: string
        triggers?: [...{
            source: { type: "user" | "system" | "custom", name: string }
            formId?: string
            componentId?: string
            stateId?: string
            params?: { threshold?: number, delay?: number, errorCode?: string }
        }]
        logicChecks?: [...{
            type: "and" | "or"
            checks: [...{
                logic: string
                onError?: "halt" | "continue"
            }]
        }]
        actions?: [...{
            type: "validate" | "update" | "create" | "delete"
            state?: string
            fields?: [...string]
        }]
        metadata?: { createdBy?: string, tags?: [...string] }
    }]
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
    version?: string
    metadata?: {
        createdBy?: string
        tags?: [...string]
    }
}

#UI: {
    id:          string & !=""
    description: string
    bindsTo?:    string
    components?: [...{
        type:   string
        fields: [...{
            name:        string
            type:        string
            description: string
        }]
        submitEvent?: string
    }]
    views?: [...{
        path?:   string
        layout?: string
    }]
    logicPreview?: string
    version?: string
    metadata?: {
        createdBy?: string
        tags?: [...string]
    }
}

#Test: {
    id:    string & !=""
    setup?: {}
    testCases?: [...{
        id:          string & !=""
        description: string
        setup?:      {}
        input?: {
            event?: string
            data?:  {}
        }
        expected?:   {}
        expectError?: string
        notes?:      string
    }]
    version?: string
    metadata?: {
        createdBy?: string
        tags?: [...string]
    }
}

#Binding: {
    id:   string & !=""
    type: "technology" | "architecture" | "deployment" | "integration"
    
    // Technology bindings (traditional)
    tech?: string
    mappings?: [...{
        source:   string
        target:   string
        strategy: string
    }]
    
    // Architecture bindings (NEW)
    architecture?: {
        pattern: string  // e.g., "microservices", "monolith", "serverless"
        
        components?: [...{
            id:              string & !=""
            name:            string
            description?:    string
            responsibilities: [...string]
            interfaces:      [...string]
            
            // Capability requirements (abstract)
            capabilities?: [...string]
            
            // Dependencies on other components
            dependsOn?: [...string]
            
            // Communication patterns
            communicatesWith?: [...{
                component: string
                protocol:  string
                pattern:   "sync" | "async" | "event-driven"
            }]
        }]
        
        // System-wide patterns
        communicationPatterns?: [...{
            id:          string
            description: string
            path:        [...string]
            protocol:    string
            pattern:     "request-response" | "event-driven" | "streaming"
        }]
        
        // Quality attributes
        qualityAttributes?: {
            scalability?:   string
            availability?:  string
            performance?:   string
            security?:      string
            maintainability?: string
        }
        
        // Deployment topology
        topology?: {
            distribution: "single-node" | "multi-node" | "cloud-native" | "edge"
            networking?:  string
            storage?:     string
        }
    }
    
    // Deployment bindings (concrete implementation)
    deployment?: {
        technology?: string
        framework?:  string
        port?:       int
        environment?: {
            containerization?: string
            orchestration?:    string
            monitoring?:       string
            logging?:          string
        }
        
        // Resource requirements
        resources?: {
            cpu?:    string
            memory?: string
            storage?: string
        }
        
        // Configuration
        config?: {
            [key=string]: string | int | bool
        }
    }
    
    version?: string
    metadata?: {
        createdBy?: string
        tags?: [...string]
        layer?: "requirements" | "architecture" | "implementation"
    }
}

// =============================================
// 🔹 TOP-LEVEL STRUCTURE
// =============================================

RacSystem: {
    version?: string | *"1.0"
    states:   [...#State]
    events:   [...#Event]
    logic:    [...#Logic]
    ui:       [...#UI]
    tests:    [...#Test]
    bindings: [...#Binding]
}
```