import React from 'react'
import { UserIcon, ArrowRightOnRectangleIcon } from '@heroicons/react/24/outline'
import { useAuthStore } from '../stores'

interface AuthStatusProps {
  isAuthenticated: boolean
  onAuthChange?: () => void
}

export const AuthStatus: React.FC<AuthStatusProps> = ({ isAuthenticated, onAuthChange }) => {
  const { login, logout, user } = useAuthStore()

  const handleAuthAction = () => {
    if (isAuthenticated) {
      logout()
    } else {
      login()
    }
    onAuthChange?.()
  }

  return (
    <div className="flex items-center space-x-2">
      {isAuthenticated ? (
        <div className="flex items-center space-x-2 text-green-600">
          <UserIcon className="h-5 w-5" />
          <span className="text-sm font-medium">
            {user?.email || 'Authenticated'}
          </span>
          <button
            onClick={handleAuthAction}
            className="text-gray-500 hover:text-gray-700"
            title="Logout"
          >
            <ArrowRightOnRectangleIcon className="h-4 w-4" />
          </button>
        </div>
      ) : (
        <div className="flex items-center space-x-2 text-red-600">
          <UserIcon className="h-5 w-5" />
          <span className="text-sm font-medium">Not authenticated</span>
          <button
            onClick={handleAuthAction}
            className="text-blue-600 hover:text-blue-800 text-sm font-medium"
          >
            Connect Google Workspace
          </button>
        </div>
      )}
    </div>
  )
}
