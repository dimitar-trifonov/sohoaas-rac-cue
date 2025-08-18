import React, { useState, useEffect } from 'react'
import { XMarkIcon, PlayIcon, CheckCircleIcon, ExclamationCircleIcon } from '@heroicons/react/24/outline'
import { Button, Card } from './ui'
import type { WorkflowStep, WorkflowOrFile } from '../types'

interface WorkflowViewerProps {
  workflow: WorkflowOrFile | null
  isOpen: boolean
  onClose: () => void
  onExecuteStep?: (stepId: string, parameters?: Record<string, any>) => Promise<void>
  onExecuteWorkflow?: (workflow: WorkflowOrFile) => Promise<void>
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
  const [userParameters, setUserParameters] = useState<Record<string, any>>({})

  // Generate a unique key for this workflow's parameters
  const getParameterStorageKey = (workflowName: string) => {
    return `sohoaas_workflow_params_${workflowName.replace(/[^a-zA-Z0-9]/g, '_')}`
  }

  // Load saved parameters from localStorage
  const loadSavedParameters = (workflowName: string): Record<string, any> => {
    try {
      const storageKey = getParameterStorageKey(workflowName)
      const saved = localStorage.getItem(storageKey)
      return saved ? JSON.parse(saved) : {}
    } catch (error) {
      console.warn('Failed to load saved parameters:', error)
      return {}
    }
  }

  // Save parameters to localStorage
  const saveParameters = (workflowName: string, params: Record<string, any>) => {
    try {
      const storageKey = getParameterStorageKey(workflowName)
      localStorage.setItem(storageKey, JSON.stringify(params))
    } catch (error) {
      console.warn('Failed to save parameters:', error)
    }
  }

  // Clear saved parameters for this workflow
  const clearSavedParameters = () => {
    if (!workflow?.name) return
    
    try {
      const storageKey = getParameterStorageKey(workflow.name)
      localStorage.removeItem(storageKey)
      
      // Reset parameters to default values
      const defaultParams: Record<string, any> = {}
      if (workflow.user_parameters) {
        if (Array.isArray(workflow.user_parameters)) {
          workflow.user_parameters.forEach(param => {
            defaultParams[param.name] = param.default ?? (param.type === 'boolean' ? false : '')
          })
        } else {
          Object.entries(workflow.user_parameters).forEach(([paramName, paramDef]) => {
            if (typeof paramDef === 'object' && paramDef !== null) {
              const def = paramDef as any
              defaultParams[paramName] = def.default ?? (def.type === 'boolean' ? false : '')
            }
          })
        }
      }
      setUserParameters(defaultParams)
    } catch (error) {
      console.warn('Failed to clear saved parameters:', error)
    }
  }

  useEffect(() => {
    if (workflow) {
      // Initialize step states
      const initialStates: StepExecutionState = {}
      if (workflow.steps) {
        workflow.steps.forEach(step => {
          initialStates[step.id] = { status: 'pending' }
        })
      }
      setStepStates(initialStates)
      
      // Load saved parameters first, then merge with defaults
      const savedParams = loadSavedParameters(workflow.name || 'unnamed_workflow')
      const initialParams: Record<string, any> = {}
      
      // Check both direct user_parameters (legacy) and parsed_data.user_parameters (current)
      const userParams = workflow.user_parameters || ('parsed_data' in workflow ? workflow.parsed_data?.user_parameters : undefined)
      if (userParams) {
        // Handle both array format (legacy) and object format (CUE)
        if (Array.isArray(userParams)) {
          // Legacy array format
          userParams.forEach(param => {
            const defaultValue = param.default ?? (param.type === 'boolean' ? false : '')
            initialParams[param.name] = savedParams[param.name] ?? defaultValue
          })
        } else {
          // CUE object format: { paramName: { type, prompt, required, default } }
          Object.entries(userParams).forEach(([paramName, paramDef]) => {
            if (typeof paramDef === 'object' && paramDef !== null) {
              const def = paramDef as any
              let defaultValue = def.default ?? (def.type === 'boolean' ? false : '')
              
              // Keep datetime defaults as-is for input display (timezone conversion happens on execution)
              
              initialParams[paramName] = savedParams[paramName] ?? defaultValue
            }
          })
        }
      }
      setUserParameters(initialParams)
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

    // Validate required parameters
    let missingParams: string[] = []
    const userParams = workflow.user_parameters || ('parsed_data' in workflow ? workflow.parsed_data?.user_parameters : undefined)
    if (userParams) {
      if (Array.isArray(userParams)) {
        // Legacy array format
        missingParams = userParams
          .filter(param => param.required && !userParameters[param.name])
          .map(param => param.name)
      } else {
        // CUE object format: { paramName: { type, prompt, required } }
        missingParams = Object.entries(userParams)
          .filter(([paramName, paramDef]) => {
            const def = paramDef as any
            return def && def.required && !userParameters[paramName]
          })
          .map(([paramName]) => paramName)
      }
    }
    
    if (missingParams.length > 0) {
      alert(`Please fill in required parameters: ${missingParams.join(', ')}`)
      return
    }

    setIsExecutingWorkflow(true)
    try {
      // Pass user parameters to execution
      const workflowWithParams = { ...workflow, executionParameters: userParameters }
      await onExecuteWorkflow(workflowWithParams)
    } finally {
      setIsExecutingWorkflow(false)
    }
  }

  const handleParameterChange = (paramName: string, value: any) => {
    const updatedParams = {
      ...userParameters,
      [paramName]: value
    }
    setUserParameters(updatedParams)
  
    // Save parameters to localStorage whenever they change
    if (workflow?.name) {
      saveParameters(workflow.name, updatedParams)
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
                title={workflow.user_parameters && (Array.isArray(workflow.user_parameters) ? workflow.user_parameters.some(p => p.required) : Object.values(workflow.user_parameters).some((p: any) => p.required)) ? 'Fill in required parameters before executing' : 'Execute workflow'}
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
            {/* Check if we have parsed data available */}
            {'parsed_data' in workflow && workflow.parsed_data ? (
              /* Display parsed workflow data */
              <div>
                {/* Workflow Description */}
                <div className="mb-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-2">Description</h3>
                  <p className="text-gray-600">{workflow.parsed_data.description || workflow.description}</p>
                </div>

                {/* Original Intent */}
                {workflow.parsed_data.original_intent && (
                  <div className="mb-6">
                    <h3 className="text-lg font-medium text-gray-900 mb-2">Original Request</h3>
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                      <p className="text-blue-800 italic">"{workflow.parsed_data.original_intent}"</p>
                    </div>
                  </div>
                )}

                {/* User Parameters */}
                {(() => {
                  const userParams = workflow.user_parameters || ('parsed_data' in workflow ? workflow.parsed_data?.user_parameters : undefined)
                  return userParams && Object.keys(userParams).length > 0
                })() && (
                  <div className="mb-6">
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="text-lg font-medium text-gray-900">Parameters</h3>
                      <div className="flex items-center space-x-2">
                        <span className="text-xs text-green-600 bg-green-50 px-2 py-1 rounded-full">
                          Auto-saved
                        </span>
                        <button
                          onClick={clearSavedParameters}
                          className="text-xs text-gray-500 hover:text-red-600 underline"
                          title="Clear saved parameter values"
                        >
                          Clear
                        </button>
                      </div>
                    </div>
                    <div className="space-y-4">
                      {Object.entries(
                        workflow.user_parameters || ('parsed_data' in workflow ? workflow.parsed_data?.user_parameters : undefined) || {}
                      ).map(([paramName, paramConfig]: [string, any]) => (
                        <Card key={paramName} className="p-4">
                          <div className="flex items-start justify-between mb-2">
                            <label htmlFor={paramName} className="block text-sm font-medium text-gray-700">
                              {paramName}
                              {paramConfig.required && (
                                <span className="text-xs text-red-500 font-medium ml-1">Required</span>
                              )}
                            </label>
                          </div>
                          <p className="text-sm text-gray-600 mb-3">{paramConfig.prompt || paramConfig.description}</p>
                          
                          {/* Parameter Input Field */}
                          {paramConfig.type === 'string' && (
                            <input
                              id={paramName}
                              type="text"
                              value={userParameters[paramName] ?? paramConfig.default ?? ''}
                              onChange={(e) => handleParameterChange(paramName, e.target.value)}
                              placeholder={paramConfig.prompt || `Enter ${paramName}`}
                              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                              required={paramConfig.required}
                            />
                          )}
                          
                          {paramConfig.type === 'number' && (
                            <input
                              id={paramName}
                              type="number"
                              value={userParameters[paramName] ?? paramConfig.default ?? ''}
                              onChange={(e) => handleParameterChange(paramName, parseFloat(e.target.value))}
                              placeholder={paramConfig.prompt || `Enter ${paramName}`}
                              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                              required={paramConfig.required}
                            />
                          )}
                          
                          {paramConfig.type === 'boolean' && (
                            <label className="flex items-center">
                              <input
                                id={paramName}
                                type="checkbox"
                                checked={userParameters[paramName] ?? paramConfig.default ?? false}
                                onChange={(e) => handleParameterChange(paramName, e.target.checked)}
                                className="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
                              />
                              <span className="ml-2 text-sm text-gray-600">{paramConfig.prompt || `Enable ${paramName}`}</span>
                            </label>
                          )}
                          
                          {paramConfig.type === 'email' && (
                            <input
                              id={paramName}
                              type="email"
                              value={userParameters[paramName] ?? paramConfig.default ?? ''}
                              onChange={(e) => handleParameterChange(paramName, e.target.value)}
                              placeholder={paramConfig.prompt || `Enter ${paramName}`}
                              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                              required={paramConfig.required}
                            />
                          )}
                          
                          {paramConfig.type === 'datetime' && (
                            <input
                              id={paramName}
                              type="datetime-local"
                              value={(() => {
                                // Convert stored ISO string back to datetime-local format for display
                                const storedValue = userParameters[paramName] ?? paramConfig.default ?? '';
                                if (storedValue && typeof storedValue === 'string' && storedValue.includes('T')) {
                                  // Parse ISO string and convert to local datetime-local format
                                  try {
                                    const date = new Date(storedValue);
                                    if (!isNaN(date.getTime())) {
                                      // Format as YYYY-MM-DDTHH:MM for datetime-local input
                                      const year = date.getFullYear();
                                      const month = (date.getMonth() + 1).toString().padStart(2, '0');
                                      const day = date.getDate().toString().padStart(2, '0');
                                      const hours = date.getHours().toString().padStart(2, '0');
                                      const minutes = date.getMinutes().toString().padStart(2, '0');
                                      return `${year}-${month}-${day}T${hours}:${minutes}`;
                                    }
                                  } catch (e) {
                                    console.warn('Failed to parse datetime value:', storedValue);
                                  }
                                }
                                return storedValue;
                              })()}
                              onChange={(e) => {
                                // Send datetime without timezone - backend will handle timezone conversion
                                const localDateTime = e.target.value;
                                if (localDateTime) {
                                  // Convert to ISO format without timezone (backend will add timezone)
                                  const isoDateTime = localDateTime + ':00';
                                  handleParameterChange(paramName, isoDateTime);
                                } else {
                                  handleParameterChange(paramName, localDateTime);
                                }
                              }}
                              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                              required={paramConfig.required}
                            />
                          )}
                          
                          {paramConfig.validation && (
                            <p className="text-xs text-gray-500 mt-1">Validation: {paramConfig.validation}</p>
                          )}
                        </Card>
                      ))}
                    </div>
                  </div>
                )}

                {/* Workflow Steps */}
                <div className="mb-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-3">
                    Steps ({workflow.parsed_data.steps?.length || 0})
                  </h3>
                  <div className="space-y-4">
                    {(workflow.parsed_data.steps || []).map((step: any, index: number) => {
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
                              <span className="font-medium text-gray-700">Action:</span>{' '}
                              <span className="text-gray-600">{step.action}</span>
                            </div>
                            
                            {step.parameters && Object.keys(step.parameters).length > 0 && (
                              <div className="text-sm">
                                <span className="font-medium text-gray-700">Parameters:</span>
                                <div className="mt-1 bg-gray-50 rounded p-2">
                                  <pre className="text-xs text-gray-600 whitespace-pre-wrap">
                                    {JSON.stringify(step.parameters, null, 2)}
                                  </pre>
                                </div>
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
                {workflow.parsed_data.service_bindings && Object.keys(workflow.parsed_data.service_bindings).length > 0 && (
                  <div className="mb-6">
                    <h3 className="text-lg font-medium text-gray-900 mb-3">Service Bindings</h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {Object.entries(workflow.parsed_data.service_bindings).map(([service, config]: [string, any], index) => (
                        <Card key={index} className="p-4">
                          <h4 className="font-medium text-gray-900 mb-2">{service}</h4>
                          <div className="text-sm text-gray-600">
                            {config.endpoint && <div>Endpoint: {config.endpoint}</div>}
                            {config.oauth_scopes && (
                              <div className="mt-1">
                                Scopes: {Array.isArray(config.oauth_scopes) ? config.oauth_scopes.join(', ') : config.oauth_scopes}
                              </div>
                            )}
                          </div>
                        </Card>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ) : 'content' in workflow && !workflow.steps ? (
              /* WorkflowFile Display */
              <div className="space-y-6">
                <div className="mb-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-2">Workflow File</h3>
                  <p className="text-gray-600 mb-4">{workflow.description}</p>
                  <div className="bg-gray-50 rounded-lg p-4">
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div><span className="font-medium">Status:</span> {workflow.status}</div>
                      <div><span className="font-medium">Created:</span> {new Date(workflow.created_at).toLocaleString()}</div>
                      <div><span className="font-medium">File:</span> {'filename' in workflow ? workflow.filename : 'N/A'}</div>
                      <div><span className="font-medium">User ID:</span> {'user_id' in workflow ? workflow.user_id : 'N/A'}</div>
                    </div>
                  </div>
                </div>
                
                <div className="mb-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-3">CUE Workflow Content</h3>
                  <pre className="bg-gray-900 text-green-400 p-4 rounded-lg overflow-x-auto text-sm font-mono">
                    {'content' in workflow ? workflow.content : 'No content available'}
                  </pre>
                </div>
                
                <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                  <div className="flex">
                    <ExclamationCircleIcon className="w-5 h-5 text-yellow-400 mt-0.5 mr-3 flex-shrink-0" />
                    <div>
                      <h4 className="text-sm font-medium text-yellow-800">Workflow File Display</h4>
                      <p className="text-sm text-yellow-700 mt-1">
                        This workflow is stored as a CUE file. To see parsed steps and parameters, 
                        the backend needs to parse the CUE content into a structured workflow object.
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              /* Parsed Workflow Display */
              <div>
                {/* Workflow Description */}
                <div className="mb-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-2">Description</h3>
                  <p className="text-gray-600">{workflow.description}</p>
                </div>

            {/* User Parameters */}
            {workflow.user_parameters && (
              (Array.isArray(workflow.user_parameters) ? workflow.user_parameters.length > 0 : Object.keys(workflow.user_parameters).length > 0)
            ) && (
              <div className="mb-6">
                <h3 className="text-lg font-medium text-gray-900 mb-3">Parameters</h3>
                <div className="grid grid-cols-1 gap-4">
                  {Array.isArray(workflow.user_parameters) ? (
                    // Legacy array format
                    workflow.user_parameters.map((param, index) => (
                      <Card key={index} className="p-4">
                        <div className="flex justify-between items-start mb-2">
                          <label className="font-medium text-gray-900" htmlFor={param.name}>
                            {param.name}
                          </label>
                          {param.required && (
                            <span className="text-xs text-red-500 font-medium">Required</span>
                          )}
                        </div>
                        <p className="text-sm text-gray-600 mb-3">{param.description || param.prompt}</p>
                        
                        {/* Parameter Input Field */}
                        {param.type === 'string' && (
                          <input
                            id={param.name}
                            type="text"
                            value={userParameters[param.name] ?? param.default ?? ''}
                            onChange={(e) => handleParameterChange(param.name, e.target.value)}
                            placeholder={param.prompt || `Enter ${param.name}`}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            required={param.required}
                          />
                        )}
                        
                        {param.type === 'number' && (
                          <input
                            id={param.name}
                            type="number"
                            value={userParameters[param.name] ?? param.default ?? ''}
                            onChange={(e) => handleParameterChange(param.name, parseFloat(e.target.value))}
                            placeholder={param.prompt || `Enter ${param.name}`}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            required={param.required}
                          />
                        )}
                        
                        {param.type === 'boolean' && (
                          <div className="flex items-center">
                            <input
                              id={param.name}
                              type="checkbox"
                              checked={userParameters[param.name] || false}
                              onChange={(e) => handleParameterChange(param.name, e.target.checked)}
                              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                            />
                            <label htmlFor={param.name} className="ml-2 text-sm text-gray-700">
                              {param.prompt || `Enable ${param.name}`}
                            </label>
                          </div>
                        )}
                      </Card>
                    ))
                  ) : (
                    // CUE object format: { paramName: { type, prompt, required } }
                    Object.entries(workflow.user_parameters).map(([paramName, paramConfig]: [string, any]) => (
                      <Card key={paramName} className="p-4">
                        <div className="flex justify-between items-start mb-2">
                          <label className="font-medium text-gray-900" htmlFor={paramName}>
                            {paramName}
                          </label>
                          {paramConfig.required && (
                            <span className="text-xs text-red-500 font-medium">Required</span>
                          )}
                        </div>
                        <p className="text-sm text-gray-600 mb-3">{paramConfig.description || paramConfig.prompt}</p>
                        
                        {/* Parameter Input Field */}
                        {paramConfig.type === 'string' && (
                          <input
                            id={paramName}
                            type="text"
                            value={userParameters[paramName] ?? paramConfig.default ?? ''}
                            onChange={(e) => handleParameterChange(paramName, e.target.value)}
                            placeholder={paramConfig.prompt || `Enter ${paramName}`}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            required={paramConfig.required}
                          />
                        )}
                        
                        {paramConfig.type === 'number' && (
                          <input
                            id={paramName}
                            type="number"
                            value={userParameters[paramName] ?? paramConfig.default ?? ''}
                            onChange={(e) => handleParameterChange(paramName, parseFloat(e.target.value))}
                            placeholder={paramConfig.prompt || `Enter ${paramName}`}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            required={paramConfig.required}
                          />
                        )}
                        
                        {paramConfig.type === 'boolean' && (
                          <div className="flex items-center">
                            <input
                              id={paramName}
                              type="checkbox"
                              checked={userParameters[paramName] || false}
                              onChange={(e) => handleParameterChange(paramName, e.target.checked)}
                              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                            />
                            <label htmlFor={paramName} className="ml-2 text-sm text-gray-700">
                              {paramConfig.prompt || `Enable ${paramName}`}
                            </label>
                          </div>
                        )}
                      </Card>
                    ))
                  )}
                </div>
              </div>
            )}

            {/* Workflow Steps */}
            <div className="mb-6">
              <h3 className="text-lg font-medium text-gray-900 mb-3">
                Steps ({workflow.steps?.length || 0})
              </h3>
              <div className="space-y-4">
                {(workflow.steps || []).map((step, index) => {
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
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
