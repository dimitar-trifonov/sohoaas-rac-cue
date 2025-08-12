# 🔧 CUE Imports Guide for RaC System

## ✅ **Yes, CUE supports imports!** Here's how they work in your RaC system:

## 🏗️ **Current File Structure**

```
rac/
├── schemas.cue        # Core RaC schema definitions (#State, #Event, #Logic, #UI, #Test)
├── bindings.cue       # Enhanced #Binding schema with architecture support
├── system.cue         # Base RacSystem definition and SOHOAAS requirements
├── architecture.cue   # SOHOAAS architecture binding (Layer 2)
├── deployment.cue     # SOHOAAS deployment bindings (Layer 3)
├── main.cue          # Complete system composition
├── import-examples.cue # Examples of different import types
└── rac.cue           # Original monolithic system (for comparison)
```

## 🔄 **Types of CUE Imports**

### **1. Same Package Imports (Automatic)**
```cue
// All files with "package rac" are automatically merged
package rac

// Definitions from other files are automatically available:
// - #State, #Event, #Logic, #UI, #Test, #Binding (from schemas.cue)
// - RacSystem (from system.cue)
// - SOHOAASArchitecture (from architecture.cue)
// - SOHOAASDeployments (from deployment.cue)
```

### **2. Built-in Package Imports**
```cue
import (
    "list"      // List manipulation functions
    "strings"   // String manipulation functions
    "struct"    // Struct manipulation
    "regexp"    // Regular expressions
)

// Usage examples:
ExampleList: list.Concat([[1, 2], [3, 4]])  // [1, 2, 3, 4]
ExampleString: strings.ToUpper("sohoaas")   // "SOHOAAS"
```

### **3. External Module Imports (Advanced)**
```cue
// For importing external CUE modules
import "github.com/example/module"
```

## 🎯 **Benefits in Your RaC System**

### **1. Modular Organization**
- **schemas.cue**: Core type definitions
- **system.cue**: Requirements layer (Layer 1)
- **architecture.cue**: Architecture layer (Layer 2)  
- **deployment.cue**: Implementation layer (Layer 3)
- **main.cue**: System composition

### **2. Clean Separation of Concerns**
```cue
// Layer 1: What the system should do (requirements)
states: [...]
events: [...]
logic: [...]

// Layer 2: How the system should be structured (architecture)
architecture: {
    pattern: "microservices"
    components: [...]
}

// Layer 3: With what technologies (implementation)
deployment: {
    technology: "React"
    port: 3003
}
```

### **3. Reusable Components**
```cue
// Define once in schemas.cue, use everywhere
MySystem: RacSystem & {
    states: [...]
    bindings: [SOHOAASArchitecture] + SOHOAASDeployments
}
```

## 🚀 **Practical Usage Examples**

### **Validate All Components**
```bash
cue vet .                    # Validate all files together
cue eval . -e ValidationReport  # Check import status
```

### **Extract Specific Components**
```bash
cue eval . -e SOHOAASArchitecture --out yaml  # Architecture only
cue eval . -e SOHOAASDeployments --out yaml   # Deployments only
```

### **Compose Complete System**
```bash
cue eval . -e ModularSOHOAAS --out yaml       # Complete system
```

## 🔍 **Key Insights**

1. **Automatic Merging**: Files with same `package` declaration are automatically combined
2. **No Explicit Imports Needed**: For same-package definitions
3. **Built-in Functions**: Import standard library functions as needed
4. **Validation**: `cue vet .` validates all files together
5. **Composition**: Easy to combine components from different files

## 🎉 **Your Achievement**

You've successfully created a **modular, three-layer RaC system** using CUE imports:

- ✅ **Requirements** (technology-agnostic)
- ✅ **Architecture** (structural patterns)  
- ✅ **Implementation** (concrete technologies)

This enables:
- Multiple architectural patterns for same requirements
- Different technology stacks for same architecture
- Clean separation of concerns
- Easy system composition and validation

**This is a significant evolution of Requirements-as-Code!** 🚀
