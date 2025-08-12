import React, { useState, useEffect, useRef } from 'react'

interface MCPResource {
  uri: string
  name: string
  description?: string
  mimeType?: string
}

interface MCPTool {
  name: string
  description?: string
  inputSchema: any
}

interface MCPClientProps {
  isAuthenticated: boolean
}

const MCPClient: React.FC<MCPClientProps> = ({ isAuthenticated }) => {
  const [connected, setConnected] = useState(false)
  const [connecting, setConnecting] = useState(false)
  const [resources, setResources] = useState<MCPResource[]>([])
  const [tools, setTools] = useState<MCPTool[]>([])
  const [selectedTool, setSelectedTool] = useState<MCPTool | null>(null)
  const [toolArgs, setToolArgs] = useState('')
  const [toolResult, setToolResult] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [logs, setLogs] = useState<string[]>([])
  
  const wsRef = useRef<WebSocket | null>(null)
  const requestIdRef = useRef(1)

  const addLog = (message: string) => {
    setLogs(prev => [...prev.slice(-19), `${new Date().toLocaleTimeString()}: ${message}`])
  }

  const connectToMCP = async () => {
    if (!isAuthenticated) {
      setError('Authentication required to connect to MCP server')
      return
    }

    setConnecting(true)
    setError('')
    
    try {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      // Connect through nginx proxy on the same host/port as the frontend
      const wsUrl = `${protocol}//${window.location.host}/mcp`
      
      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        addLog('Connected to MCP server')
        setConnected(true)
        setConnecting(false)
        
        // Send initialize request
        const initRequest = {
          jsonrpc: '2.0',
          id: requestIdRef.current++,
          method: 'initialize',
          params: {
            protocolVersion: '2024-11-05',
            capabilities: {
              roots: {
                listChanged: true
              },
              sampling: {}
            },
            clientInfo: {
              name: 'RAC-MCP-Client',
              version: '1.0.0'
            }
          }
        }
        
        ws.send(JSON.stringify(initRequest))
        addLog('Sent initialize request')
      }

      ws.onmessage = (event) => {
        try {
          const response = JSON.parse(event.data)
          handleMCPResponse(response)
        } catch (err) {
          addLog(`Error parsing message: ${err}`)
        }
      }

      ws.onclose = () => {
        addLog('Disconnected from MCP server')
        setConnected(false)
        setConnecting(false)
      }

      ws.onerror = (err) => {
        addLog(`WebSocket error: ${err}`)
        setError('Failed to connect to MCP server')
        setConnecting(false)
      }

    } catch (err) {
      setError(`Connection failed: ${err}`)
      setConnecting(false)
    }
  }

  const handleMCPResponse = (response: any) => {
    addLog(`Received: ${response.method || 'response'} (ID: ${response.id})`)
    
    if (response.method === 'initialize' || response.id === 1) {
      // Initialize response - now load resources and tools
      loadResources()
      loadTools()
    }
  }

  const sendMCPRequest = (method: string, params: any = {}) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      setError('Not connected to MCP server')
      return
    }

    const request = {
      jsonrpc: '2.0',
      id: requestIdRef.current++,
      method,
      params
    }

    wsRef.current.send(JSON.stringify(request))
    addLog(`Sent: ${method}`)
    return request.id
  }

  const loadResources = () => {
    const requestId = sendMCPRequest('resources/list')
    
    // Handle response (in a real implementation, you'd track request IDs)
    const handleResourcesResponse = (event: MessageEvent) => {
      try {
        const response = JSON.parse(event.data)
        if (response.id === requestId && response.result) {
          setResources(response.result.resources || [])
          addLog(`Loaded ${response.result.resources?.length || 0} resources`)
        }
      } catch (err) {
        addLog(`Error handling resources response: ${err}`)
      }
    }

    if (wsRef.current) {
      wsRef.current.addEventListener('message', handleResourcesResponse, { once: true })
    }
  }

  const loadTools = () => {
    const requestId = sendMCPRequest('tools/list')
    
    const handleToolsResponse = (event: MessageEvent) => {
      try {
        const response = JSON.parse(event.data)
        if (response.id === requestId && response.result) {
          setTools(response.result.tools || [])
          addLog(`Loaded ${response.result.tools?.length || 0} tools`)
        }
      } catch (err) {
        addLog(`Error handling tools response: ${err}`)
      }
    }

    if (wsRef.current) {
      wsRef.current.addEventListener('message', handleToolsResponse, { once: true })
    }
  }

  const executeTool = async () => {
    if (!selectedTool) return

    setLoading(true)
    setError('')
    setToolResult(null)

    try {
      const args = toolArgs ? JSON.parse(toolArgs) : {}
      const requestId = sendMCPRequest('tools/call', {
        name: selectedTool.name,
        arguments: args
      })

      const handleToolResponse = (event: MessageEvent) => {
        try {
          const response = JSON.parse(event.data)
          if (response.id === requestId) {
            if (response.result) {
              setToolResult(response.result)
              addLog(`Tool executed successfully: ${selectedTool.name}`)
            } else if (response.error) {
              setError(`Tool execution failed: ${response.error.message}`)
              addLog(`Tool execution failed: ${response.error.message}`)
            }
            setLoading(false)
          }
        } catch (err) {
          setError(`Error handling tool response: ${err}`)
          setLoading(false)
        }
      }

      if (wsRef.current) {
        wsRef.current.addEventListener('message', handleToolResponse, { once: true })
      }

    } catch (err) {
      setError(`Invalid JSON in tool arguments: ${err}`)
      setLoading(false)
    }
  }

  const disconnect = () => {
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    setConnected(false)
    setResources([])
    setTools([])
    setSelectedTool(null)
    setToolResult(null)
    addLog('Disconnected')
  }

  useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [])

  return (
    <div className="card">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-secondary-900">
          MCP Client
        </h2>
        <div className={`px-3 py-1 rounded-full text-xs font-medium ${
          connected 
            ? 'bg-green-100 text-green-800' 
            : connecting
            ? 'bg-yellow-100 text-yellow-800'
            : 'bg-gray-100 text-gray-800'
        }`}>
          {connected ? 'Connected' : connecting ? 'Connecting...' : 'Disconnected'}
        </div>
      </div>

      {/* Connection Controls */}
      <div className="mb-6">
        {!connected ? (
          <button
            onClick={connectToMCP}
            disabled={!isAuthenticated || connecting}
            className="btn-primary"
          >
            {connecting ? 'Connecting...' : 'Connect to MCP Server'}
          </button>
        ) : (
          <button
            onClick={disconnect}
            className="btn-secondary"
          >
            Disconnect
          </button>
        )}
        
        {!isAuthenticated && (
          <p className="text-sm text-yellow-600 mt-2">
            Authentication required to connect to MCP server
          </p>
        )}
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}

      {connected && (
        <div className="space-y-6">
          {/* Resources Section */}
          <div>
            <h3 className="text-lg font-medium text-secondary-800 mb-3">
              Available Resources ({resources.length})
            </h3>
            <div className="grid grid-cols-1 gap-2">
              {resources.map((resource, index) => (
                <div
                  key={index}
                  className="p-3 bg-secondary-50 rounded-lg border border-secondary-200"
                >
                  <div className="font-medium text-secondary-900">{resource.name}</div>
                  <div className="text-sm text-secondary-600">{resource.description}</div>
                  <div className="text-xs text-secondary-500 mt-1">URI: {resource.uri}</div>
                </div>
              ))}
            </div>
          </div>

          {/* Tools Section */}
          <div>
            <h3 className="text-lg font-medium text-secondary-800 mb-3">
              Available Tools ({tools.length})
            </h3>
            
            <div className="mb-4">
              <label className="block text-sm font-medium text-secondary-700 mb-2">
                Select Tool
              </label>
              <select
                value={selectedTool?.name || ''}
                onChange={(e) => {
                  const tool = tools.find(t => t.name === e.target.value)
                  setSelectedTool(tool || null)
                  setToolArgs('')
                  setToolResult(null)
                }}
                className="input-field"
              >
                <option value="">Choose a tool...</option>
                {tools.map((tool) => (
                  <option key={tool.name} value={tool.name}>
                    {tool.name}
                  </option>
                ))}
              </select>
            </div>

            {selectedTool && (
              <div className="space-y-4">
                <div>
                  <p className="text-sm text-secondary-600 mb-2">
                    {selectedTool.description}
                  </p>
                  
                  <label className="block text-sm font-medium text-secondary-700 mb-2">
                    Tool Arguments (JSON)
                  </label>
                  <textarea
                    value={toolArgs}
                    onChange={(e) => setToolArgs(e.target.value)}
                    placeholder='{"key": "value"}'
                    className="input-field h-24 font-mono text-sm"
                  />
                </div>

                <button
                  onClick={executeTool}
                  disabled={loading}
                  className="btn-primary"
                >
                  {loading ? 'Executing...' : 'Execute Tool'}
                </button>

                {toolResult && (
                  <div>
                    <h4 className="text-sm font-medium text-secondary-700 mb-2">
                      Tool Result:
                    </h4>
                    <pre className="bg-secondary-50 p-3 rounded-lg text-sm overflow-auto max-h-64">
                      {JSON.stringify(toolResult, null, 2)}
                    </pre>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Connection Log */}
          <div>
            <h3 className="text-lg font-medium text-secondary-800 mb-3">
              Connection Log
            </h3>
            <div className="bg-gray-900 text-green-400 p-3 rounded-lg font-mono text-xs max-h-48 overflow-y-auto">
              {logs.map((log, index) => (
                <div key={index}>{log}</div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default MCPClient
