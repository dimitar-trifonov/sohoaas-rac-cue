/**
 * Navigation Component - Responsive navigation optimized for large screen presentations
 */

import React from 'react'
import { PlusIcon, ClockIcon } from '@heroicons/react/24/outline'
import { cn } from '../../design-system'
import { Container } from './Container'

interface NavigationProps {
  activeTab: 'create' | 'workflows' | 'agents'
  onTabChange: (tab: 'create' | 'workflows' | 'agents') => void
  workflowCount?: number
  demoMode?: boolean
  className?: string
}

export const Navigation: React.FC<NavigationProps> = ({
  activeTab,
  onTabChange,
  workflowCount = 0,
  demoMode = false,
  className,
}) => {
  return (
    <nav className={cn('bg-white border-b', className)}>
      <Container size={demoMode ? 'demo' : 'xl'}>
        <div className="flex space-x-8 lg:space-x-12">
          <button
            onClick={() => onTabChange('create')}
            className={cn(
              'py-4 lg:py-6 px-1 border-b-2 font-medium transition-colors duration-200',
              'text-sm lg:text-base xl:text-lg',
              'flex items-center space-x-2 lg:space-x-3',
              activeTab === 'create'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            )}
          >
            <PlusIcon className={cn(
              'inline-block',
              demoMode ? 'w-8 h-8' : 'w-5 h-5 lg:w-6 lg:h-6'
            )} />
            <span>Create Workflow</span>
          </button>
          
          <button
            onClick={() => onTabChange('workflows')}
            className={cn(
              'py-4 lg:py-6 px-1 border-b-2 font-medium transition-colors duration-200',
              'text-sm lg:text-base xl:text-lg',
              'flex items-center space-x-2 lg:space-x-3',
              activeTab === 'workflows'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            )}
          >
            <ClockIcon className={cn(
              'inline-block',
              demoMode ? 'w-8 h-8' : 'w-5 h-5 lg:w-6 lg:h-6'
            )} />
            <span>My Workflows ({workflowCount})</span>
          </button>
        </div>
      </Container>
    </nav>
  )
}
