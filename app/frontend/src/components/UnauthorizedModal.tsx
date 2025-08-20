import React from 'react'

interface UnauthorizedModalProps {
  isOpen: boolean
  onClose: () => void
}

export const UnauthorizedModal: React.FC<UnauthorizedModalProps> = ({ isOpen, onClose }) => {
  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
        <div className="flex items-center mb-4">
          <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mr-3">
            <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900">Access Denied</h3>
        </div>
        
        <div className="mb-6">
          <p className="text-gray-700 mb-4">
            Your account is not authorized to access this SOHOAAS demo application.
          </p>
          <p className="text-gray-700 mb-4">
            To request access, please contact:
          </p>
          <div className="bg-gray-50 p-3 rounded-md border">
            <p className="font-medium text-gray-900">Dimitar Trifonov</p>
            <a 
              href="mailto:trifonov.dimitar@gmail.com?subject=SOHOAAS Demo Access Request"
              className="text-blue-600 hover:text-blue-800 underline"
            >
              trifonov.dimitar@gmail.com
            </a>
          </div>
          <p className="text-sm text-gray-600 mt-3">
            Please include your email address and a brief description of your interest in the demo.
          </p>
        </div>
        
        <div className="flex justify-end">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  )
}
