import React, { useState, useEffect } from 'react'

interface FunctionMetadata {
  name: string
  display_name: string
  description: string
  example_payload: any
  required_fields: string[]
}

interface ServiceMetadata {
  display_name: string
  description: string
  functions: { [key: string]: FunctionMetadata }
}

interface ProviderMetadata {
  display_name: string
  description: string
  services: { [key: string]: ServiceMetadata }
}

interface ProxyTesterProps {
  providers: string[]
  isAuthenticated: boolean
}

const ProxyTester: React.FC<ProxyTesterProps> = ({
  providers,
  isAuthenticated
}) => {
  const [selectedProvider, setSelectedProvider] = useState('')
  const [selectedService, setSelectedService] = useState('')
  const [selectedFunction, setSelectedFunction] = useState('')
  const [payload, setPayload] = useState('')
  const [result, setResult] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [serviceMetadata, setServiceMetadata] = useState<{ [key: string]: ProviderMetadata }>({})
  const [availableServices, setAvailableServices] = useState<string[]>([])
  const [availableFunctions, setAvailableFunctions] = useState<FunctionMetadata[]>([])

  // Load service metadata from backend
  useEffect(() => {
    const loadServiceMetadata = async () => {
      try {
        const response = await fetch('/api/services')
        if (response.ok) {
          const data = await response.json()
          setServiceMetadata(data.providers || {})
        }
      } catch (error) {
        console.error('Failed to load service metadata:', error)
      }
    }

    loadServiceMetadata()
  }, [])

  // Update available services when provider changes
  useEffect(() => {
    if (selectedProvider && serviceMetadata[selectedProvider]?.services) {
      const services = Object.keys(serviceMetadata[selectedProvider].services || {})
      setAvailableServices(services)
      setSelectedService('')
      setSelectedFunction('')
      setAvailableFunctions([])
    } else {
      setAvailableServices([])
      setSelectedService('')
      setSelectedFunction('')
      setAvailableFunctions([])
    }
  }, [selectedProvider, serviceMetadata])

  // Update available functions when service changes
  useEffect(() => {
    if (selectedProvider && selectedService && serviceMetadata[selectedProvider]?.services?.[selectedService]?.functions) {
      const functions = Object.values(serviceMetadata[selectedProvider].services[selectedService].functions || {})
      setAvailableFunctions(functions)
      setSelectedFunction('')
    } else {
      setAvailableFunctions([])
      setSelectedFunction('')
    }
  }, [selectedService, selectedProvider, serviceMetadata])

  // Update payload when function changes
  useEffect(() => {
    if (selectedFunction && selectedProvider && selectedService) {
      const functionMeta = serviceMetadata[selectedProvider]?.services[selectedService]?.functions[selectedFunction]
      if (functionMeta) {
        setPayload(JSON.stringify(functionMeta.example_payload, null, 2))
      }
    }
  }, [selectedFunction, selectedProvider, selectedService, serviceMetadata])

  const testWorkflow = async () => {
    if (!selectedProvider || !selectedService || !selectedFunction) {
      setError('Please select provider, service, and function')
      return
    }

    setLoading(true)
    setError('')
    setResult(null)

    try {
      const workflowPayload = {
        steps: [
          {
            provider: selectedProvider,
            service: selectedService,
            function: selectedFunction,
            payload: payload ? JSON.parse(payload) : {}
          }
        ],
        input: {}
      }

      const response = await fetch('/api/workflow/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(workflowPayload)
      })

      const data = await response.json()
      
      if (response.ok) {
        setResult(data)
      } else {
        setError(data.error || 'Failed to execute workflow')
      }
    } catch (err) {
      setError('Failed to execute workflow: ' + (err as Error).message)
    } finally {
      setLoading(false)
    }
  }



  return (
    <div className="card">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-secondary-900">
          Proxy Tester
        </h2>
        <div className={`px-3 py-1 rounded-full text-xs font-medium ${
          isAuthenticated 
            ? 'bg-green-100 text-green-800' 
            : 'bg-red-100 text-red-800'
        }`}>
          {isAuthenticated ? 'Ready to Test' : 'Auth Required'}
        </div>
      </div>

      {/* Provider Selection */}
      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Provider
        </label>
        <select
          value={selectedProvider}
          onChange={(e) => setSelectedProvider(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="">Select a provider</option>
          {providers.map((provider) => (
            <option key={provider} value={provider}>
              {serviceMetadata[provider]?.display_name || provider.charAt(0).toUpperCase() + provider.slice(1)}
            </option>
          ))}
        </select>
      </div>

      {/* Service Selection */}
      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Service
        </label>
        <select
          value={selectedService}
          onChange={(e) => setSelectedService(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          disabled={!selectedProvider || availableServices.length === 0}
        >
          <option value="">Select a service</option>
          {availableServices.map((service) => {
            const serviceMeta = serviceMetadata[selectedProvider]?.services[service]
            return (
              <option key={service} value={service}>
                {serviceMeta?.display_name || service}
              </option>
            )
          })}
        </select>
      </div>

      {/* Function Selection */}
      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Function
        </label>
        <select
          value={selectedFunction}
          onChange={(e) => setSelectedFunction(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          disabled={!selectedService || availableFunctions.length === 0}
        >
          <option value="">Select a function</option>
          {availableFunctions.map((func) => (
            <option key={func.name} value={func.name}>
              {func.display_name}
            </option>
          ))}
        </select>
        {selectedFunction && availableFunctions.find(f => f.name === selectedFunction) && (
          <p className="text-sm text-gray-600 mt-1">
            {availableFunctions.find(f => f.name === selectedFunction)?.description}
          </p>
        )}
      </div>

      {/* Payload Input */}
      <div className="mb-4">
        <label className="block text-sm font-medium text-secondary-700 mb-2">
          Payload (JSON)
        </label>
        <textarea
          value={payload}
          onChange={(e) => setPayload(e.target.value)}
          className="input-field h-32 font-mono text-sm"
          placeholder="Enter JSON payload..."
          disabled={!isAuthenticated}
        />
      </div>

      {/* Test Button */}
      <button
        onClick={testWorkflow}
        disabled={!isAuthenticated || loading || !selectedProvider || !selectedService}
        className="btn-primary w-full mb-4 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {loading ? (
          <div className="flex items-center justify-center">
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
            Testing...
          </div>
        ) : (
          'Test Workflow'
        )}
      </button>

      {/* Error Display */}
      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
          <div className="text-sm text-red-800">
            <strong>Error:</strong> {error}
          </div>
        </div>
      )}

      {/* Result Display */}
      {result && (
        <div className="mb-4">
          <h3 className="text-lg font-medium text-secondary-800 mb-2">
            Result
          </h3>
          <div className="bg-secondary-50 border border-secondary-200 rounded-lg p-4">
            <pre className="text-sm text-secondary-800 whitespace-pre-wrap overflow-x-auto">
              {JSON.stringify(result, null, 2)}
            </pre>
          </div>
        </div>
      )}

      {/* Instructions */}
      <div className="mt-6 p-4 bg-primary-50 border border-primary-200 rounded-lg">
        <h4 className="text-sm font-medium text-primary-900 mb-2">
          How to use:
        </h4>
        <ul className="text-sm text-primary-800 space-y-1">
          <li>1. Make sure you're authenticated with Google</li>
          <li>2. Select a provider and service function</li>
          <li>3. Modify the JSON payload as needed</li>
          <li>4. Click "Test Workflow" to execute</li>
        </ul>
      </div>
    </div>
  )
}

export default ProxyTester
