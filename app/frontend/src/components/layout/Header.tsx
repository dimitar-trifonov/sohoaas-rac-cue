/**
 * Header Component - Responsive header optimized for large screen presentations
 */

import React from 'react'
import { cn, typographyVariants } from '../../design-system'
import { Container } from './Container'

interface HeaderProps {
  title?: string
  subtitle?: string
  children?: React.ReactNode
  demoMode?: boolean
  className?: string
}

export const Header: React.FC<HeaderProps> = ({
  title = 'SOHOAAS',
  subtitle = 'Small Office/Home Office Automation as a Service',
  children,
  demoMode = false,
  className,
}) => {
  return (
    <header className={cn('bg-white shadow-sm border-b', className)}>
      <Container size={demoMode ? 'demo' : 'xl'}>
        <div className="flex justify-between items-center h-16 lg:h-20 xl:h-24">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <h1 className={cn(
                demoMode 
                  ? typographyVariants.demo.title
                  : 'text-2xl lg:text-3xl xl:text-4xl font-bold text-gray-900'
              )}>
                {title}
              </h1>
              <p className={cn(
                demoMode
                  ? typographyVariants.demo.subtitle
                  : 'text-sm lg:text-base xl:text-lg text-gray-500'
              )}>
                {subtitle}
              </p>
            </div>
          </div>
          
          {children && (
            <div className="flex items-center space-x-4 lg:space-x-6">
              {children}
            </div>
          )}
        </div>
      </Container>
    </header>
  )
}
