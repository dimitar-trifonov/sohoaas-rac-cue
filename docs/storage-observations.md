# SOHOAAS Storage Observations and Recommendations (PoC)

This note summarizes the current storage design and practical recommendations for running locally and on Google Cloud Run with Google Cloud Storage (GCS).

## Findings

- **Unified interface**: Storage is abstracted by `storage.WorkflowStorage` (`app/backend/internal/storage/interfaces.go`). API handlers depend only on this interface.
- **Local backend**: Implemented in `app/backend/internal/storage/local_storage.go`. Stores workflows under `<workflows_dir>/<userID>/<workflowID>/workflow.cue` and parses CUE content when saving/listing.
- **GCS backend**: Implemented in `app/backend/internal/storage/gcs_storage.go`. Stores workflows under `gs://<bucket>/<prefix>/<userID>/<workflowID>/workflow.cue` and lists by prefix.
- **Factory**: `storage.StorageFactory.NewStorage()` selects backend from `StorageConfig` (`app/backend/internal/storage/interfaces.go`).
- **API usage**: Endpoints call only the interface methods:
  - `GetUserWorkflows()` → `workflowStorage.ListUserWorkflows(userID)` in `app/backend/internal/api/handlers.go`.
  - `GetWorkflow()` → `workflowStorage.GetWorkflow(userID, workflowID)`.
  - `ExecuteWorkflow()` → loads workflow, then passes `workflow.Content` to `ExecutionEngine`.
- **Stateful execution by ID**: Execution uses `workflow_id`; CUE content is loaded server-side. See `app/backend/internal/api/handlers.go` `ExecuteWorkflow()`.
- **ID format**: Backends use `ID = userID_workflowID`. Local and GCS handle trimming `userID_` internally when reading.
- **Legacy service (not used by handlers)**: `app/backend/internal/services/workflow_storage.go` contains a legacy filesystem storage helper; current API uses the `storage` package backends.

## Compatibility Assessment (Local + GCS)

- **PoC-ready**: Both backends are compatible with current handlers and the execution engine. Using Local for development/tests and GCS for Cloud Run is supported.
- **Parsed data parity**:
  - Local: parses CUE to `ParsedData` on save/list.
  - GCS: returns raw content (no parse). Execution does not require `ParsedData`, but UI features might. Consider parity (see recommendations).

## Recommendations

- **Configuration**
  - Use `StorageFactory.NewStorage()` with `StorageConfig.Backend = "gcs"` on Cloud Run and `"local"` for local dev.
  - For GCS, set `BucketName` and `WorkflowsPrefix` (ensure it ends with `/`). Prefer application default credentials (Workload Identity) in Cloud Run; avoid embedding `ServiceAccountKey`.

- **ID handling**
  - Keep the `userID_workflowID` convention end-to-end. Ensure the frontend always uses the exact ID returned by the backend (`workflow_file.id`) when executing.

- **Parity improvement (optional)**
  - Add CUE parsing in `gcs_storage.go` after `GetWorkflow()` and in `ListUserWorkflows()` similar to Local by calling the shared parser (`app/backend/internal/storage/workflow_parser.go`) to populate `ParsedData`. This helps the UI render consistent metadata.

- **Operational tips**
  - For Cloud Run: grant the service account `roles/storage.objectAdmin` on the bucket; set `GOOGLE_CLOUD_PROJECT`. No local writes in production (ephemeral FS).
  - For large tenancy, consider listing pagination in GCS (PoC can skip).
  - Keep artifact writing (`SavePrompt`, `SaveResponse`, `SaveExecutionLog`) on the same backend to centralize execution records.

## Quick Checks Before Deploy

- **[config]** Verify `StorageConfig` is loaded once during service init and wired into `Handler`.
- **[gcs]** Confirm bucket exists and service account has access; `workflows_prefix` normalized with trailing `/`.
- **[id]** Execute a flow that generates a workflow and then call `ExecuteWorkflow` with the returned `workflow_id`.
- **[token]** Confirm OAuth token storage/retrieval (`TokenManager`) works in Cloud Run for MCP calls.
- **[engine]** Run a simple 2-step workflow to verify `${user.*}` substitution and `${steps.*}` chaining across storage backends.

## Relevant Files

- `app/backend/internal/storage/interfaces.go`
- `app/backend/internal/storage/local_storage.go`
- `app/backend/internal/storage/gcs_storage.go`
- `app/backend/internal/storage/workflow_parser.go`
- `app/backend/internal/api/handlers.go`
- `app/backend/internal/services/execution_engine.go`

## Status

- Current implementation is PoC-ready for Local + GCS. No blockers identified. Apply parity parsing for GCS if the UI needs structured workflow metadata.
