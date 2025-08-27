package sohoaas_intent

#Meta: {
  name:    "SOHOAAS Demo — Intent RaC"
  version: "1.2.0"
  purpose: "RaC model for validating, coaching, and binding user intents"
}

// ========== STATE ==========
#ActionCatalog: {
  Categories: {
    "Create & Organize": ["create_document", "create_event", "create_folder"]
    "Share & Access":    ["share_resource", "grant_access", "revoke_access"]
    "Communicate":       ["send_message", "draft_message"]
    "Search & Collect":  ["find_messages", "list_items"]
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
    if =~"(?i)share|access|link".match(text) {"share_resource"},
    if =~"(?i)email|notify|message".match(text) {"send_message"},
    if =~"(?i)meeting|event|schedule".match(text) {"create_event"},
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
  "create_document": { provider: "workspace", service: "docs",    fn: "create_document" }
  "share_resource":  { provider: "workspace", service: "drive",   fn: "share_file" }
  "send_message":    { provider: "workspace", service: "gmail",   fn: "send_message" }
  "create_event":    { provider: "workspace", service: "calendar",fn: "create_event" }
}

#Resolver(mcp: #MCP, a: string): {
  let b = #Bindings[a]
  fnDef: mcp.providers[b.provider][b.service].functions[b.fn]
  required_fields: fnDef.required_fields | *[]
  example_payload: fnDef.example_payload | *{}
}

// Progressive coaching helpers
#Logic.WithBindings: {
  AbstractOverview: [...string] & [
    for cat, list in #ActionCatalog.Categories {
      "\(cat): " + strings.Join(list, ", ")
    }
  ]

  ParamHints(mcp: #MCP, actions: [...string]): [...string] & [
    for a in actions {
      let r = #Resolver(mcp, a)
      if len(r.required_fields) > 0 {
        "For \(a), include: \(strings.Join(r.required_fields, \", \"))."
      }
    }
  ]

  IsDiscovery(text: string): bool & (
    =~"(?i)what'?s possible|what can i do|show me|help me discover".match(text)
  )
}

// ========== HANDLERS ==========
#Handlers: {
  "user.intent.submit"(in: #Events."user.intent.submit".input, mcp: #MCP): #Events."user.intent.submit".output & {
    let v    = #Logic.Validate(in.text)
    let acts = #Logic.InferActions(in.text)

    let abstract = if #Logic.WithBindings.IsDiscovery(in.text) {
      #Logic.WithBindings.AbstractOverview
    } else { [] }

    let details = if (!#Logic.WithBindings.IsDiscovery(in.text) && v.valid) {
      #Logic.WithBindings.ParamHints(mcp, acts)
    } else { [] }

    valid:  v.valid
    errors: v.errors
    coach:  #Logic.Coaching(v.errors) + abstract + details

    parsed: if valid {
      {
        actions: acts
        steps:   ["Interpret the intent step by step as written."]
      }
    }
  }

  "system.intent.parse"(in: #Events."system.intent.parse".input): #Events."system.intent.parse".output & {
    actions: #Logic.InferActions(in.text)
    steps:   ["Sequential steps derived from the text."]
  }
}

// ========== TESTS ==========
#Tests: {
  Valid: [
    {
      name: "Document then share"
      in:   { text: "Draft a new document and then share it with my team via a link." }
      out:  { valid: true }
    },
    {
      name: "Meeting and notify"
      in:   { text: "Schedule a meeting tomorrow and send a message to attendees." }
      out:  { valid: true }
    },
  ]

  Invalid: [
    {
      name: "Too short"
      in:   { text: "Make doc." }
      out:  { valid: false coach: ["Try writing a longer, clearer description."] }
    },
    {
      name: "JSON-like text"
      in:   { text: "{action: create_document}" }
      out:  { valid: false coach: ["Write in plain sentences, not code or JSON."] }
    },
  ]
}
