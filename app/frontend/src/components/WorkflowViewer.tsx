import React, { useState, useEffect } from 'react'
import { XMarkIcon, PlayIcon, CheckCircleIcon, ExclamationCircleIcon } from '@heroicons/react/24/outline'
import { Button, Card } from './ui'
import type { Workflow, WorkflowStep } from '../types'

interface WorkflowViewerProps {
  workflow: Workflow | null
  isOpen: boolean
  onClose: () => void
  onExecuteStep?: (stepId: string, parameters?: Record<string, any>) => Promise<void>
  onExecuteWorkflow?: (workflow: Workflow) => Promise<void>
}

interface StepExecutionState {
  [stepId: string]: {
    status: 'pending' | 'running' | 'completed' | 'failed'
    result?: any
    error?: string
  }
}

export const WorkflowViewer: React.FC<WorkflowViewerProps> = ({
  workflow,
  isOpen,
  onClose,
  onExecuteStep,
  onExecuteWorkflow
}) => {
  const [stepStates, setStepStates] = useState<StepExecutionState>({})
  const [isExecutingWorkflow, setIsExecutingWorkflow] = useState(false)

  useEffect(() => {
    if (workflow) {
      // Initialize step states
      const initialStates: StepExecutionState = {}
      workflow.steps.forEach(step => {
        initialStates[step.id] = { status: 'pending' }
      })
      setStepStates(initialStates)
    }
  }, [workflow])

  const handleExecuteStep = async (step: WorkflowStep) => {
    if (!onExecuteStep) return

    setStepStates(prev => ({
      ...prev,
      [step.id]: { ...prev[step.id], status: 'running' }
    }))

    try {
      await onExecuteStep(step.id)
      setStepStates(prev => ({
        ...prev,
        [step.id]: { ...prev[step.id], status: 'completed' }
      }))
    } catch (error) {
      setStepStates(prev => ({
        ...prev,
        [step.id]: { 
          ...prev[step.id], 
          status: 'failed',
          error: error instanceof Error ? error.message : 'Execution failed'
        }
      }))
    }
  }

  const handleExecuteWorkflow = async () => {
    if (!workflow || !onExecuteWorkflow) return

    setIsExecutingWorkflow(true)
    try {
      await onExecuteWorkflow(workflow)
    } finally {
      setIsExecutingWorkflow(false)
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running':
        return <div className="animate-spin w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full" />
      case 'completed':
        return <CheckCircleIcon className="w-4 h-4 text-green-500" />
      case 'failed':
        return <ExclamationCircleIcon className="w-4 h-4 text-red-500" />
      default:
        return <div className="w-4 h-4 rounded-full border-2 border-gray-300" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'draft': return 'bg-yellow-100 text-yellow-800'
      case 'active': return 'bg-blue-100 text-blue-800'
      case 'completed': return 'bg-green-100 text-green-800'
      case 'error': return 'bg-red-100 text-red-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  if (!isOpen || !workflow) return null

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="fixed inset-0 bg-black bg-opacity-50" onClick={onClose} />
        
        <div className="relative bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-hidden">
          {/* Header */}
          <div className="flex items-center justify-between p-6 border-b">
            <div className="flex items-center space-x-4">
              <h2 className="text-xl font-semibold text-gray-900">{workflow.name}</h2>
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(workflow.status)}`}>
                {workflow.status}
              </span>
            </div>
            <div className="flex items-center space-x-2">
              <Button
                variant="primary"
                size="sm"
                onClick={handleExecuteWorkflow}
                disabled={isExecutingWorkflow}
                className="flex items-center space-x-2"
              >
                {isExecutingWorkflow ? (
                  <>
                    <div className="animate-spin w-4 h-4 border-2 border-white border-t-transparent rounded-full" />
                    <span>Executing...</span>
                  </>
                ) : (
                  <>
                    <PlayIcon className="w-4 h-4" />
                    <span>Execute Workflow</span>
                  </>
                )}
              </Button>
              <button
                onClick={onClose}
                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <XMarkIcon className="w-5 h-5" />
              </button>
            </div>
          </div>

          {/* Content */}
          <div className="p-6 overflow-y-auto max-h-[calc(90vh-120px)]">
            {/* Workflow Description */}
            <div className="mb-6">
              <h3 className="text-lg font-medium text-gray-900 mb-2">Description</h3>
              <p className="text-gray-600">{workflow.description}</p>
            </div>

            {/* User Parameters */}
            {workflow.user_parameters && workflow.user_parameters.length > 0 && (
              <div className="mb-6">
                <h3 className="text-lg font-medium text-gray-900 mb-3">Parameters</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {workflow.user_parameters.map((param, index) => (
                    <Card key={index} className="p-4">
                      <div className="flex justify-between items-start mb-2">
                        <h4 className="font-medium text-gray-900">{param.name}</h4>
                        {param.required && (
                          <span className="text-xs text-red-500 font-medium">Required</span>
                        )}
                      </div>
                      <p className="text-sm text-gray-600 mb-2">{param.description}</p>
                      <div className="text-xs text-gray-500">
                        Type: {param.type}
                      </div>
                    </Card>
                  ))}
                </div>
              </div>
            )}

            {/* Workflow Steps */}
            <div className="mb-6">
              <h3 className="text-lg font-medium text-gray-900 mb-3">
                Steps ({workflow.steps.length})
              </h3>
              <div className="space-y-4">
                {workflow.steps.map((step, index) => {
                  const stepState = stepStates[step.id] || { status: 'pending' }
                  
                  return (
                    <Card key={step.id} className="p-4">
                      <div className="flex items-start justify-between mb-3">
                        <div className="flex items-center space-x-3">
                          <div className="flex-shrink-0 w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center text-sm font-medium text-gray-600">
                            {index + 1}
                          </div>
                          <div>
                            <h4 className="font-medium text-gray-900">{step.name}</h4>
                            <p className="text-sm text-gray-600">Step {index + 1}</p>
                          </div>
                        </div>
                        <div className="flex items-center space-x-2">
                          {getStatusIcon(stepState.status)}
                          <Button
                            variant="secondary"
                            size="sm"
                            onClick={() => handleExecuteStep(step)}
                            disabled={stepState.status === 'running'}
                            className="flex items-center space-x-1"
                          >
                            <PlayIcon className="w-3 h-3" />
                            <span>Run</span>
                          </Button>
                        </div>
                      </div>

                      {/* Step Details */}
                      <div className="ml-9 space-y-2">
                        <div className="text-sm">
                          <span className="font-medium text-gray-700">Service:</span>{' '}
                          <span className="text-gray-600">{step.service}</span>
                        </div>
                        <div className="text-sm">
                          <span className="font-medium text-gray-700">Action:</span>{' '}
                          <span className="text-gray-600">{step.action}</span>
                        </div>
                        
                        {step.depends_on && step.depends_on.length > 0 && (
                          <div className="text-sm">
                            <span className="font-medium text-gray-700">Dependencies:</span>{' '}
                            <span className="text-gray-600">{step.depends_on.join(', ')}</span>
                          </div>
                        )}

                        {stepState.error && (
                          <div className="mt-2 p-2 bg-red-50 border border-red-200 rounded text-sm text-red-700">
                            <strong>Error:</strong> {stepState.error}
                          </div>
                        )}
                      </div>
                    </Card>
                  )
                })}
              </div>
            </div>

            {/* Service Bindings */}
            {workflow.service_bindings && workflow.service_bindings.length > 0 && (
              <div className="mb-6">
                <h3 className="text-lg font-medium text-gray-900 mb-3">Service Bindings</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {workflow.service_bindings.map((binding, index) => (
                    <Card key={index} className="p-4">
                      <h4 className="font-medium text-gray-900 mb-2">{binding.service}</h4>
                      <div className="text-sm text-gray-600">
                        {binding.endpoint && <div>Endpoint: {binding.endpoint}</div>}
                        {binding.oauth_scopes && (
                          <div className="mt-1">
                            Scopes: {binding.oauth_scopes.join(', ')}
                          </div>
                        )}
                      </div>
                    </Card>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
