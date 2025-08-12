import React, { useState } from 'react'
import { PlayIcon, EyeIcon, ClockIcon, CheckCircleIcon, ExclamationCircleIcon } from '@heroicons/react/24/outline'
import type { Workflow } from '../types'
import { WorkflowViewer } from './WorkflowViewer'
import { Button } from './ui'
import { cn } from '../design-system'

interface WorkflowListProps {
  workflows: Workflow[]
  onExecuteWorkflow?: (workflow: Workflow) => Promise<void>
  onViewWorkflow?: (workflow: Workflow) => void
}

export const WorkflowList: React.FC<WorkflowListProps> = ({ 
  workflows, 
  onExecuteWorkflow,
  onViewWorkflow 
}) => {
  const [executingWorkflow, setExecutingWorkflow] = useState<string | null>(null)
  const [viewerOpen, setViewerOpen] = useState(false)
  const [selectedWorkflow, setSelectedWorkflow] = useState<Workflow | null>(null)
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircleIcon className="h-5 w-5 text-green-500" />
      case 'error':
        return <ExclamationCircleIcon className="h-5 w-5 text-red-500" />
      case 'active':
        return <ClockIcon className="h-5 w-5 text-blue-500 animate-spin" />
      default:
        return <ClockIcon className="h-5 w-5 text-gray-400" />
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'completed':
        return 'Completed'
      case 'error':
        return 'Error'
      case 'active':
      case 'running':
        return 'Running'
      default:
        return 'Draft'
    }
  }

  if (workflows.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Your Workflows</h2>
        <div className="text-center py-8">
          <ClockIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
          <p className="text-gray-500">No workflows created yet</p>
          <p className="text-sm text-gray-400 mt-1">Create your first automation workflow above</p>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Your Workflows</h2>
      <div className="space-y-3">
        {workflows.map((workflow) => {
          const isExecuting = executingWorkflow === workflow.id
          const canExecute = workflow.status === 'draft' || workflow.status === 'completed' || workflow.status === 'active'
          
          const handleExecute = async () => {
            if (!onExecuteWorkflow || !canExecute) return
            
            setExecutingWorkflow(workflow.id)
            try {
              await onExecuteWorkflow(workflow)
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
          
          return (
            <div key={workflow.id} className={cn(
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
                  {getStatusIcon(workflow.status)}
                  <span className="text-sm font-medium text-gray-700">
                    {getStatusText(workflow.status)}
                  </span>
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
        onExecuteWorkflow={onExecuteWorkflow}
      />
    </div>
  )
}
