# SOHOAAS Phase 3 RaC Structure Updates

## 📋 **Update Summary**
Successfully updated the RaC structure to align with the current Phase 3 implementation based on architectural decisions made during development.

## 🔄 **Key Changes Made**

### **1. System Specification Updates** (`system.cue`)
- **Version**: Updated from 2.0.0 → 3.0.0
- **Architecture**: Reflected in-process deployment with 4-agent architecture
- **States Updated**:
  - `personal_chat` → `workflow_discovery` (multi-turn conversation)
  - `workflow_intent` → `intent_analysis` (5 PoC parameters)
  - `generated_workflow` → `deterministic_workflow` (CUE-based execution)
- **Events Updated**:
  - Removed story coaching events (deferred for PoC)
  - Updated to 4-agent pipeline flow
  - Added deterministic workflow generation events

### **2. Architecture Binding Updates** (`bindings.cue`)
- **New Architecture Pattern**: `monolith_with_agents`
- **Component Definitions**:
  - Agent Manager (orchestrator, in-process)
  - Personal Capabilities Agent (service proxy, static mapping)
  - Intent Gatherer Agent (LLM agent, multi-turn discovery)
  - Intent Analyst Agent (LLM agent, 5 PoC parameters)
  - Workflow Generator Agent (LLM agent, deterministic CUE generation)
  - Genkit Service (in-process LLM integration)
  - MCP Service (external OAuth2 service)
- **Communication Patterns**: In-process function calls + HTTP for MCP
- **Quality Attributes**: Optimized for PoC development speed

### **3. Deployment Configuration Updates** (`deployment.cue`)
- **Backend Binding**: Updated to reflect Golang + Genkit + 4 agents
- **Port Configuration**: Backend 8081, MCP 8080, Genkit reflection 3101
- **Resource Allocation**: 2.0 cores, 4Gi memory for shared agent processing
- **Environment Variables**: Updated for Google API integration
- **MCP Service Binding**: Added separate OAuth2 service deployment

### **4. Individual Agent Updates**

#### **Agent Manager** (`agent_manager.cue`)
- Updated event routing for 4-agent pipeline
- Removed story coaching routes
- Added service catalog management responsibilities
- Updated state transitions for new workflow states

#### **Personal Capabilities Agent** (`personal_capabilities.cue`)
- Changed from dynamic MCP discovery to static service catalog
- Updated to Google Workspace service focus
- Simplified capability mapping approach

#### **Intent Analyst Agent** (`intent_analyst.cue`)
- **Simplified Parameters**: Reduced from 9 complex parameters to 5 PoC parameters:
  1. `is_automation_request` (boolean)
  2. `required_services` (Google Workspace only)
  3. `can_fulfill` (boolean)
  4. `missing_info` (array)
  5. `next_action` (enum)
- Updated state from `workflow_intent` to `intent_analysis`
- Added service validation against catalog

#### **Intent Gatherer Agent** (`intent_gatherer.cue`)
- **Multi-turn Discovery**: Updated for conversation-based workflow pattern discovery
- **Discovery Phases**: Pattern → Trigger → Action → Data → Validation
- Removed single-message processing approach
- Added conversation history management

#### **Workflow Generator Agent** (`workflow_generator.cue`)
- **Deterministic Output**: Updated to generate CUE workflows with steps-based execution
- **New Structure**: Sequential steps with dependencies, user parameters, service bindings
- **Complete CUE Files**: Single executable specification output
- Updated from `workflow_generated` to `deterministic_workflow_generated` event

## 🎯 **Alignment Achieved**

### **Current Implementation Reflected**:
✅ **4-Agent Architecture**: Personal Capabilities + Intent Gatherer + Intent Analyst + Workflow Generator  
✅ **In-Process Deployment**: Monolithic backend with shared resources  
✅ **OAuth2 Authentication**: Personal Google Workspace integration  
✅ **Simplified Intent Analysis**: 5 PoC parameters instead of 9  
✅ **Multi-turn Workflow Discovery**: Conversation-based pattern identification  
✅ **Deterministic CUE Generation**: Steps-based executable workflows  
✅ **Service Catalog**: Centralized Google Workspace service management  
✅ **Genkit Integration**: LLM-powered agent implementation  

### **Removed/Deferred Elements**:
🔄 **Story Coaching Agent**: Removed from PoC pipeline (future enhancement)  
🔄 **Workflow Editor Agent**: Deferred for PoC simplicity  
🔄 **Frontend Deployment**: Focus on backend pipeline completion  
🔄 **Microservices**: Simplified to monolithic deployment  

## 📈 **Next Steps**
With RaC structure now aligned with implementation:
1. ✅ **RaC Updates Complete** - Structure matches Phase 3 implementation
2. 🎯 **Ready for OAuth2 Integration** - Next phase can proceed
3. 📋 **Documentation Synchronized** - Architecture decisions properly recorded

## 🔗 **Related Files Updated**
- `/rac/system.cue` - Core system specification
- `/rac/bindings.cue` - Architecture and deployment bindings  
- `/rac/deployment.cue` - Implementation deployment configuration
- `/rac/agents/agent_manager.cue` - Orchestrator updates
- `/rac/agents/personal_capabilities.cue` - Service catalog approach
- `/rac/agents/intent_analyst.cue` - 5 PoC parameters
- `/rac/agents/intent_gatherer.cue` - Multi-turn discovery
- `/rac/agents/workflow_generator.cue` - Deterministic CUE generation

The RaC structure now accurately reflects the current SOHOAAS Phase 3 implementation and is ready to support the next phase of OAuth2 Google Workspace integration.
