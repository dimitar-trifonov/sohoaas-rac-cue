package rac

// intent_spec.cue — Single source of truth for SOHOAAS demo intent authoring
// Versioned spec intentionally abstracts away concrete service/function names.
// Custom GPT should reference ONLY this file; do not hardcode services in prompts.
// Demo scope: Google Workspace services (Gmail, Drive, Docs, Calendar). End‑users need
// not remember services—write a single‑paragraph, plain‑text intent and the system
// will generate a sequential workflow. The Custom GPT is the only user‑facing
// guidance source and must rely on this file as Knowledge.

IntentSpec: {
  version: "0.1.2"

  // Execution constraints for the demo
  execution: {
    mode: "sequential" // PoC: no parallelism
    dependencies: "explicit when an action uses a prior action's outputs"
    batch: "not_supported" // No batch/bulk actions in the current demo
  }

  // Canonical abstract actions (service-agnostic)
  // Concrete MCP mappings live elsewhere and can evolve without changing this file.
  actions: {
    CreateDocument: "create_document"
    ShareResource:  "share_resource"
    SendMessage:    "send_message"
    CreateEvent:    "create_event"
  }

  // User-facing guidance (authoritative; UI and Custom GPT can echo this verbatim)
  guidance: {
    demoScope:       "Google Workspace services: Gmail, Drive, Docs, Calendar"
    userInstruction: "Write a single-paragraph, plain-text intent about what you want to do with those services; we will generate a sequential workflow for you."
    sourceOfTruth:   "Custom GPT referencing rac/intent_spec.cue only"
    coaching:        "This demo automates Google Workspace (Gmail, Drive, Docs, Calendar). Provide a single-paragraph intent; the system will generate a sequential workflow. If critical details are missing, ask one precise follow-up; otherwise proceed. Output must be one paragraph, plain text, with ordering hints; no lists/JSON/service names/batch/bulk."
  }

  // Parameter reference convention used by downstream analysis/generation.
  // The Custom GPT should not emit these tokens, but can hint dependencies in text.
  parameterReferences: {
    user:     "${user.<param>}"
    steps:    "${steps.<step_id>.outputs.<field>}"
    computed: "${computed.<expr>}"
  }

  // Required content qualities for natural-language intents
  intentTextRules: {
    // The intent MUST be a short, plain-text description (no JSON, no lists, no code blocks).
    format: "plain_text"

    mustInclude: [
      // Clear goal and expected outcomes
      "goal_and_outcomes",
      // Concrete details when applicable (examples):
      "recipient_emails_when_messaging",
      "dates_times_timezone_for_events",
      "document_titles_and_content_outline",
      "share_permissions_and_audience",
      // Ordering hints when actions depend on prior ones, e.g.,
      // "after creating the document, share it, then email the link"
      "ordering_hints_for_dependencies",
    ]

    mustAvoid: [
      "json_or_structured_formats",
      "explicit_service_or_function_names",
      "parallel_actions",
      "batch_or_bulk_actions",
      "ambiguous_placeholders_without_follow_up",
    ]

    style: {
      length: "1-3 sentences or a compact paragraph"
      tone:   "concise, business-friendly, testable"
    }

    missingInfoBehavior: "If critical details are missing, ask a precise follow-up question (single sentence)."
  }

  // Examples (authoritative; used by Custom GPT for few-shot understanding)
  examples: [
    #"Create a project kickoff document titled \"Project Phoenix – Kickoff\" with sections: Objectives, Scope, Timeline, Owners. After creating it, share with team@acme.com as reader, then email the link to manager@acme.com with subject \"Phoenix Kickoff Doc\" and a short note requesting review."#,
    #"Draft a meeting summary document titled \"Customer Sync – 2025-09-15 09:00 EEST\" using these points: onboarding feedback, support SLAs, next steps. After it is created, share with alice@acme.com as commenter, then email the link to alice@acme.com and cc bob@acme.com with subject \"Customer Sync Summary\"."#,
    #"Schedule a 30-minute weekly review on 2025-09-18 10:00 EEST titled \"Marketing Weekly\" with attendee lead@acme.com. Create an agenda document \"Marketing Weekly – 2025-09-18\" with sections: Wins, Metrics, Blockers. Share the document as reader with the attendee and include the link in the event description."#,
  ]
}

// Optional structured intent schema for downstream validators and tests.
// The Custom GPT does NOT need to output this format, but the pipeline may
// normalize text into this for validation and execution.
StructuredIntent: {
  title:       string & != ""
  description: string & != ""

  user_parameters?: [string]: {
    type:        "string" | "number" | "boolean" | "array" | "object"
    description: string & != ""
    required?:   bool
    enum?:       [...string]
    examples?:   [...string]
  }

  steps: [...Step]

  success_criteria?: [...string]
}

Step: {
  id:          string & =~ "^[a-z0-9]+(-[a-z0-9]+)*$"
  action:      IntentSpec.actions.CreateDocument | IntentSpec.actions.ShareResource | IntentSpec.actions.SendMessage | IntentSpec.actions.CreateEvent
  description: string & != ""
  parameters:  { _ }: _
  depends_on?: [...string]
}
