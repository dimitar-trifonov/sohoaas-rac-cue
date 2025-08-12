package rac

import "list"

// =============================================
// 🔹 MAIN SOHOAAS SYSTEM DEFINITION
// =============================================

// Complete SOHOAAS system combining all layers
// Note: SOHOAASSystem, SOHOAASArchitecture, SOHOAASDeployments are defined in other files
// but available here because they share the same package
CompleteSOHOAASSystem: RacSystem & {
    version: "2.0.0"
    
    // Layer 1: Requirements (from system.cue)
    states: SOHOAASSystem.states
    events: SOHOAASSystem.events  
    logic: SOHOAASSystem.logic
    ui: []
    tests: []
    
    // Layer 2 & 3: Architecture and Implementation (from architecture.cue + deployment.cue)
    bindings: list.Concat([[SOHOAASArchitecture], SOHOAASDeployments])
}

// Export for validation and tooling
sohoaas: CompleteSOHOAASSystem
