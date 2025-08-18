#!/bin/bash

# Workflow Prompt CUE Validation Test
# Validates that workflow-prompt.cue reflects our enhanced MCP response schema implementation

echo "🔍 Workflow Prompt CUE Validation Test"
echo "======================================"

WORKFLOW_PROMPT_FILE="/home/dimitar/dim/rac/sohoaas/rac/agents/prompts/workflow-prompt.cue"
PASSED=0
FAILED=0

# Test 1: Validate CUE syntax
echo "Test 1: CUE Syntax Validation"
if cue vet "$WORKFLOW_PROMPT_FILE" >/dev/null 2>&1; then
    echo "✅ CUE syntax is valid"
    ((PASSED++))
else
    echo "❌ CUE syntax validation failed"
    ((FAILED++))
fi

# Test 2: Check for correct Gmail function (send_message not get_message)
echo "Test 2: Gmail Function Name Validation"
if grep -q "gmail.send_message" "$WORKFLOW_PROMPT_FILE"; then
    echo "✅ Uses correct gmail.send_message function"
    ((PASSED++))
else
    echo "❌ Should use gmail.send_message function"
    ((FAILED++))
fi

# Test 3: Check for correct Gmail output fields matching MCP server
echo "Test 3: Gmail Output Fields Validation"
if grep -q "message_id.*thread_id.*status.*sent_at" "$WORKFLOW_PROMPT_FILE"; then
    echo "✅ Gmail output fields match MCP server implementation"
    ((PASSED++))
else
    echo "❌ Gmail output fields should match MCP server: message_id, thread_id, status, sent_at"
    ((FAILED++))
fi

# Test 4: Check for correct Docs output fields (url not document_url)
echo "Test 4: Docs Output Fields Validation"
if grep -q '"url"' "$WORKFLOW_PROMPT_FILE" && ! grep -q '"document_url"' "$WORKFLOW_PROMPT_FILE"; then
    echo "✅ Docs output fields use correct 'url' field name"
    ((PASSED++))
else
    echo "❌ Docs should use 'url' not 'document_url' to match MCP server"
    ((FAILED++))
fi

# Test 5: Check for Calendar output fields
echo "Test 5: Calendar Output Fields Validation"
if grep -q "event_id.*html_link.*created_at" "$WORKFLOW_PROMPT_FILE"; then
    echo "✅ Calendar output fields include enhanced response schema fields"
    ((PASSED++))
else
    echo "❌ Calendar should include event_id, html_link, created_at fields"
    ((FAILED++))
fi

# Test 6: Check for correct step output references
echo "Test 6: Step Output References Validation"
if grep -q '${steps\..*\.outputs\..*}' "$WORKFLOW_PROMPT_FILE"; then
    echo "✅ Contains proper step output references format"
    ((PASSED++))
else
    echo "❌ Should contain step output references: \${steps.step_id.outputs.field}"
    ((FAILED++))
fi

# Test 7: Check for required fields matching MCP server
echo "Test 7: Required Fields Validation"
if grep -q 'required_fields.*"to".*"subject".*"body"' "$WORKFLOW_PROMPT_FILE"; then
    echo "✅ Gmail required fields match MCP server implementation"
    ((PASSED++))
else
    echo "❌ Gmail required fields should be: to, subject, body"
    ((FAILED++))
fi

# Summary
echo ""
echo "======================================"
echo "Test Results: $PASSED passed, $FAILED failed"

if [ $FAILED -eq 0 ]; then
    echo "🎉 SUCCESS: All workflow prompt validations passed!"
    echo "✅ Workflow prompt CUE file correctly reflects enhanced MCP response schemas"
    exit 0
else
    echo "❌ FAILURE: $FAILED validation(s) failed"
    echo "The workflow prompt CUE file needs additional updates"
    exit 1
fi
