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
  
  // Initialize authentication
  await useAuthStore.getState().checkAuth()
  
  // Load services if authenticated
  const { isAuthenticated } = useAuthStore.getState()
  if (isAuthenticated) {
    await useServiceStore.getState().refreshServices()
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
