package workflow

// Real generated workflow example captured from generated_workflows
// Serves as an end-to-end parsing fixture.

workflow: {
	version: "1.0.0"
	name: "Generated Workflow"
	description: "Creates a folder for Project Delta, adds a kickoff Google Doc inside it, and schedules a calendar event with the document link in the description."
	original_intent: "Create a folder called 'Project Delta Docs', add a Google Doc titled 'Project Delta – Kickoff Notes' inside it, then schedule a calendar event for tomorrow at 10:00 titled 'Project Delta Kickoff' and include the document link in the event description"

	steps: [
		{
			id: "create_project_folder"
			name: "Step 1"
			action: "drive.create_folder"
			parameters: {
				name: "${user.folder_name}"
			}
		},
		{
			id: "create_kickoff_doc"
			name: "Step 2"
			action: "docs.create_document"
			parameters: {
				title: "${user.document_title}"
			}
		},
		{
			id: "move_doc_to_folder"
			name: "Step 3"
			action: "drive.move_file"
			parameters: {
				file_id: "${steps.create_kickoff_doc.outputs.document_id}"
				new_parent_id: "${steps.create_project_folder.outputs.folder_id}"
			}
		},
		{
			id: "schedule_kickoff_event"
			name: "Step 4"
			action: "calendar.create_event"
			parameters: {
				title: "${user.event_title}"
				description: "Kickoff notes: ${steps.create_kickoff_doc.outputs.url}"
				endTime: "${user.event_end_time}"
				startTime: "${user.event_start_time}"
			}
		}
	]

	user_parameters: {
		event_start_time: {
			type: "datetime"
			prompt: "Start time for the kickoff event (ISO 8601 format)"
			required: true
			default: ""
		}
		event_title: {
			type: "string"
			prompt: "Title of the calendar event for the kickoff meeting"
			required: true
			default: "Project Delta Kickoff"
		}
		folder_name: {
			type: "string"
			prompt: "Name of the folder to create for Project Delta documents"
			required: true
			default: "Project Delta Docs"
		}
		document_title: {
			type: "string"
			prompt: "Title of the Google Doc to create for kickoff notes"
			required: true
			default: "Project Delta – Kickoff Notes"
		}
		event_end_time: {
			type: "datetime"
			prompt: "End time for the kickoff event (ISO 8601 format)"
			required: true
			default: ""
		}
	}

	service_bindings: {}
	execution_config: {
		mode: "sequential"
		timeout: "5m"
		environment: "development"
	}
}
