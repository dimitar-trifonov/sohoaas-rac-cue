import React from 'react'

interface AuthStatusProps {
  isAuthenticated: boolean
  authUrl: string
  onGetAuthUrl: () => void
  onAuthChange: () => void
}

const AuthStatus: React.FC<AuthStatusProps> = ({
  isAuthenticated,
  authUrl,
  onGetAuthUrl,
  onAuthChange
}) => {
  const handleLogin = () => {
    if (authUrl) {
      // Redirect in the same tab instead of opening a new one
      window.location.href = authUrl
    } else {
      onGetAuthUrl()
    }
  }

  return (
    <div className="flex items-center space-x-4">
      <div className="flex items-center space-x-2">
        <div className={`w-3 h-3 rounded-full ${
          isAuthenticated ? 'bg-green-500' : 'bg-red-500'
        }`} />
        <span className="text-sm font-medium text-secondary-700">
          {isAuthenticated ? 'Authenticated' : 'Not Authenticated'}
        </span>
      </div>
      
      {!isAuthenticated && (
        <button
          onClick={handleLogin}
          className="btn-primary text-sm"
        >
          {authUrl ? 'Login with Google' : 'Get Auth URL'}
        </button>
      )}
      
      {isAuthenticated && (
        <button
          onClick={onAuthChange}
          className="btn-secondary text-sm"
        >
          Refresh Status
        </button>
      )}
    </div>
  )
}

export default AuthStatus
