.PHONY: help init gcloud-login \
  mcp-build mcp-push mcp-deploy mcp-deploy-src \
  backend-build backend-build-local backend-push backend-build-push \
  oidc-proxy-build oidc-proxy-build-local oidc-proxy-push oidc-proxy-build-push \
  backend-deploy \
  frontend-build frontend-push frontend-deploy \
  services-urls images-list

# ===== Configurable variables (override via environment: make VAR=value target) =====
REGION ?= us-central1
PROJECT_ID ?= sohoaas-dev
REPO ?= mcp

# Tags/Images
MCP_IMAGE ?= service-proxies
MCP_TAG ?= poctag
FRONTEND_IMAGE ?= frontend
FRONTEND_TAG ?= v1
BACKEND_IMAGE ?= sohoaas-backend
BACKEND_TAG ?= poctag
OIDC_PROXY_IMAGE ?= oidc-proxy
OIDC_PROXY_TAG ?= poctag

# URLs (set after deployment)
BACKEND_URL ?= https://sohoaas-backend-$(shell gcloud projects describe $(PROJECT_ID) --format='value(projectNumber)').$(REGION).run.app
MCP_URL ?= https://mcp-backend-$(shell gcloud projects describe $(PROJECT_ID) --format='value(projectNumber)').$(REGION).run.app

# Paths
BACKEND_OIDC_YAML := deploy/backend-oidc-deployment.yaml
MCP_BACKEND_DIR := mcp/server/backend
FRONTEND_DIR := app/frontend
BACKEND_DIR := app/backend
OIDC_PROXY_DIR := app/oidc-proxy

help:
	@echo "Available targets:"
	@echo "  init                 - Enable APIs and create Artifact Registry repo (idempotent)"
	@echo "  gcloud-login         - Login and set project"
	@echo "  mcp-build            - Build MCP image with Cloud Build and push to Artifact Registry"
	@echo "  mcp-deploy           - Deploy MCP to Cloud Run from Artifact Registry image"
	@echo "  mcp-deploy-src       - Deploy MCP to Cloud Run from source (build by Cloud Run)"
	@echo "  backend-build        - Build backend image via Cloud Build (context: app/backend)"
	@echo "  backend-build-local  - Build backend image locally from repo root using app/backend/Dockerfile"
	@echo "  backend-push         - Push backend image to Artifact Registry"
	@echo "  backend-build-push   - Build (local) + push backend image"
	@echo "  oidc-proxy-build     - Build OIDC proxy image via Cloud Build (context: app/oidc-proxy)"
	@echo "  oidc-proxy-build-local - Build OIDC proxy image locally from app/oidc-proxy/"
	@echo "  oidc-proxy-push      - Push OIDC proxy image to Artifact Registry"
	@echo "  oidc-proxy-build-push - Build (local) + push OIDC proxy image"
	@echo "  backend-deploy       - Deploy backend with OIDC sidecar using $(BACKEND_OIDC_YAML)"
	@echo "  backend-build-push-stamped - Build+push backend with unique tag (STAMP)"
	@echo "  oidc-proxy-build-push-stamped - Build+push oidc-proxy with unique tag (STAMP)"
	@echo "  backend-deploy-stamped - Render YAML with STAMP tags and deploy (original YAML remains unchanged)"
	@echo "  frontend-build       - Docker build the frontend image"
	@echo "  frontend-push        - Push frontend image to Artifact Registry"
	@echo "  frontend-deploy      - Deploy frontend to Cloud Run with Nginx reverse proxy"
	@echo "  services-urls        - Print Cloud Run service URLs (frontend, backend, mcp)"

# ===== Common =====
PROJECT_NUMBER := $(shell gcloud projects describe $(PROJECT_ID) --format='value(projectNumber)' 2>/dev/null)
AR_REPO_PATH := $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPO)

# Unique tag (e.g., 20250909_130501-abc123). Requires git; falls back to timestamp only.
GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null)
STAMP := $(shell date +%Y%m%d_%H%M%S)$(if $(GIT_SHA),-$(GIT_SHA),)

init:
	gcloud services enable \
	  artifactregistry.googleapis.com \
	  cloudbuild.googleapis.com \
	  run.googleapis.com || true
	gcloud artifacts repositories create $(REPO) \
	  --repository-format=docker \
	  --location=$(REGION) \
	  --description="SOHOAAS containers" || true
	gcloud auth configure-docker $(REGION)-docker.pkg.dev -q

gcloud-login:
	gcloud auth login
	gcloud config set project $(PROJECT_ID)

# ===== MCP (Docs: docs/09.MCP_DEPLOYMENT.md) =====
mcp-build:
	cd $(MCP_BACKEND_DIR) && \
	gcloud builds submit --tag $(AR_REPO_PATH)/$(MCP_IMAGE):$(MCP_TAG) ./

mcp-deploy:
	gcloud run deploy mcp-backend \
	  --image $(AR_REPO_PATH)/$(MCP_IMAGE):$(MCP_TAG) \
	  --region $(REGION) \
	  --platform managed \
	  --allow-unauthenticated \
	  --port 8080 \
	  --set-env-vars LOG_LEVEL=info

# Optional: deploy from source without pre-building
mcp-deploy-src:
	cd $(MCP_BACKEND_DIR) && \
	gcloud run deploy mcp-backend \
	  --source . \
	  --region $(REGION) \
	  --platform managed \
	  --allow-unauthenticated \
	  --port 8080 \
	  --set-env-vars LOG_LEVEL=info

# ===== Backend with OIDC sidecar (Docs: docs/10.OIDC_PROXY_DEPLOYMENT.md) =====
backend-build:
	@test -d $(BACKEND_DIR) || (echo "Missing $(BACKEND_DIR)" && exit 1)
	cd $(BACKEND_DIR) && \
	gcloud builds submit --tag $(AR_REPO_PATH)/$(BACKEND_IMAGE):$(BACKEND_TAG) ./

# Build backend image locally from repo root (needed because Dockerfile copies ./rac and ./app/backend)
backend-build-local:
	docker build \
	  -t $(AR_REPO_PATH)/$(BACKEND_IMAGE):$(BACKEND_TAG) \
	  -f app/backend/Dockerfile .

backend-push:
	docker push $(AR_REPO_PATH)/$(BACKEND_IMAGE):$(BACKEND_TAG)

backend-build-push: backend-build-local backend-push

oidc-proxy-build:
	@test -d $(OIDC_PROXY_DIR) || (echo "Missing $(OIDC_PROXY_DIR)" && exit 1)
	cd $(OIDC_PROXY_DIR) && \
	gcloud builds submit --tag $(AR_REPO_PATH)/$(OIDC_PROXY_IMAGE):$(OIDC_PROXY_TAG) ./

# Local docker build for OIDC proxy (context is already app/oidc-proxy)
oidc-proxy-build-local:
	docker build -t $(AR_REPO_PATH)/$(OIDC_PROXY_IMAGE):$(OIDC_PROXY_TAG) app/oidc-proxy

oidc-proxy-push:
	docker push $(AR_REPO_PATH)/$(OIDC_PROXY_IMAGE):$(OIDC_PROXY_TAG)

oidc-proxy-build-push: oidc-proxy-build-local oidc-proxy-push

backend-deploy:
	@test -f $(BACKEND_OIDC_YAML) || (echo "Missing $(BACKEND_OIDC_YAML)" && exit 1)
	gcloud run services replace $(BACKEND_OIDC_YAML) --region=$(REGION) --platform=managed --project=$(PROJECT_ID)

# ===== Stamped build + deploy (uses original YAML content, rendered to a temp file) =====
backend-build-push-stamped:
	@echo "[STAMP] Using tag: $(STAMP)"
	docker build \
	  -t $(AR_REPO_PATH)/$(BACKEND_IMAGE):$(STAMP) \
	  -f app/backend/Dockerfile .
	docker push $(AR_REPO_PATH)/$(BACKEND_IMAGE):$(STAMP)
	@echo "[STAMP] Pushed backend: $(AR_REPO_PATH)/$(BACKEND_IMAGE):$(STAMP)"

oidc-proxy-build-push-stamped:
	@echo "[STAMP] Using tag: $(STAMP)"
	docker build -t $(AR_REPO_PATH)/$(OIDC_PROXY_IMAGE):$(STAMP) app/oidc-proxy
	docker push $(AR_REPO_PATH)/$(OIDC_PROXY_IMAGE):$(STAMP)
	@echo "[STAMP] Pushed oidc-proxy: $(AR_REPO_PATH)/$(OIDC_PROXY_IMAGE):$(STAMP)"

backend-deploy-stamped: backend-build-push-stamped oidc-proxy-build-push-stamped
	@echo "[STAMP] Rendering $(BACKEND_OIDC_YAML) to use tag $(STAMP)"
	@cp $(BACKEND_OIDC_YAML) /tmp/backend-oidc-deployment.$(STAMP).yaml
	@# Replace only the tags for sohoaas-backend:... and oidc-proxy:...
	@sed -i \
	  -e 's#\(image:\s\+.*sohoaas-backend:\)\S\+#\1$(STAMP)#' \
	  -e 's#\(image:\s\+.*oidc-proxy:\)\S\+#\1$(STAMP)#' \
	  /tmp/backend-oidc-deployment.$(STAMP).yaml
	@echo "[STAMP] Deploying temp YAML: /tmp/backend-oidc-deployment.$(STAMP).yaml"
	gcloud run services replace /tmp/backend-oidc-deployment.$(STAMP).yaml --region=$(REGION) --platform=managed --project=$(PROJECT_ID)

# ===== Frontend (Docs: docs/11.FRONTEND_DEPLOYMENT_CLOUD_RUN.md) =====
frontend-build:
	docker build -t $(AR_REPO_PATH)/$(FRONTEND_IMAGE):$(FRONTEND_TAG) $(FRONTEND_DIR)

frontend-push:
	docker push $(AR_REPO_PATH)/$(FRONTEND_IMAGE):$(FRONTEND_TAG)

frontend-deploy:
	@test -n "$(BACKEND_URL)" || (echo "Set BACKEND_URL to your backend Cloud Run URL" && exit 1)
	gcloud run deploy sohoaas-frontend \
	  --image=$(AR_REPO_PATH)/$(FRONTEND_IMAGE):$(FRONTEND_TAG) \
	  --region=$(REGION) \
	  --platform=managed \
	  --allow-unauthenticated \
	  --port=8080 \
	  --set-env-vars=NGINX_PORT=8080 \
	  --set-env-vars=BACKEND_SERVICE_URL=$(BACKEND_URL) \
	  --cpu=1 --memory=512Mi

services-urls:
	@echo "Frontend URL:" && gcloud run services describe sohoaas-frontend --region=$(REGION) --format='value(status.url)' || true
	@echo "Backend URL:" && gcloud run services describe sohoaas-backend --region=$(REGION) --format='value(status.url)' || true
	@echo "MCP URL:" && gcloud run services describe mcp-backend --region=$(REGION) --format='value(status.url)' || true

images-list:
	gcloud artifacts docker images list $(AR_REPO_PATH) --include-tags
