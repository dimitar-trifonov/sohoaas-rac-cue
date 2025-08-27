package sohoaas_intent

#Meta: {
  name:    "SOHOAAS Demo — Intent RaC"
  version: "1.4.0"
  purpose: "RaC model for validating, demo-aware coaching, and binding user intents"
}

// ========== STATE ==========
#ActionCatalog: {
  Categories: {
    "Create & Organize": ["create_event", "delete_event", "update_event", "batch_update", "create_document", "insert_text", "update_document", "create_folder", "move_file", "upload_file"]
    "Share & Access": ["share_file"]
    "Communicate": ["send_message"]
    "Search & Collect": ["get_event", "list_events", "get_document", "get_file", "list_files", "get_message", "list_messages", "search_messages"]
  }
}

#State: {
  AllowedActions: [...string] & [
    for _, list in #ActionCatalog.Categories
      for a in list { a }
  ]
  IntentText: string
  ParsedIntent: {
    actions?: [string]
    steps?:   [...string]
  }
  EnabledActions?: [...string]
}

// ========== EVENTS ==========
#Events: {
  "user.intent.submit": {
    input:  { text: string }
    output: {
      valid:  bool
      errors: [...string] | *[]
      parsed: #State.ParsedIntent | *{}
      coach?: [...string]
    }
  }

  "system.intent.parse": {
    input:  { text: string }
    output: #State.ParsedIntent
  }
}

// ========== LOGIC ==========
#Logic: {
  Constraints: {
    minChars:   20
    maxChars:   600
    forbidJSON: true
    oneParagraph: true
  }

  Rx: {
    hasJSON:    =~"[\\{\\}\\[\\]]"
    hasNewline: =~"(\\r?\\n){2,}"
  }

  Validate(text: string): {
    errors: [...string] & [
      if len(text) < Constraints.minChars {"intent too short"},
      if len(text) > Constraints.maxChars {"intent too long"},
      if Constraints.forbidJSON && Rx.hasJSON.match(text) {"avoid JSON-like text"},
      if Constraints.oneParagraph && Rx.hasNewline.match(text) {"use one paragraph"},
    ]
    valid: len(errors) == 0
  }

  InferActions(text: string): [...string] & [
    if =~"(?i)document|draft|write".match(text) {"create_document"},
    if =~"(?i)update".match(text) {"update_document"},
    if =~"(?i)list documents".match(text) {"list_documents"},
    if =~"(?i)folder".match(text) {"create_folder"},
    if =~"(?i)event".match(text) {"create_event"},
    if =~"(?i)update event".match(text) {"update_event"},
    if =~"(?i)list events".match(text) {"list_events"},
    if =~"(?i)share|access|link".match(text) {"share_resource"},
    if =~"(?i)grant".match(text) {"grant_access"},
    if =~"(?i)revoke".match(text) {"revoke_access"},
    if =~"(?i)email|notify|message".match(text) {"send_message"},
    if =~"(?i)list messages".match(text) {"list_messages"},
    if =~"(?i)get message".match(text) {"get_message"},
  ]

  Coaching(errors: [...string]): [...string] & [
    for e in errors {
      if e == "intent too short" {"Try writing a longer, clearer description."}
      if e == "intent too long" {"Shorten your text to the main idea."}
      if e == "avoid JSON-like text" {"Write in plain sentences, not code or JSON."}
      if e == "use one paragraph" {"Keep everything in a single paragraph."}
    }
  ]
}

// ========== BINDINGS (from MCP metadata) ==========
#MCP: {
  providers: {
    [providerName=string]: {
      [serviceName=string]: {
        functions: {
          [fnName=string]: {
            name:           string
            display_name?:  string
            description?:   string
            example_payload?: _
            required_fields?: [...string]
          }
        }
      }
    }
  }
}

#Bindings: {
  "share_resource": { provider: "workspace", service: "drive", fn: "share_file" }
  "create_document": { provider: "workspace", service: "docs", fn: "create_document" }
  "send_message": { provider: "workspace", service: "gmail", fn: "send_message" }
  "create_event": { provider: "workspace", service: "calendar", fn: "create_event" }
  "batch_update": { provider: "workspace", service: "docs", fn: "batch_update" }
  "get_document": { provider: "workspace", service: "docs", fn: "get_document" }
  "insert_text": { provider: "workspace", service: "docs", fn: "insert_text" }
  "update_document": { provider: "workspace", service: "docs", fn: "update_document" }
  "create_folder": { provider: "workspace", service: "drive", fn: "create_folder" }
  "get_file": { provider: "workspace", service: "drive", fn: "get_file" }
  "list_files": { provider: "workspace", service: "drive", fn: "list_files" }
  "move_file": { provider: "workspace", service: "drive", fn: "move_file" }
  "share_file": { provider: "workspace", service: "drive", fn: "share_file" }
  "upload_file": { provider: "workspace", service: "drive", fn: "upload_file" }
  "get_message": { provider: "workspace", service: "gmail", fn: "get_message" }
  "list_messages": { provider: "workspace", service: "gmail", fn: "list_messages" }
  "search_messages": { provider: "workspace", service: "gmail", fn: "search_messages" }
  "delete_event": { provider: "workspace", service: "calendar", fn: "delete_event" }
  "get_event": { provider: "workspace", service: "calendar", fn: "get_event" }
  "list_events": { provider: "workspace", service: "calendar", fn: "list_events" }
  "update_event": { provider: "workspace", service: "calendar", fn: "update_event" }
}

#Resolver(mcp: #MCP, a: string): {
  let b = #Bindings[a]
  fnDef: mcp.providers[b.provider][b.service].functions[b.fn]
  required_fields: fnDef.required_fields | *[]
  example_payload: fnDef.example_payload | *{}
}

// Progressive & demo-aware coaching helpers
#Logic.WithBindings: {
  BoundActions: [...string] & [ for k,_ in #Bindings { "\(k)" } ]

  _in(a: string, list: [...string]): bool & (len([ for x in list if x == a { x } ]) > 0)

  _isBoundEnabled(a: string, enabled: [...string] | *[]): bool & (
    _in(a, BoundActions) && (len(enabled) == 0 || _in(a, enabled))
  )

  SupportedOverview(mcp: #MCP): [...string] & [
    let enabled = #State.EnabledActions | *[]
    for cat, list in #ActionCatalog.Categories {
      let sup = [ for a in list if _isBoundEnabled(a, enabled) { a } ]
      if len(sup) > 0 { "\(cat): " + strings.Join(sup, ", ") }
    }
  ]

  ParamHintsEnabled(mcp: #MCP, actions: [...string]): [...string] & [
    let enabled = #State.EnabledActions | *[]
    for a in actions if _isBoundEnabled(a, enabled) {
      let r = #Resolver(mcp, a)
      if len(r.required_fields) > 0 {
        "For \(a), include: \(strings.Join(r.required_fields, ", "))."
      }
    }
  ]

  IsDiscovery(text: string): bool & (
    =~"(?i)what'?s possible|what can i do|show me|help me discover".match(text)
  )

  NearestAlternative(a: string): string | *"" & {
    if a == "send_message"  { "send a message via Email" }
    if a == "create_event"  { "create a calendar event" }
    if a == "share_resource"{ "share a file link" }
    if a == "create_document"{ "create a document" }
  }
}

// ========== HANDLERS (unchanged)...
// (rest of file stays the same as previous version)
