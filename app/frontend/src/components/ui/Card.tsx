/**
 * Card Component - Design system card with responsive sizing
 */

import React from 'react'
import { cn, getCardClasses } from '../../design-system'

interface CardProps {
  children: React.ReactNode
  padding?: 'sm' | 'md' | 'lg' | 'xl'
  interactive?: boolean
  className?: string
  onClick?: () => void
}

export const Card: React.FC<CardProps> = ({
  children,
  padding = 'md',
  interactive = false,
  className,
  onClick,
}) => {
  return (
    <div
      className={cn(
        getCardClasses(padding, interactive),
        className
      )}
      onClick={onClick}
    >
      {children}
    </div>
  )
}
