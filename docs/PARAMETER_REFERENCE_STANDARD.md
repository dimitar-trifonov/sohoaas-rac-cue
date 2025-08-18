# Parameter Reference Standardization

## Overview
This document defines the unified parameter reference format used across all SOHOAAS workflows, schemas, and implementations.

## Standardized Format

### User Parameters
**Format:** `${user.parameter_name}`
- **Purpose:** Reference user-provided input parameters
- **Examples:**
  - `${user.recipient_email}`
  - `${user.document_title}`
  - `${user.folder_name}`

### Step Outputs
**Format:** `${steps.step_id.outputs.output_name}`
- **Purpose:** Reference outputs from previous workflow steps
- **Examples:**
  - `${steps.create_document.outputs.document_id}`
  - `${steps.share_document.outputs.share_url}`
  - `${steps.create_folder.outputs.folder_id}`

### Computed Values
**Format:** `${computed.expression}`
- **Purpose:** Reference dynamically computed values
- **Examples:**
  - `${computed.current_date}`
  - `${computed.timestamp}`

### Environment Variables
**Format:** `${ENV_VAR_NAME}`
- **Purpose:** Reference deployment environment variables
- **Examples:**
  - `${OPENAI_API_KEY}`
  - `${FIREBASE_PROJECT_ID}`

## Deprecated Formats (DO NOT USE)
- `${USER_INPUT:param}` ❌ - Legacy format, replaced by `${user.param}`
- `$(step.outputs.field)` ❌ - Inconsistent syntax
- `{user.param}` ❌ - Missing dollar sign prefix

## Implementation Requirements

### JSON Schema Files
```json
{
  "parameters": {
    "to": "${user.recipient_email}",
    "file_id": "${steps.create_document.outputs.document_id}"
  }
}
```

### CUE Schema Files
```cue
parameters: {
    title: "${user.document_title}"
    file_id: "${steps.create_document.outputs.document_id}"
}
```

### Go Backend Code
```go
// Parameter resolution
if strings.HasPrefix(value, "${user.") && strings.HasSuffix(value, "}") {
    paramName := strings.TrimSuffix(strings.TrimPrefix(value, "${user."), "}")
    //## Next Critical Gaps

With parameter standardization complete, the remaining implementation blockers are:

1. **Missing Validation Actions** - Implement referenced validation services in cue_generator
2. **Service Binding Structure** - Align JSON vs CUE service binding formats

*Note: Execution Order Logic skipped for PoC - sequential execution only for minimalistic approach*
}
```

## Validation Rules

1. **Consistency:** All parameter references must use the standardized format
2. **Completeness:** Referenced parameters must be defined in user_parameters section
3. **Dependencies:** Step output references must respect dependency order
4. **Type Safety:** Parameter types must match expected MCP service input types

## Migration Guide

### From Legacy Format
```diff
- "${USER_INPUT:recipient}"
+ "${user.recipient}"

- "$(step.create_doc.id)"
+ "${steps.create_doc.outputs.document_id}"
```

### Validation Commands
```bash
# Check for deprecated formats
grep -r "\${USER_INPUT:" rac/ app/
grep -r "\$(" rac/ app/

# Validate standardized format
grep -r "\${user\." rac/ app/
grep -r "\${steps\." rac/ app/
```

## Benefits

1. **Consistency:** Single format across all components
2. **Readability:** Clear distinction between user params and step outputs
3. **Maintainability:** Easier to parse and validate
4. **Tool Support:** Better IDE and validation tool support
5. **Documentation:** Self-documenting parameter sources

## Files Updated
- `rac/schemas/deterministic_workflow.cue`
- `rac/schemas/workflow_json_schema.json`
- `rac/agents/workflow_validator.cue`
- `rac/services/workflow_executor.cue`
- `rac/services/cue_generator.cue`
- `app/backend/internal/services/execution_engine.go`
- `app/backend/internal/services/cue_builder_test.go`
- `app/backend/internal/services/workflow_generator_test.go`
- `app/backend/prompts/workflow_generator.prompt`

This standardization ensures the workflow_generator → cue_generator → execution_engine pipeline functions correctly with consistent parameter resolution.
