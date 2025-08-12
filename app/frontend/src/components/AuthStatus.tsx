import React from 'react'
import { UserIcon, ArrowRightOnRectangleIcon } from '@heroicons/react/24/outline'

interface AuthStatusProps {
  isAuthenticated: boolean
  onAuthChange: () => void
}

export const AuthStatus: React.FC<AuthStatusProps> = ({ isAuthenticated, onAuthChange }) => {
  return (
    <div className="flex items-center space-x-2">
      {isAuthenticated ? (
        <div className="flex items-center space-x-2 text-green-600">
          <UserIcon className="h-5 w-5" />
          <span className="text-sm font-medium">Authenticated</span>
          <button
            onClick={onAuthChange}
            className="text-gray-500 hover:text-gray-700"
            title="Refresh auth status"
          >
            <ArrowRightOnRectangleIcon className="h-4 w-4" />
          </button>
        </div>
      ) : (
        <div className="flex items-center space-x-2 text-red-600">
          <UserIcon className="h-5 w-5" />
          <span className="text-sm font-medium">Not authenticated</span>
          <button
            onClick={onAuthChange}
            className="text-blue-600 hover:text-blue-800 text-sm font-medium"
          >
            Login
          </button>
        </div>
      )}
    </div>
  )
}
