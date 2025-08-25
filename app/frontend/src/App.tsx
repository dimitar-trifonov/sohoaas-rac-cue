import { useEffect, useState } from 'react'
import { useAuthStore, useWorkflowStore, useUIStore, initializeStores } from './stores'
import { AuthStatus, WorkflowCreator, WorkflowList } from './components'
import { Header, Navigation, Container } from './components/layout'
import { Button, Card } from './components/ui'
import { UnauthorizedModal } from './components/UnauthorizedModal'
import { AlertModal } from './components/AlertModal'
import { cn, typographyVariants } from './design-system'


function App() {
  // External state management - outside React rendering cycle
  const { isAuthenticated, checkAuth, error, clearError } = useAuthStore()
  const { workflows, loadWorkflows } = useWorkflowStore()
  const { activeTab, setActiveTab } = useUIStore()
  const [showUnauthorizedModal, setShowUnauthorizedModal] = useState(false)
  const [showAlertModal, setShowAlertModal] = useState(false)
  const [alertMessage, setAlertMessage] = useState('')
  const [alertVariant, setAlertVariant] = useState<'success' | 'error' | 'info' | 'warning'>('info')

  useEffect(() => {
    // Initialize all stores on app start
    initializeStores()
  }, [])

  // Handle unauthorized access error
  useEffect(() => {
    if (error === 'UNAUTHORIZED_ACCESS') {
      setShowUnauthorizedModal(true)
    }
  }, [error])

  const handleCloseUnauthorizedModal = () => {
    setShowUnauthorizedModal(false)
    clearError()
  }

  const checkAuthStatus = () => {
    checkAuth()
  }

  const handleWorkflowCreated = (workflow: any) => {
    // Handle workflow creation
    console.log('Workflow created:', workflow)
    loadWorkflows()
  }

  const handleExecuteWorkflow = async (workflow: any) => {
    try {
      console.log('Executing workflow:', workflow)
      
      // Extract user parameters from workflow object (passed by WorkflowViewer)
      const userParameters = workflow.executionParameters || {}
      console.log('User parameters for execution:', userParameters)
      
      // Execute via workflow store/API (handles proxy and auth)
      const result = await useWorkflowStore.getState().executeWorkflow(
        workflow.id || workflow.ID,
        userParameters
      )
      if (!result) {
        throw new Error('Execution failed')
      }
      console.log('Workflow execution result:', result)
      
      // Refresh workflows to show updated status
      loadWorkflows()
      
      // Show success message
      setAlertVariant('success')
      setAlertMessage('Workflow executed successfully!')
      setShowAlertModal(true)
    } catch (error) {
      console.error('Failed to execute workflow:', error)
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred'
      setAlertVariant('error')
      setAlertMessage(`Failed to execute workflow: ${errorMessage}`)
      setShowAlertModal(true)
    }
  }

  const handleViewWorkflow = (workflow: any) => {
    console.log('Viewing workflow:', workflow)
    // WorkflowList component handles the view functionality internally
    // No additional action needed here
  }

  const handleLogin = async () => {
    // Use Firebase Auth for authentication
    const { login } = useAuthStore.getState()
    await login()
  }

  // Demo mode detection (for large screen presentations)
  const isDemoMode = window.innerWidth >= 1920

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Responsive Header */}
      <Header demoMode={isDemoMode}>
        <AuthStatus 
          isAuthenticated={isAuthenticated} 
          onAuthChange={checkAuthStatus}
        />
      </Header>

      {/* Responsive Navigation */}
      <Navigation
        activeTab={activeTab}
        onTabChange={setActiveTab}
        workflowCount={workflows.length}
        demoMode={isDemoMode}
      />

      {/* Responsive Main Content */}
      <main className={cn(
        'py-6 lg:py-8 xl:py-12',
        isDemoMode && 'py-16'
      )}>
        <Container size={isDemoMode ? 'demo' : 'xl'}>
          {!isAuthenticated ? (
            <div className="text-center py-12 lg:py-16 xl:py-24">
              <Card 
                padding={isDemoMode ? 'xl' : 'lg'} 
                className="max-w-2xl mx-auto"
              >
                <h2 className={cn(
                  isDemoMode 
                    ? typographyVariants.demo.title
                    : 'text-xl lg:text-2xl xl:text-3xl font-semibold text-gray-900 mb-4 lg:mb-6'
                )}>
                  Authentication Required
                </h2>
                <p className={cn(
                  isDemoMode
                    ? typographyVariants.demo.body
                    : 'text-gray-600 mb-6 lg:mb-8 text-base lg:text-lg'
                )}>
                  Please authenticate with Google Workspace to create and manage your automation workflows.
                </p>
                <Button
                  onClick={handleLogin}
                  size={isDemoMode ? 'xl' : 'lg'}
                  className="w-full"
                >
                  Connect Google Workspace
                </Button>
              </Card>
            </div>
          ) : (
            <div className={cn(
              'space-y-6 lg:space-y-8',
              isDemoMode && 'space-y-12'
            )}>
              {activeTab === 'create' && (
                <WorkflowCreator onWorkflowCreated={handleWorkflowCreated} />
              )}
              {activeTab === 'workflows' && (
                <WorkflowList 
                  workflows={workflows}
                  onExecuteWorkflow={handleExecuteWorkflow}
                  onViewWorkflow={handleViewWorkflow}
                />
              )}
            </div>
          )}
        </Container>
      </main>
      
      {/* Unauthorized Access Modal */}
      <UnauthorizedModal 
        isOpen={showUnauthorizedModal}
        onClose={handleCloseUnauthorizedModal}
      />
      {/* Global Alert Modal */}
      <AlertModal
        isOpen={showAlertModal}
        message={alertMessage}
        variant={alertVariant}
        onClose={() => setShowAlertModal(false)}
      />
    </div>
  )
}

export default App
