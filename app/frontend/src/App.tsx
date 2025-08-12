import { useEffect } from 'react'
import { useAuthStore, useWorkflowStore, useUIStore, initializeStores } from './stores'
import { AuthStatus, WorkflowCreator, WorkflowList } from './components'
import { Header, Navigation, Container } from './components/layout'
import { Button, Card } from './components/ui'
import { cn, typographyVariants } from './design-system'

function App() {
  // External state management - outside React rendering cycle
  const { isAuthenticated, checkAuth } = useAuthStore()
  const { workflows, loadWorkflows } = useWorkflowStore()
  const { activeTab, setActiveTab } = useUIStore()

  useEffect(() => {
    // Initialize all stores on app start
    initializeStores()
  }, [])

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
      
      // Call backend execution API
      const response = await fetch('http://localhost:8081/api/v1/workflow/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('oauth_token')}`,
        },
        body: JSON.stringify({
          workflow_cue: workflow.workflow_cue || workflow.content,
          user_parameters: {} // Add user parameters if needed
        })
      })
      
      if (!response.ok) {
        throw new Error(`Execution failed: ${response.statusText}`)
      }
      
      const result = await response.json()
      console.log('Workflow execution result:', result)
      
      // Refresh workflows to show updated status
      loadWorkflows()
      
      // Show success message
      alert('Workflow executed successfully!')
    } catch (error) {
      console.error('Failed to execute workflow:', error)
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred'
      alert(`Failed to execute workflow: ${errorMessage}`)
    }
  }

  const handleViewWorkflow = (workflow: any) => {
    console.log('Viewing workflow:', workflow)
    // TODO: Implement workflow details modal or page
    alert(`Workflow Details:\n\nMessage: ${workflow.user_message}\nStatus: ${workflow.status}\nCreated: ${new Date(workflow.created_at).toLocaleString()}`)
  }

  const handleLogin = async () => {
    try {
      // Get the OAuth URL from the MCP service via nginx proxy
      const response = await fetch('http://localhost:3000/api/auth/login')
      const data = await response.json()
      
      if (data.auth_url) {
        // Open the Google OAuth URL in a new window
        window.open(data.auth_url, '_blank')
      } else {
        console.error('No auth_url received:', data)
      }
    } catch (error) {
      console.error('Failed to get OAuth URL:', error)
    }
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
    </div>
  )
}

export default App
