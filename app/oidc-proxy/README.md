# OIDC Proxy (sidecar)

Tiny Go reverse proxy that injects a Google OIDC identity token on each outbound request to an upstream service (e.g., MCP on Cloud Run). Used as a sidecar so callers don’t need to implement token minting.

## Env contract
- `LISTEN_ADDR` (default `:8070`) — local listen address
- `UPSTREAM_URL` (required) — full upstream base URL, e.g. `https://mcp-xxxx-uc.a.run.app`
- `AUDIENCE` (default = `UPSTREAM_URL`) — OIDC audience value
- `FORWARD_HEADERS` (default `Authorization,X-Firebase-Authorization,Content-Type`) — comma-separated headers to copy from client to upstream
- `TIMEOUT_SECONDS` (default `30`)
- `CORS_ALLOW_ORIGIN` (optional) — set for local testing (e.g., `*`)
- `LOG_REQUESTS` (optional) — any non-empty value enables basic request dump
- `LOG_STARTUP_DUMP` (default `true`) — startup log line

## Health/Debug
- `GET /debug` → returns upstream/audience

## Local dev
- Upstream unauthenticated: point backend directly to MCP and skip sidecar.
- Upstream authenticated: run proxy locally and set `GOOGLE_APPLICATION_CREDENTIALS` to a service account JSON with permission to invoke the upstream.

## Cloud Run (multi-container)
Example using container names `backend` and `oidc-proxy` in one service (YAML):

```yaml
apiVersion: run.googleapis.com/v1
kind: Service
metadata:
  name: sohoaas-backend
spec:
  template:
    spec:
      containers:
      - name: backend
        image: <YOUR_BACKEND_IMAGE>
        env:
        - name: MCP_SERVICE_URL
          value: http://localhost:8070    # talk to sidecar
        - name: ENVIRONMENT
          value: production
        ports:
        - containerPort: 8080
      - name: oidc-proxy
        image: <YOUR_OIDC_PROXY_IMAGE>
        env:
        - name: LISTEN_ADDR
          value: ":8070"
        - name: UPSTREAM_URL
          value: https://<mcp-run-url>
        - name: AUDIENCE
          value: https://<mcp-run-url>
        ports:
        - containerPort: 8070
```

Grant Cloud Run Invoker on MCP only to the backend’s service account. The sidecar will mint tokens via metadata server automatically in Cloud Run.

## Build
```
# from app/oidc-proxy
docker build -t oidc-proxy:local .
```
