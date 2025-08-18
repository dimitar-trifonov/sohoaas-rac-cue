// SOHOAAS Store Index
// Centralized export for all Zustand stores
// External state management outside React rendering cycle

export { useAuthStore } from './authStore'
export { useWorkflowStore } from './workflowStore'
export { useServiceStore } from './serviceStore'
export { useUIStore } from './uiStore'

// Store initialization helper
export const initializeStores = async () => {
  const { useAuthStore } = await import('./authStore')
  const { useServiceStore } = await import('./serviceStore')
  const { useWorkflowStore } = await import('./workflowStore')
  
  // Initialize authentication
  await useAuthStore.getState().checkAuth()
  
  // Load services and workflows if authenticated
  const { isAuthenticated } = useAuthStore.getState()
  if (isAuthenticated) {
    await useServiceStore.getState().refreshServices()
    await useWorkflowStore.getState().loadWorkflows()
  }
}

// Store reset helper for logout
export const resetStores = () => {
  import('./workflowStore').then(({ useWorkflowStore }) => {
    useWorkflowStore.getState().reset()
  })
  import('./serviceStore').then(({ useServiceStore }) => {
    useServiceStore.getState().clearError()
  })
  import('./uiStore').then(({ useUIStore }) => {
    useUIStore.getState().clearNotifications()
  })
}
