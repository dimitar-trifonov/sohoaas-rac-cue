package workflow

// Deterministic example that conforms to rac/schemas/deterministic_workflow.cue
// It intentionally uses the schema conjunction so our parser's sanitizer path
// is exercised: `#DeterministicWorkflow & { ... }`.

workflow: #DeterministicWorkflow & {
	version:     "1.0.0"
	name:        "Deterministic Docs Share and Notify"
	description: "Create a doc, share it, and notify collaborator via email"
	original_intent: "Create a Google Doc titled 'RaC Spec', share with test@example.com, and email them the link"

	steps: [
		{
			id:     "create_doc"
			name:   "Create Google Doc"
			action: "docs.create_document"
			parameters: {
				title: "${user.document_title}"
			}
			outputs: {
				document_id: {
					type:        "string"
					description: "Document ID"
				}
				url: {
					type:        "string"
					description: "Document URL"
				}
			}
			_mcp_service_type: "docs"
		},
		{
			id:     "share_doc"
			name:   "Share Document"
			action: "drive.share_file"
			parameters: {
				file_id: "${steps.create_doc.outputs.document_id}"
				email:   "${user.collaborator_email}"
				role:    "writer"
			}
			depends_on: ["create_doc"]
			outputs: {
				share_url: {
					type:        "string"
					description: "Shared URL"
				}
			}
			_mcp_service_type: "drive"
		},
		{
			id:     "notify"
			name:   "Send Notification"
			action: "gmail.send_message"
			parameters: {
				to:      "${user.collaborator_email}"
				subject: "Document Ready: ${user.document_title}"
				body:    "The document is available here: ${steps.share_doc.outputs.share_url}"
			}
			depends_on: ["share_doc"]
			_mcp_service_type: "gmail"
		},
	]

	user_parameters: {
		document_title: {
			type:       "string"
			required:   true
			prompt:     "Title of the document"
			default:    "RaC Spec"
		}
		collaborator_email: {
			type:       "string"
			required:   true
			prompt:     "Email to share with"
			validation: "email"
		}
	}

	service_bindings: {
		drive: {
			type:     "mcp_service"
			provider: "workspace"
			auth: {
				method: "oauth2"
				oauth2: {
					scopes: ["https://www.googleapis.com/auth/drive.file"]
					token_source: "user"
				}
			}
		}
		docs: {
			type:     "mcp_service"
			provider: "workspace"
			auth: {
				method: "oauth2"
				oauth2: {
					scopes: ["https://www.googleapis.com/auth/documents"]
					token_source: "user"
				}
			}
		}
		gmail: {
			type:     "mcp_service"
			provider: "workspace"
			auth: {
				method: "oauth2"
				oauth2: {
					scopes: ["https://www.googleapis.com/auth/gmail.send"]
					token_source: "user"
				}
			}
		}
	}

	execution_config: {
		mode:        "sequential"
		timeout:     "5m"
		environment: "development"
	}
}
