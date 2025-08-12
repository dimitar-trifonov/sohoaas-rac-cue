/**
 * Container Component - Responsive container with demo-optimized sizing
 */

import React from 'react'
import { cn, layoutVariants } from '../../design-system'

interface ContainerProps {
  children: React.ReactNode
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'demo'
  className?: string
}

export const Container: React.FC<ContainerProps> = ({
  children,
  size = 'xl',
  className,
}) => {
  return (
    <div
      className={cn(
        layoutVariants.container.base,
        layoutVariants.container.size[size],
        className
      )}
    >
      {children}
    </div>
  )
}
