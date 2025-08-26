package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"google.golang.org/api/idtoken"
)

// Env vars
// LISTEN_ADDR: e.g. ":8082" (default)
// UPSTREAM_URL: required, full URL to MCP service (e.g., https://mcp-...run.app)
// AUDIENCE: optional, defaults to UPSTREAM_URL
// FORWARD_HEADERS: optional comma-separated list of headers to forward from client (default: Authorization,X-Firebase-Authorization,Content-Type)
// TIMEOUT_SECONDS: optional HTTP client timeout (default: 30)
// CORS_ALLOW_ORIGIN: optional for local testing, e.g. "*" or specific origin

func main() {
	listenAddr := getEnv("LISTEN_ADDR", ":8070")
	upstream := getEnv("UPSTREAM_URL", "")
	if upstream == "" {
		log.Fatalf("UPSTREAM_URL is required")
	}
	audience := getEnv("AUDIENCE", upstream)
	forwardHeaders := parseCSV(getEnv("FORWARD_HEADERS", "Authorization,X-Firebase-Authorization,Content-Type"))
	clientTimeout := getEnvInt("TIMEOUT_SECONDS", 30)
	allowOrigin := os.Getenv("CORS_ALLOW_ORIGIN")

	targetURL, err := url.Parse(upstream)
	if err != nil {
		log.Fatalf("invalid UPSTREAM_URL: %v", err)
	}

	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	client := &http.Client{Transport: transport, Timeout: time.Duration(clientTimeout) * time.Second}

	// Prepare token source (works in Cloud Run and locally with GOOGLE_APPLICATION_CREDENTIALS)
	ts, err := idtoken.NewTokenSource(context.Background(), audience)
	if err != nil {
		log.Fatalf("failed to create idtoken source: %v", err)
	}

	reverseProxy := func(w http.ResponseWriter, r *http.Request) {
		if allowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Firebase-Authorization,Accept")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Build upstream request
		up := &http.Request{Method: r.Method}
		up = up.WithContext(r.Context())
		upURL := *targetURL
		upURL.Path = singleJoin(upURL.Path, r.URL.Path)
		upURL.RawQuery = r.URL.RawQuery
		up.URL = &upURL

		// Copy body
		if r.Body != nil {
			defer r.Body.Close()
			b, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("read body error: %v", err), http.StatusBadRequest)
				return
			}
			up.Body = io.NopCloser(strings.NewReader(string(b)))
			up.ContentLength = int64(len(b))
		}

		up.Header = make(http.Header)
		// Forward selected headers from client
		for _, h := range forwardHeaders {
			if v := r.Header.Get(h); v != "" {
				up.Header.Set(h, v)
			}
		}

		// Mint OIDC token for this audience and attach
		tok, err := ts.Token()
		if err != nil {
			log.Printf("token mint error: %v", err)
			http.Error(w, "upstream auth error", http.StatusUnauthorized)
			return
		}
		up.Header.Set("Authorization", "Bearer "+tok.AccessToken)

		resp, err := client.Do(up)
		if err != nil {
			log.Printf("proxy error: %v", err)
			http.Error(w, "upstream request failed", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response
		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("copy body error: %v", err)
		}
	}

	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "oidc-proxy OK\nUpstream: %s\nAudience: %s\n", upstream, audience)
	})
	http.HandleFunc("/", reverseProxy)

	dump := getEnvBool("LOG_STARTUP_DUMP", true)
	if dump {
		log.Printf("Starting oidc-proxy on %s -> %s (aud=%s)", listenAddr, upstream, audience)
	}
	log.Fatal(http.ListenAndServe(listenAddr, loggingMiddleware(http.DefaultServeMux)))
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getEnvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		var n int
		_, err := fmt.Sscanf(v, "%d", &n)
		if err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		v = strings.ToLower(v)
		return v == "1" || v == "true" || v == "yes"
	}
	return def
}

func parseCSV(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, http.CanonicalHeaderKey(p))
		}
	}
	return out
}

func singleJoin(a, b string) string {
	if strings.HasSuffix(a, "/") && strings.HasPrefix(b, "/") {
		return a + strings.TrimPrefix(b, "/")
	}
	if !strings.HasSuffix(a, "/") && !strings.HasPrefix(b, "/") {
		return a + "/" + b
	}
	return a + b
}

func copyHeaders(dst, src http.Header) {
	for k, vals := range src {
		for _, v := range vals {
			dst.Add(k, v)
		}
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("LOG_REQUESTS") != "" {
			if dump, err := httputil.DumpRequest(r, false); err == nil {
				log.Printf("REQ %s", strings.ReplaceAll(string(dump), "\n", " | "))
			}
		}
		next.ServeHTTP(w, r)
	})
}
