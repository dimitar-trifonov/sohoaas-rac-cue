# Multi-stage build for oidc-proxy
FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/oidc-proxy ./...

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /
COPY --from=builder /out/oidc-proxy /usr/local/bin/oidc-proxy
USER nonroot:nonroot
EXPOSE 8070
ENTRYPOINT ["/usr/local/bin/oidc-proxy"]
