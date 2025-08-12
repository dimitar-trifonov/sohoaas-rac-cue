import { useState, useEffect } from 'react'
import ServiceProxyDashboard from './components/ServiceProxyDashboard'
import AuthStatus from './components/AuthStatus'
import ProxyTester from './components/ProxyTester'
import MCPClient from './components/MCPClient'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [authUrl, setAuthUrl] = useState('')
  const [providers, setProviders] = useState<string[]>([])
  const [activeTab, setActiveTab] = useState<'rest' | 'mcp'>('rest')

  useEffect(() => {
    // Handle OAuth callback parameters
    const urlParams = new URLSearchParams(window.location.search)
    const authSuccess = urlParams.get('auth_success')
    const authError = urlParams.get('auth_error')
    
    if (authSuccess) {
      // Clear URL parameters and check auth status
      window.history.replaceState({}, document.title, window.location.pathname)
      setTimeout(() => {
        checkAuthStatus()
      }, 500)
    } else if (authError) {
      // Handle OAuth error
      console.error('OAuth error:', authError)
      alert(`Authentication failed: ${authError}`)
      // Clear URL parameters
      window.history.replaceState({}, document.title, window.location.pathname)
    }
    
    // Check authentication status and load providers
    checkAuthStatus()
    loadProviders()
  }, [])

  const checkAuthStatus = async () => {
    try {
      const response = await fetch('/api/auth/token')
      if (response.ok) {
        const data = await response.json()
        setIsAuthenticated(!!data.access_token)
      }
    } catch (error) {
      console.error('Failed to check auth status:', error)
    }
  }

  const loadProviders = async () => {
    try {
      const response = await fetch('/api/providers')
      if (response.ok) {
        const data = await response.json()
        setProviders(data.providers || [])
      }
    } catch (error) {
      console.error('Failed to load providers:', error)
    }
  }

  const getAuthUrl = async () => {
    try {
      const response = await fetch('/api/auth/login')
      if (response.ok) {
        const data = await response.json()
        setAuthUrl(data.auth_url)
      }
    } catch (error) {
      console.error('Failed to get auth URL:', error)
    }
  }

  return (
    <div className="min-h-screen bg-secondary-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-secondary-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div className="flex items-center">
              <h1 className="text-2xl font-bold text-secondary-900">
                RAC Service Proxies
              </h1>
              <span className="ml-3 px-2 py-1 text-xs font-medium bg-primary-100 text-primary-800 rounded-full">
                REST + MCP Dashboard
              </span>
            </div>
            <AuthStatus 
              isAuthenticated={isAuthenticated}
              authUrl={authUrl}
              onGetAuthUrl={getAuthUrl}
              onAuthChange={checkAuthStatus}
            />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Tab Navigation */}
        <div className="mb-8">
          <div className="border-b border-secondary-200">
            <nav className="-mb-px flex space-x-8">
              <button
                onClick={() => setActiveTab('rest')}
                className={`py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === 'rest'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-secondary-500 hover:text-secondary-700 hover:border-secondary-300'
                }`}
              >
                REST API Interface
              </button>
              <button
                onClick={() => setActiveTab('mcp')}
                className={`py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === 'mcp'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-secondary-500 hover:text-secondary-700 hover:border-secondary-300'
                }`}
              >
                MCP Protocol Interface
              </button>
            </nav>
          </div>
        </div>

        {/* Tab Content */}
        {activeTab === 'rest' ? (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            {/* Service Proxy Dashboard */}
            <div className="lg:col-span-1">
              <ServiceProxyDashboard 
                providers={providers}
                isAuthenticated={isAuthenticated}
              />
            </div>

            {/* Proxy Tester */}
            <div className="lg:col-span-1">
              <ProxyTester 
                providers={providers}
                isAuthenticated={isAuthenticated}
              />
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-8">
            {/* MCP Client */}
            <div>
              <MCPClient 
                isAuthenticated={isAuthenticated}
              />
            </div>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-secondary-200 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="text-center text-sm text-secondary-600">
            RAC Service Proxies - REST API & MCP Protocol Support
          </div>
        </div>
      </footer>
    </div>
  )
}

export default App
