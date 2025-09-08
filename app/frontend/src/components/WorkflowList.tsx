import React, { useEffect, useState } from 'react'
import { PlayIcon, EyeIcon, TrashIcon } from '@heroicons/react/24/outline'
import type { WorkflowOrFile } from '../types'
import { WorkflowViewer } from './WorkflowViewer'
import { Button } from './ui'
import { cn } from '../design-system'
import { AlertModal } from './AlertModal'
import { sohoaasApi } from '../services/api'
import { useWorkflowStore } from '../stores'

interface WorkflowListProps {
  workflows: WorkflowOrFile[]
  onExecuteWorkflow?: (workflow: WorkflowOrFile) => Promise<void>
  onViewWorkflow?: (workflow: WorkflowOrFile) => void
}

export const WorkflowList: React.FC<WorkflowListProps> = ({ 
  workflows, 
  onExecuteWorkflow,
  onViewWorkflow 
}) => {
  const [executingWorkflow, setExecutingWorkflow] = useState<string | null>(null)
  const [viewerOpen, setViewerOpen] = useState(false)
  const [selectedWorkflow, setSelectedWorkflow] = useState<WorkflowOrFile | null>(null)
  const [showAlert, setShowAlert] = useState(false)
  const [alertMessage, setAlertMessage] = useState('')
  const [alertVariant, setAlertVariant] = useState<'success' | 'error' | 'info' | 'warning'>('info')
  // bump to trigger re-render when production flag changes elsewhere
  const [prodVersion, setProdVersion] = useState(0)
  // local copy to allow immediate UI updates after deletion
  const [items, setItems] = useState<WorkflowOrFile[]>(workflows)

  const getParameterStorageKey = (workflowId: string) => `sohoaas_workflow_params_${workflowId}`
  const getProductionFlagKey = (workflowId: string) => `sohoaas_workflow_prod_${workflowId}`
  const isProductionReady = (workflowId: string | undefined) => {
    if (!workflowId) return false
    try {
      return localStorage.getItem(getProductionFlagKey(workflowId)) === 'true'
    } catch {
      return false
    }
  }
  const loadSavedParameters = (workflowId: string | undefined): Record<string, any> => {
    if (!workflowId) return {}
    try {
      const raw = localStorage.getItem(getParameterStorageKey(workflowId))
      return raw ? JSON.parse(raw) : {}
    } catch {
      return {}
    }
  }

  // Listen for production mode updates coming from WorkflowViewer
  useEffect(() => {
    const handler = (_e: Event) => setProdVersion(v => v + 1)
    window.addEventListener('sohoaas:production-updated', handler as EventListener)
    return () => {
      window.removeEventListener('sohoaas:production-updated', handler as EventListener)
    }
  }, [])

  // Sync local items when parent workflows prop changes
  useEffect(() => {
    setItems(workflows)
  }, [workflows])

  const handleExecuteStep = async (stepId: string, _parameters?: Record<string, any>) => {
    // TODO: Implement individual step execution API
    setAlertVariant('info')
    setAlertMessage(`Step execution not yet implemented. Step ID: ${stepId}`)
    setShowAlert(true)
  }

  const handleExecuteWorkflow = async (workflow: WorkflowOrFile) => {
    if (!onExecuteWorkflow) return
    
    setExecutingWorkflow(workflow.id)
    try {
      await onExecuteWorkflow(workflow)
    } finally {
      setExecutingWorkflow(null)
    }
  }

  if (items.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Your Workflows</h2>
        <div className="text-center py-8">
          <p className="text-gray-500 mb-1">No workflows created yet</p>
          <p className="text-sm text-gray-400 mt-1">Create your first automation workflow above</p>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Your Workflows</h2>
      <div className="space-y-3">
        {items.map((workflow) => {
          const isExecuting = executingWorkflow === workflow.id
          const prodReady = isProductionReady(workflow.id)
          const canExecute = prodReady
          
          const handleExecute = async () => {
            if (!onExecuteWorkflow || !canExecute) return
            
            setExecutingWorkflow(workflow.id)
            try {
              const params = loadSavedParameters(workflow.id)
              await onExecuteWorkflow({ ...workflow, executionParameters: params } as any)
            } finally {
              setExecutingWorkflow(null)
            }
          }
          
          const handleView = () => {
            setSelectedWorkflow(workflow)
            setViewerOpen(true)
            if (onViewWorkflow) {
              onViewWorkflow(workflow)
            }
          }

          const handleDelete = async () => {
            if (!workflow.id) return
            const confirmMsg = `Delete this workflow? This will remove it from your backend storage. Artifacts outside the workflow folder remain.`
            if (!window.confirm(confirmMsg)) return
            const ok = await sohoaasApi.deleteWorkflow(workflow.id)
            if (ok) {
              try {
                localStorage.removeItem(getParameterStorageKey(workflow.id))
                localStorage.removeItem(getProductionFlagKey(workflow.id))
              } catch {}
              setItems(prev => prev.filter(w => w.id !== workflow.id))
              // Refresh global store so tab counter and parent state update
              try {
                const { loadWorkflows } = useWorkflowStore.getState()
                await loadWorkflows()
              } catch (e) {
                console.warn('Failed to refresh workflows after delete:', e)
              }
              setAlertVariant('success')
              setAlertMessage('Workflow deleted')
              setShowAlert(true)
            } else {
              setAlertVariant('error')
              setAlertMessage('Failed to delete workflow')
              setShowAlert(true)
            }
          }
          
          return (
            <div key={`${workflow.id}-${prodVersion}`} className={cn(
              "border border-gray-200 rounded-lg p-4 transition-all duration-200",
              "hover:border-primary-300 hover:shadow-md"
            )}>
              <div className="flex items-start justify-between">
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-900 mb-1 truncate">
                    {workflow.user_message}
                  </p>
                  <p className="text-xs text-gray-500 mb-3">
                    Created: {new Date(workflow.created_at).toLocaleString()}
                  </p>
                  
                  {/* Action Buttons */}
                  <div className="flex items-center space-x-2">
                    <Button
                      size="sm"
                      variant={canExecute ? "primary" : "ghost"}
                      onClick={handleExecute}
                      disabled={!canExecute || isExecuting}
                      className="flex items-center space-x-1"
                      title={prodReady ? 'Run workflow' : 'Run disabled. Open the workflow and click Save after filling required parameters to enable production mode.'}
                    >
                      <PlayIcon className="w-4 h-4" />
                      <span>{isExecuting ? 'Running...' : 'Run'}</span>
                    </Button>
                    
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={handleView}
                      className="flex items-center space-x-1"
                    >
                      <EyeIcon className="w-4 h-4" />
                      <span>View</span>
                    </Button>
                  </div>
                </div>
                
                <div className="flex items-center space-x-2 ml-4 flex-shrink-0">
                  <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${prodReady ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}`}>
                    {prodReady ? 'production' : 'draft'}
                  </span>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={handleDelete}
                    className="flex items-center space-x-1 border-red-200 text-red-600 hover:bg-red-50"
                    title="Delete workflow"
                  >
                    <TrashIcon className="w-4 h-4" />
                    <span>Delete</span>
                  </Button>
                </div>
              </div>
            </div>
          )
        })}
      </div>
      
      <WorkflowViewer
        workflow={selectedWorkflow}
        isOpen={viewerOpen}
        onClose={() => {
          setViewerOpen(false)
          setSelectedWorkflow(null)
        }}
        onExecuteStep={handleExecuteStep}
        onExecuteWorkflow={handleExecuteWorkflow}
      />
      <AlertModal
        isOpen={showAlert}
        message={alertMessage}
        variant={alertVariant}
        onClose={() => setShowAlert(false)}
      />
    </div>
  )
}
