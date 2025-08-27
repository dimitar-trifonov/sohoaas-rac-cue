package sohoaas_intent

#Meta: {
  name:    "SOHOAAS Demo — Intent Cheat‑Sheet"
  version: "1.0.0"
  purpose: "Renderable cheat‑sheet for the Custom GPT coach: discovery lines, one‑liners, field hints, and copy‑ready templates. Demo‑aware (bindings + enabled)."
}

// ========= INPUT STATE (runtime) =========
#State: {
  // Optional: if provided, only these actions are shown (demo on/off switch)
  EnabledActions?: [...string]
}

// ========= WORKSPACE BINDINGS =========
#MCP: {
  providers: { [provider=string]: { [service=string]: { functions: { [fn=string]: {
    name: string
    display_name?: string
    description?: string
    required_fields?: [...string]
    example_payload?: _
  }}}}}
}

#Bindings: {
  // Calendar
  "create_event": { provider: "workspace", service: "calendar", fn: "create_event" }
  "delete_event": { provider: "workspace", service: "calendar", fn: "delete_event" }
  "get_event":    { provider: "workspace", service: "calendar", fn: "get_event" }
  "list_events":  { provider: "workspace", service: "calendar", fn: "list_events" }
  "update_event": { provider: "workspace", service: "calendar", fn: "update_event" }
  // Docs
  "batch_update":     { provider: "workspace", service: "docs", fn: "batch_update" }
  "create_document":  { provider: "workspace", service: "docs", fn: "create_document" }
  "get_document":     { provider: "workspace", service: "docs", fn: "get_document" }
  "insert_text":      { provider: "workspace", service: "docs", fn: "insert_text" }
  "update_document":  { provider: "workspace", service: "docs", fn: "update_document" }
  // Drive
  "create_folder": { provider: "workspace", service: "drive", fn: "create_folder" }
  "get_file":      { provider: "workspace", service: "drive", fn: "get_file" }
  "list_files":    { provider: "workspace", service: "drive", fn: "list_files" }
  "move_file":     { provider: "workspace", service: "drive", fn: "move_file" }
  "share_file":    { provider: "workspace", service: "drive", fn: "share_file" }
  "upload_file":   { provider: "workspace", service: "drive", fn: "upload_file" }
  // Gmail
  "get_message":    { provider: "workspace", service: "gmail", fn: "get_message" }
  "list_messages":  { provider: "workspace", service: "gmail", fn: "list_messages" }
  "search_messages":{ provider: "workspace", service: "gmail", fn: "search_messages" }
  "send_message":   { provider: "workspace", service: "gmail", fn: "send_message" }
  // Legacy alias kept for compatibility
  "share_resource": { provider: "workspace", service: "drive", fn: "share_file" }
}

#Resolver(mcp: #MCP, a: string): {
  let b = #Bindings[a]
  fnDef: mcp.providers[b.provider][b.service].functions[b.fn]
  required_fields: fnDef.required_fields | *[]
}

// ========= CATALOG (what to show and how) =========
#Catalog: {
  Categories: {
    "Create & Organize": ["create_event", "delete_event", "update_event", "batch_update", "create_document", "insert_text", "update_document", "create_folder", "move_file", "upload_file"]
    "Share & Access":    ["share_file"]
    "Communicate":       ["send_message"]
    "Search & Collect":  ["get_event", "list_events", "get_document", "get_file", "list_files", "get_message", "list_messages", "search_messages"]
  }
}

#Cheat: {
  // One‑liner marketing descriptions per action (service‑agnostic)
  QuickLine(a: string): string & {
    if a == "create_event"   { "Schedule a meeting with a title and time." }
    if a == "delete_event"   { "Remove a meeting from the calendar." }
    if a == "update_event"   { "Change the meeting’s time or details." }
    if a == "list_events"    { "See upcoming meetings in a time range." }
    if a == "get_event"      { "Show details for a specific meeting." }

    if a == "create_document"{ "Start a new document by title." }
    if a == "insert_text"    { "Add text into an existing document." }
    if a == "update_document"{ "Edit document properties or structure." }
    if a == "batch_update"   { "Apply multiple edits to a doc at once." }

    if a == "create_folder"  { "Create a new folder for your files." }
    if a == "upload_file"    { "Upload a file to your drive." }
    if a == "move_file"      { "Move a file into a folder." }
    if a == "share_file"     { "Share a file or folder with people." }
    if a == "get_file"       { "See details for a file by id." }
    if a == "list_files"     { "Browse files that match a filter." }

    if a == "send_message"   { "Email people with a subject and body." }
    if a == "get_message"    { "Open one email by id." }
    if a == "list_messages"  { "List recent emails." }
    if a == "search_messages"{ "Find emails by query (from:, subject:, etc.)." }

    // fallback
    "Use this action in a simple sentence."
  }

  // Copy‑ready natural language template per action
  Template(a: string): string & {
    if a == "create_event"   { "Create an event titled \"{title}\" from {startTime} to {endTime}." }
    if a == "update_event"   { "Update event {eventId} to run from {startTime} to {endTime}." }
    if a == "delete_event"   { "Delete event {eventId}." }
    if a == "list_events"    { "List events between {startTime} and {endTime}." }
    if a == "get_event"      { "Get details for event {eventId}." }

    if a == "create_document"{ "Create a document titled \"{title}\"." }
    if a == "insert_text"    { "Insert this text into document {documentId}: {text}." }
    if a == "update_document"{ "Update document {documentId} with: {changes}." }
    if a == "batch_update"   { "Apply multiple edits to document {documentId}." }

    if a == "create_folder"  { "Create a folder named \"{name}\"." }
    if a == "upload_file"    { "Upload {filepath} into folder {folderId}." }
    if a == "move_file"      { "Move file {fileId} into folder {folderId}." }
    if a == "share_file"     { "Share {fileId} with {recipients}." }
    if a == "get_file"       { "Get details for file {fileId}." }
    if a == "list_files"     { "List files where {filter}." }

    if a == "send_message"   { "Send a message to {to} with subject \"{subject}\": {body}." }
    if a == "get_message"    { "Get message {messageId}." }
    if a == "list_messages"  { "List my latest {count} messages." }
    if a == "search_messages"{ "Search messages with query: {query}." }

    // fallback
    "Write a simple sentence that uses required fields."
  }

  // Compose a cheat‑sheet line with fields from MCP (demo‑aware)
  Line(mcp: #MCP, a: string): string & {
    let r = #Resolver(mcp, a)
    let fields = if len(r.required_fields) > 0 {
      " [fields: " + strings.Join(r.required_fields, ", ") + "]"
    } else { "" }
    #Cheat.QuickLine(a) + fields
  }
}

// ========= HANDLERS =========
#Handlers: {
  // Returns a list of category lines, each action as one line: QuickLine + [fields]
  "cheatsheet.get"(mcp: #MCP): {
    let enabled = #State.EnabledActions | *[]
    lines: [...string] & [
      for cat, list in #Catalog.Categories {
        "== " + cat + " ==",
        for a in list if (len(enabled) == 0 || (len([for x in enabled if x == a {x}]) > 0)) {
          "- " + #Cheat.Line(mcp, a)
        },
        ""
      }
    ]
  }

  // Returns a single action’s copy‑ready example sentence
  "cheatsheet.example"(a: string): { text: string & #Cheat.Template(a) }
}
