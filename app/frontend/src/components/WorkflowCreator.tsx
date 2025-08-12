import React, { useState } from 'react'
import { PlusIcon, SparklesIcon } from '@heroicons/react/24/outline'
import { sohoaasApi } from '../services/api'

interface WorkflowCreatorProps {
  onWorkflowCreated: (workflow: any) => void
}

export const WorkflowCreator: React.FC<WorkflowCreatorProps> = ({ onWorkflowCreated }) => {
  const [userMessage, setUserMessage] = useState('')
  const [isCreating, setIsCreating] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!userMessage.trim()) return

    setIsCreating(true)
    try {
      console.log('Creating workflow with user message:', userMessage)
      
      // Call the real workflow generation API
      const response = await sohoaasApi.generateWorkflow(userMessage)
      
      if (response && response.agent_response) {
        console.log('Workflow generation successful:', response)
        
        // Create workflow object from API response
        const workflow = {
          id: response.agent_response.output?.workflow_file?.id || Date.now().toString(),
          message: userMessage,
          status: response.agent_response.error ? 'error' : 'created',
          timestamp: new Date().toISOString(),
          filename: response.agent_response.output?.workflow_file?.filename,
          path: response.agent_response.output?.workflow_file?.path,
          response: response.agent_response
        }
        
        onWorkflowCreated(workflow)
        setUserMessage('')
      } else {
        throw new Error('No valid response from workflow generation API')
      }
    } catch (error) {
      console.error('Failed to create workflow:', error)
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred'
      alert(`Failed to create workflow: ${errorMessage}`)
    } finally {
      setIsCreating(false)
    }
  }

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="flex items-center space-x-2 mb-4">
        <SparklesIcon className="h-6 w-6 text-blue-600" />
        <h2 className="text-lg font-semibold text-gray-900">Create Workflow</h2>
      </div>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="workflow-message" className="block text-sm font-medium text-gray-700 mb-2">
            Describe what you want to automate
          </label>
          <textarea
            id="workflow-message"
            value={userMessage}
            onChange={(e) => setUserMessage(e.target.value)}
            placeholder="e.g., Every Friday, send review reminders with document links"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            rows={3}
            disabled={isCreating}
          />
        </div>
        
        <button
          type="submit"
          disabled={!userMessage.trim() || isCreating}
          className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <PlusIcon className="h-4 w-4" />
          <span>{isCreating ? 'Creating...' : 'Create Workflow'}</span>
        </button>
      </form>
    </div>
  )
}