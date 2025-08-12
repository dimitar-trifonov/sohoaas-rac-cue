# SOHOAAS User Workflow

This diagram shows the complete user journey through the SOHOAAS system, from authentication to workflow execution.

```mermaid
flowchart TD
    %% Authentication & Setup
    A[User Visits SOHOAAS] --> B[OAuth2 Login with Google]
    B --> C[Personal Capability Discovery]
    C --> D[Show Available Automations]
    
    %% Two Main Paths
    D --> E{User Choice}
    E -->|Execute Automation| F[Direct Intent Input]
    E -->|Discover Opportunities| G[Coach Me - Tell Your Story]
    
    %% Coaching Path
    G --> H[User Tells Daily Story]
    H --> I[LLM Analyzes Story]
    I --> J[Show Automation Suggestions]
    J --> K{User Interested?}
    K -->|Yes| F
    K -->|No| D
    
    %% Intent Processing
    F --> L[LLM Intent Analysis]
    L --> M[Extract Parameters & MCP Bindings]
    M --> N[Generate RaC Workflow]
    N --> O[Workflow Format Validation]
    
    %% Validation Results
    O --> P{Validation Result}
    P -->|Issues Found| Q[Show Missing/Invalid Items]
    Q --> R[Fix Issues]
    R --> O
    P -->|Valid| S[Show Generated Workflow]
    
    %% User Approval
    S --> T[Demo Workflow with Test Data]
    T --> U{User Approval}
    U -->|Needs Changes| V[Edit Workflow Steps]
    V --> S
    U -->|Rejected| D
    U -->|Approved| W[Workflow Ready for Execution]
    
    %% Execution Phase
    W --> X[User Edits Parameters for This Run]
    X --> Y[Execute RaC Workflow Deterministically]
    Y --> Z[Show Execution Results]
    Z --> AA{Run Again?}
    AA -->|Yes, Same Workflow| X
    AA -->|Yes, New Automation| D
    AA -->|No| BB[End Session]
    
    %% Styling
    classDef authFlow fill:#e1f5fe
    classDef coachFlow fill:#f3e5f5
    classDef intentFlow fill:#e8f5e8
    classDef validationFlow fill:#fff3e0
    classDef executionFlow fill:#fce4ec
    
    class A,B,C,D authFlow
    class G,H,I,J,K coachFlow
    class F,L,M,N intentFlow
    class O,P,Q,R,S,T,U,V validationFlow
    class W,X,Y,Z,AA,BB executionFlow
```

## Key Workflow Phases

### 🔐 **Authentication & Discovery**
- OAuth2 login with Google services
- Discover user's personal automation capabilities
- Show available automations based on connected services

### 🎓 **Coaching Mode (Optional)**
- User tells their daily story naturally
- LLM identifies automation opportunities
- Suggests relevant workflows to create

### 🧠 **Intent Processing**
- User expresses automation intent
- LLM analyzes and extracts parameters
- Generates executable RaC workflow

### ✅ **Validation & Approval**
- Validate workflow format and completeness
- Demo workflow with user's test data
- User approves or requests modifications

### ⚡ **Execution & Reuse**
- User edits parameters for each run
- Execute workflow deterministically
- Reuse same workflow with different parameters

## User Experience Highlights

- **Personal**: Uses user's own Google accounts and data
- **Guided**: Capability discovery shows what's possible
- **Natural**: Story-based coaching for opportunity discovery
- **Transparent**: User sees and approves generated workflows
- **Reusable**: Same workflow can be run multiple times with different parameters
- **Deterministic**: Reliable execution with clear results
