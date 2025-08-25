import React from 'react'

interface AlertModalProps {
  isOpen: boolean
  title?: string
  message: string
  variant?: 'success' | 'error' | 'info' | 'warning'
  onClose: () => void
}

const variantStyles: Record<NonNullable<AlertModalProps['variant']>, {
  bg: string
  iconBg: string
  iconColor: string
  iconPath: string
}> = {
  success: {
    bg: 'bg-green-100',
    iconBg: 'bg-green-100',
    iconColor: 'text-green-600',
    iconPath: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'
  },
  error: {
    bg: 'bg-red-100',
    iconBg: 'bg-red-100',
    iconColor: 'text-red-600',
    iconPath: 'M12 9v2m0 4h.01M4.938 20h14.124c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.206 17.5C2.436 18.333 3.398 20 4.938 20z'
  },
  info: {
    bg: 'bg-blue-100',
    iconBg: 'bg-blue-100',
    iconColor: 'text-blue-600',
    iconPath: 'M13 16h-1v-4h-1m1-4h.01M12 20a8 8 0 100-16 8 8 0 000 16z'
  },
  warning: {
    bg: 'bg-yellow-100',
    iconBg: 'bg-yellow-100',
    iconColor: 'text-yellow-700',
    iconPath: 'M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z'
  }
}

export const AlertModal: React.FC<AlertModalProps> = ({ isOpen, title, message, variant = 'info', onClose }) => {
  if (!isOpen) return null
  const v = variantStyles[variant]

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
        <div className="flex items-center mb-4">
          <div className={`w-12 h-12 ${v.iconBg} rounded-full flex items-center justify-center mr-3`}>
            <svg className={`w-6 h-6 ${v.iconColor}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={v.iconPath} />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900">{title || (variant === 'success' ? 'Success' : variant === 'error' ? 'Error' : variant === 'warning' ? 'Warning' : 'Information')}</h3>
        </div>

        <div className="mb-6">
          <p className="text-gray-700">{message}</p>
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
