import React, { useState, useEffect } from 'react'

interface ServiceProxyDashboardProps {
  providers: string[]
  isAuthenticated: boolean
}

const ServiceProxyDashboard: React.FC<ServiceProxyDashboardProps> = ({
  providers,
  isAuthenticated
}) => {
  const [selectedProvider, setSelectedProvider] = useState<string>('')
  const [services, setServices] = useState<string[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (selectedProvider) {
      loadServices(selectedProvider)
    }
  }, [selectedProvider])

  const loadServices = async (provider: string) => {
    setLoading(true)
    try {
      const response = await fetch(`/api/providers/${provider}/services`)
      if (response.ok) {
        const data = await response.json()
        setServices(data.services || [])
      }
    } catch (error) {
      console.error('Failed to load services:', error)
    } finally {
      setLoading(false)
    }
  }

  const getServiceIcon = (service: string) => {
    const icons: { [key: string]: string } = {
      'gmail': '📧',
      'docs': '📄',
      'drive': '💾',
      'calendar': '📅',
      'send_email': '📤',
      'read_emails': '📥',
      'search_emails': '🔍',
      'create_document': '📝',
      'edit_document': '✏️',
      'share_document': '🔗',
      'upload_file': '⬆️',
      'download_file': '⬇️',
      'list_files': '📋',
      'create_event': '🗓️',
      'list_events': '📅',
      'update_event': '✏️'
    }
    return icons[service] || '⚙️'
  }

  const getServiceDescription = (service: string) => {
    const descriptions: { [key: string]: string } = {
      'send_email': 'Send emails via Gmail API',
      'read_emails': 'Read emails from Gmail',
      'search_emails': 'Search emails in Gmail',
      'create_document': 'Create new Google Docs',
      'edit_document': 'Edit existing documents',
      'share_document': 'Share documents with others',
      'upload_file': 'Upload files to Google Drive',
      'download_file': 'Download files from Drive',
      'list_files': 'List files in Google Drive',
      'create_event': 'Create calendar events',
      'list_events': 'List calendar events',
      'update_event': 'Update calendar events'
    }
    return descriptions[service] || 'Service operation'
  }

  return (
    <div className="card">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-secondary-900">
          Service Providers
        </h2>
        <div className={`px-3 py-1 rounded-full text-xs font-medium ${
          isAuthenticated 
            ? 'bg-green-100 text-green-800' 
            : 'bg-yellow-100 text-yellow-800'
        }`}>
          {isAuthenticated ? 'Ready' : 'Auth Required'}
        </div>
      </div>

      {/* Provider Selection */}
      <div className="mb-6">
        <label className="block text-sm font-medium text-secondary-700 mb-2">
          Select Provider
        </label>
        <select
          value={selectedProvider}
          onChange={(e) => setSelectedProvider(e.target.value)}
          className="input-field"
          disabled={!isAuthenticated}
        >
          <option value="">Choose a provider...</option>
          {providers.map((provider) => (
            <option key={provider} value={provider}>
              {provider.charAt(0).toUpperCase() + provider.slice(1)}
            </option>
          ))}
        </select>
      </div>

      {/* Services List */}
      {selectedProvider && (
        <div>
          <h3 className="text-lg font-medium text-secondary-800 mb-4">
            Available Services
          </h3>
          
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-3">
              {services.map((service) => (
                <div
                  key={service}
                  className="flex items-center p-3 bg-secondary-50 rounded-lg border border-secondary-200 hover:bg-secondary-100 transition-colors"
                >
                  <span className="text-2xl mr-3">{getServiceIcon(service)}</span>
                  <div className="flex-1">
                    <div className="font-medium text-secondary-900">{service}</div>
                    <div className="text-sm text-secondary-600">
                      {getServiceDescription(service)}
                    </div>
                  </div>
                  <div className="px-2 py-1 bg-primary-100 text-primary-800 text-xs rounded">
                    Available
                  </div>
                </div>
              ))}
            </div>
          )}
          
          {!loading && services.length === 0 && (
            <div className="text-center py-8 text-secondary-500">
              No services available for this provider
            </div>
          )}
        </div>
      )}
      
      {!selectedProvider && (
        <div className="text-center py-8 text-secondary-500">
          Select a provider to view available services
        </div>
      )}
    </div>
  )
}

export default ServiceProxyDashboard
