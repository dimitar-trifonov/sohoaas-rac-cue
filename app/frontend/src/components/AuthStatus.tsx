import React from 'react'
import { UserIcon, ArrowRightOnRectangleIcon, ArrowLeftOnRectangleIcon } from '@heroicons/react/24/outline'
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
    <div className="flex items-center space-x-3">
      {/* Status icon only: green when authenticated, red when not */}
      <UserIcon
        className={`h-5 w-5 ${isAuthenticated ? 'text-green-600' : 'text-red-600'}`}
        title={isAuthenticated ? (user?.email || 'Authenticated') : 'Not authenticated'}
      />

      {/* Auth action icon: logout when authenticated, login when not */}
      <button
        onClick={handleAuthAction}
        className="text-gray-500 hover:text-gray-700"
        title={isAuthenticated ? 'Logout' : 'Login with Google Workspace'}
        aria-label={isAuthenticated ? 'Logout' : 'Login'}
      >
        {isAuthenticated ? (
          <ArrowRightOnRectangleIcon className="h-5 w-5" />
        ) : (
          <ArrowLeftOnRectangleIcon className="h-5 w-5" />
        )}
      </button>
    </div>
  )
}
