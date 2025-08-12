package rac

// =============================================
// 🔹 CORE RAC SCHEMA DEFINITIONS
// =============================================

#State: {
    id:        string & !=""
    type:      "object" | "array" | "primitive"
    fields?: [...{
        name:     string
        type:     string
        required?: bool | *false
        format?:  string
    }]
    access?: {
        read?:  string
        write?: string
    }
    version?: string
    metadata?: {
        createdBy?: string
        tags?: [...string]
    }
}

#Event: {
    id:          string & !=""
    version:     string
    type:        "user" | "system"
    description: string
    events?: [...{
        id:          string
        type:        "user" | "system"
        description: string
        triggers?: [...{
            source: { type: "user" | "system" | "custom", name: string }
            formId?: string
            componentId?: string
            stateId?: string
            params?: { threshold?: number, delay?: number, errorCode?: string }
        }]
        logicChecks?: [...{
            type: "and" | "or"
            checks: [...{
                logic: string
                onError?: "halt" | "continue"
            }]
        }]
        actions?: [...{
            type: "validate" | "update" | "create" | "delete"
            state?: string
            fields?: [...string]
        }]
        metadata?: { createdBy?: string, tags?: [...string] }
    }]
}

#Logic: {
    id: string & !=""
    appliesTo?: string
    rules?: [...{
        if:   string
        then: [...{ error: string }]
    }]
    guards?:  [...string]
    effects?: [...string]
    version?: string
    metadata?: {
        createdBy?: string
        tags?: [...string]
    }
}

#UI: {
    id:          string & !=""
    description: string
    bindsTo?:    string
    components?: [...{
        type:   string
        fields: [...{
            name:        string
            type:        string
            description: string
        }]
        submitEvent?: string
    }]
    views?: [...{
        path?:   string
        layout?: string
    }]
    logicPreview?: string
    version?: string
    metadata?: {
        createdBy?: string
        tags?: [...string]
    }
}

#Test: {
    id:    string & !=""
    setup?: {}
    testCases?: [...{
        id:          string & !=""
        description: string
        setup?:      {}
        input?: {
            event?: string
            data?:  {...}
        }
        expected?:   {...}
        expectError?: string
        notes?:      string
    }]
    version?: string
    metadata?: {
        createdBy?: string
        tags?: [...string]
    }
}
