package workflow

workflow: {
	version: "1.0"
	name: "Example Workflow"
	description: "Example workflow for parsing tests"
	user_parameters: {
		document_title: {
			type: "string"
			prompt: "Title of the Google Doc to create"
			required: true
			default: "Project Delta – Kickoff Notes"
		}
		event_start_time: {
			type: "datetime"
			prompt: "Start time (ISO 8601)"
			required: true
			default: ""
		}
		event_end_time: {
			type: "datetime"
			prompt: "End time (ISO 8601)"
			required: true
			default: ""
		}
	}
	steps: [{
		id: "create_doc"
		name: "Create Google Doc"
		service: "docs"
		action: "create_document"
		parameters: {
			title: "${user.document_title}"
		}
	},{
		id: "create_event"
		name: "Create Calendar Event"
		service: "calendar"
		action: "create_event"
		parameters: {
			start_time: "${user.event_start_time}"
			end_time: "${user.event_end_time}"
			title: "Kickoff: ${user.document_title}"
		}
	}]
	service_bindings: {}
	execution_config: {
		mode: "sequential"
		timeout: "5m"
		environment: "development"
	}
}
