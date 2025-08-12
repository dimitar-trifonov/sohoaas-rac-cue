package rac

// =============================================
// 🔹 CUE IMPORT EXAMPLES
// =============================================

// 1. BUILT-IN PACKAGE IMPORTS
import (
    "list"      // List manipulation functions
    "strings"   // String manipulation functions
)

// 2. SAME PACKAGE DEFINITIONS (automatically available)
// All definitions in files with "package rac" are automatically merged:
// - #State, #Event, #Logic, #UI, #Test, #Binding (from schemas.cue)
// - RacSystem (from system.cue)  
// - SOHOAASArchitecture (from architecture.cue)
// - SOHOAASDeployments (from deployment.cue)

// =============================================
// 🔹 EXAMPLES USING IMPORTS
// =============================================

// Example using list functions
ExampleBindingsList: list.Concat([
    [SOHOAASArchitecture],
    SOHOAASDeployments
])

// Example using string functions
ExampleSystemName: strings.ToUpper("sohoaas")

// Example using time functions (CUE doesn't have time.Now, using static example)
ExampleTimestamp: "2025-08-07T15:03:57+03:00"

// Example using JSON encoding
ExampleConfig: {
    raw: {
        name: "SOHOAAS"
        version: "2.0.0"
        components: ["intent_capture", "intent_processor", "workflow_engine"]
    }
    // JSON encoding example (simplified for demo)
    encoded: "{\"name\":\"SOHOAAS\",\"version\":\"2.0.0\"}"
}

// =============================================
// 🔹 MODULAR SYSTEM COMPOSITION
// =============================================

// Demonstrate how imports enable modular composition
ModularSOHOAAS: RacSystem & {
    version: "2.0.0"
    
    // Requirements layer (defined in system.cue)
    states: [{
        id: "user_intent"
        type: "object"
        fields: [
            { name: "description", type: "string", required: true },
            { name: "context", type: "object" },
            { name: "status", type: "string" }
        ]
    }]
    
    events: [{
        id: "intent_processing"
        version: "1.0"
        type: "system"
        description: "Process user intent into executable workflow"
    }]
    
    logic: [{
        id: "intent_validation"
        appliesTo: "user_intent"
        rules: [{
            if: "description != null && description != ''"
            then: [{ error: "Intent description is required" }]
        }]
    }]
    
    ui: []
    tests: []
    
    // Architecture and deployment layers (imported from other files)
    bindings: ExampleBindingsList
}

// =============================================
// 🔹 VALIDATION HELPERS
// =============================================

// Helper to validate that all required components are present
ValidationReport: {
    hasSchemas: #State != _|_ && #Event != _|_ && #Logic != _|_ && #UI != _|_ && #Test != _|_ && #Binding != _|_
    hasSystem: RacSystem != _|_
    hasArchitecture: SOHOAASArchitecture != _|_
    hasDeployments: SOHOAASDeployments != _|_
    
    summary: {
        if hasSchemas && hasSystem && hasArchitecture && hasDeployments {
            status: "✅ All components successfully imported"
        }
        if !hasSchemas {
            status: "❌ Missing schema definitions"
        }
        if !hasSystem {
            status: "❌ Missing system definition"
        }
        if !hasArchitecture {
            status: "❌ Missing architecture definition"
        }
        if !hasDeployments {
            status: "❌ Missing deployment definitions"
        }
    }
}
