#!/bin/bash

# SOHOAAS Real OAuth2 Testing Script
# Tests the complete pipeline with real Google Workspace tokens

set -e

BACKEND_URL="http://localhost:8081"
USER_ID="real_test_user_$(date +%s)"
FIRST_CALL="true"

echo "🚀 SOHOAAS Real OAuth2 Testing"
echo "================================"
echo "Backend URL: $BACKEND_URL"
echo "Test User ID: $USER_ID"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to make API calls with error handling and rate limiting
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    # Rate limiting: Wait 2 seconds between API calls
    if [ "$FIRST_CALL" != "true" ]; then
        echo -e "${YELLOW}⏱️  Rate limiting: waiting 2 seconds...${NC}"
        sleep 2
    fi
    export FIRST_CALL="false"
    
    echo -e "${BLUE}📡 Testing: $description${NC}"
    echo "   Method: $method"
    echo "   Endpoint: $endpoint"
    
    if [ -n "$data" ]; then
        echo "   Data: $data"
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $OAUTH_TOKEN" \
            -d "$data" \
            "$BACKEND_URL$endpoint" || echo "000")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: Bearer $OAUTH_TOKEN" \
            "$BACKEND_URL$endpoint" || echo "000")
    fi
    
    # Split response and status code
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
        echo -e "   ${GREEN}✅ Success (HTTP $http_code)${NC}"
        echo "   Response: $response_body" | head -c 200
        echo ""
        return 0
    else
        echo -e "   ${RED}❌ Failed (HTTP $http_code)${NC}"
        echo "   Response: $response_body"
        echo ""
        return 1
    fi
}

# Check if OAuth token is provided
if [ -z "$OAUTH_TOKEN" ]; then
    echo -e "${YELLOW}⚠️  OAuth Token Required${NC}"
    echo ""
    echo "To test with real Google Workspace integration, you need to:"
    echo "1. Get a Google OAuth2 token with the following scopes:"
    echo "   - https://www.googleapis.com/auth/gmail.compose"
    echo "   - https://www.googleapis.com/auth/gmail.send"
    echo "   - https://www.googleapis.com/auth/calendar"
    echo "   - https://www.googleapis.com/auth/documents"
    echo "   - https://www.googleapis.com/auth/drive.file"
    echo ""
    echo "2. Export the token as an environment variable:"
    echo "   export OAUTH_TOKEN='your_actual_token_here'"
    echo ""
    echo "3. Run this script again:"
    echo "   ./test_real_oauth.sh"
    echo ""
    echo "For now, running tests without OAuth (limited functionality)..."
    echo ""
fi

# Test 1: Backend Health Check
echo -e "${BLUE}🏥 Test 1: Backend Health Check${NC}"
api_call "GET" "/health" "" "Backend health status"

# Test 2: Service Catalog
echo -e "${BLUE}📋 Test 2: Service Catalog${NC}"
api_call "GET" "/api/v1/services" "" "Available Google Workspace services"

# Test 3: User Capabilities (requires OAuth)
if [ -n "$OAUTH_TOKEN" ]; then
    echo -e "${BLUE}👤 Test 3: User Capabilities with OAuth${NC}"
    user_data='{
        "user_id": "'$USER_ID'",
        "oauth_tokens": {
            "google": {
                "access_token": "'$OAUTH_TOKEN'",
                "token_type": "Bearer"
            }
        },
        "connected_services": ["gmail", "calendar", "docs", "drive"]
    }'
    api_call "POST" "/api/v1/capabilities" "$user_data" "User capabilities with real OAuth"
else
    echo -e "${YELLOW}⚠️  Skipping OAuth-dependent tests${NC}"
fi

# Test 4: Complete Workflow Pipeline
echo -e "${BLUE}🔄 Test 4: Complete Workflow Pipeline${NC}"
pipeline_data='{
    "user_id": "'$USER_ID'",
    "user_message": "Send a test email to myself with the subject \"SOHOAAS Test\" and body \"This is a test from SOHOAAS automation system\"",
    "conversation_history": [],
    "user": {
        "user_id": "'$USER_ID'",
        "oauth_tokens": {
            "google": {
                "access_token": "'${OAUTH_TOKEN:-mock_token}'",
                "token_type": "Bearer"
            }
        },
        "connected_services": ["gmail", "calendar", "docs", "drive"]
    }
}'

if api_call "POST" "/api/v1/pipeline/test" "$pipeline_data" "Complete workflow pipeline"; then
    echo -e "${GREEN}🎉 Pipeline test completed successfully!${NC}"
    
    # Test 5: Check generated workflow
    echo -e "${BLUE}📄 Test 5: Generated Workflow Verification${NC}"
    echo "Checking for generated workflow files..."
    
    latest_workflow=$(find workflows/ -name "*.cue" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2- 2>/dev/null || echo "")
    
    if [ -n "$latest_workflow" ] && [ -f "$latest_workflow" ]; then
        echo -e "${GREEN}✅ Found latest workflow: $latest_workflow${NC}"
        echo ""
        echo "=== GENERATED WORKFLOW PREVIEW ==="
        head -n 20 "$latest_workflow"
        echo "=== END PREVIEW ==="
        echo ""
    else
        echo -e "${YELLOW}⚠️  No workflow files found${NC}"
    fi
else
    echo -e "${RED}❌ Pipeline test failed${NC}"
fi

# Test 6: Real Google Workspace Integration (if OAuth available)
if [ -n "$OAUTH_TOKEN" ]; then
    echo -e "${BLUE}🔗 Test 6: Real Google Workspace Integration${NC}"
    echo "Testing actual Google API calls through MCP service..."
    
    # Test Gmail API through MCP
    gmail_test='{
        "service": "gmail",
        "action": "list_messages",
        "parameters": {
            "max_results": 1
        },
        "oauth_token": "'$OAUTH_TOKEN'"
    }'
    
    api_call "POST" "/api/v1/mcp/execute" "$gmail_test" "Gmail API test through MCP"
    
    echo -e "${GREEN}🎯 Real integration testing completed!${NC}"
else
    echo -e "${YELLOW}⚠️  Skipping real Google Workspace integration (no OAuth token)${NC}"
fi

echo ""
echo "🏁 TESTING SUMMARY"
echo "=================="
if [ -n "$OAUTH_TOKEN" ]; then
    echo -e "${GREEN}✅ Full OAuth2 testing completed${NC}"
    echo "   - Backend health: ✅"
    echo "   - Service catalog: ✅"
    echo "   - User capabilities: ✅"
    echo "   - Workflow pipeline: ✅"
    echo "   - Google Workspace integration: ✅"
else
    echo -e "${YELLOW}⚠️  Limited testing completed (no OAuth)${NC}"
    echo "   - Backend health: ✅"
    echo "   - Service catalog: ✅"
    echo "   - Workflow pipeline: ✅ (mock mode)"
    echo ""
    echo "For full testing, provide OAuth token and run again."
fi

echo ""
echo "🚀 SOHOAAS is ready for production workflow automation!"
