/**
 * SOHOAAS Design System - Component Variants
 * Reusable component patterns for consistent UI
 */

// Design tokens available for future component extensions

// === BUTTON VARIANTS ===
export const buttonVariants = {
  // Base styles
  base: [
    'inline-flex items-center justify-center',
    'font-medium rounded-lg',
    'transition-all duration-200',
    'focus:outline-none focus:ring-2 focus:ring-offset-2',
    'disabled:opacity-50 disabled:cursor-not-allowed',
  ].join(' '),

  // Size variants
  size: {
    sm: 'h-8 px-3 text-sm',
    md: 'h-10 px-4 text-base',
    lg: 'h-12 px-6 text-lg',
    xl: 'h-14 px-8 text-xl', // For large screen demos
  },

  // Color variants
  variant: {
    primary: [
      'bg-primary-600 text-white',
      'hover:bg-primary-700',
      'focus:ring-primary-500',
      'shadow-sm hover:shadow-md',
    ].join(' '),
    
    secondary: [
      'bg-gray-100 text-gray-900',
      'hover:bg-gray-200',
      'focus:ring-gray-500',
      'border border-gray-300',
    ].join(' '),
    
    outline: [
      'bg-transparent text-primary-600',
      'border-2 border-primary-600',
      'hover:bg-primary-50',
      'focus:ring-primary-500',
    ].join(' '),
    
    ghost: [
      'bg-transparent text-gray-600',
      'hover:bg-gray-100',
      'focus:ring-gray-500',
    ].join(' '),
  },
}

// === CARD VARIANTS ===
export const cardVariants = {
  base: [
    'bg-white rounded-lg shadow-sm',
    'border border-gray-200',
    'transition-shadow duration-200',
  ].join(' '),

  padding: {
    sm: 'p-4',
    md: 'p-6',
    lg: 'p-8',
    xl: 'p-12', // Generous padding for large screens
  },

  hover: 'hover:shadow-md',
  interactive: 'cursor-pointer hover:shadow-lg transform hover:-translate-y-1',
}

// === INPUT VARIANTS ===
export const inputVariants = {
  base: [
    'block w-full rounded-lg border-gray-300',
    'shadow-sm transition-colors duration-200',
    'focus:border-primary-500 focus:ring-primary-500',
    'disabled:bg-gray-50 disabled:text-gray-500',
  ].join(' '),

  size: {
    sm: 'h-8 px-3 text-sm',
    md: 'h-10 px-4 text-base',
    lg: 'h-12 px-4 text-lg',
    xl: 'h-14 px-6 text-xl', // For large screen demos
  },

  state: {
    error: 'border-error-500 focus:border-error-500 focus:ring-error-500',
    success: 'border-success-500 focus:border-success-500 focus:ring-success-500',
  },
}

// === LAYOUT VARIANTS ===
export const layoutVariants = {
  container: {
    base: 'mx-auto px-4 sm:px-6 lg:px-8',
    size: {
      sm: 'max-w-3xl',
      md: 'max-w-5xl',
      lg: 'max-w-7xl',
      xl: 'max-w-screen-2xl',
      demo: 'max-w-[1800px]', // Optimized for large screen demos
    },
  },

  section: {
    base: 'py-8 lg:py-12',
    hero: 'py-16 lg:py-24',
    compact: 'py-4 lg:py-6',
  },

  grid: {
    responsive: 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
    demo: 'grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4', // Demo-optimized
  },
}

// === TYPOGRAPHY VARIANTS ===
export const typographyVariants = {
  heading: {
    h1: 'text-4xl lg:text-5xl xl:text-6xl font-bold text-gray-900',
    h2: 'text-3xl lg:text-4xl xl:text-5xl font-bold text-gray-900',
    h3: 'text-2xl lg:text-3xl xl:text-4xl font-semibold text-gray-900',
    h4: 'text-xl lg:text-2xl xl:text-3xl font-semibold text-gray-900',
    h5: 'text-lg lg:text-xl xl:text-2xl font-medium text-gray-900',
    h6: 'text-base lg:text-lg xl:text-xl font-medium text-gray-900',
  },

  body: {
    large: 'text-lg lg:text-xl text-gray-700',
    base: 'text-base lg:text-lg text-gray-700',
    small: 'text-sm lg:text-base text-gray-600',
    caption: 'text-xs lg:text-sm text-gray-500',
  },

  // Demo-specific typography (larger for presentations)
  demo: {
    title: 'text-5xl lg:text-6xl xl:text-7xl font-bold text-gray-900',
    subtitle: 'text-2xl lg:text-3xl xl:text-4xl text-gray-600',
    body: 'text-xl lg:text-2xl text-gray-700',
  },
}

// === STATUS VARIANTS ===
export const statusVariants = {
  badge: {
    base: 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
    size: {
      sm: 'px-2 py-1 text-xs',
      md: 'px-3 py-1 text-sm',
      lg: 'px-4 py-2 text-base', // For large screen visibility
    },
    variant: {
      success: 'bg-success-100 text-success-800',
      warning: 'bg-warning-100 text-warning-800',
      error: 'bg-error-100 text-error-800',
      info: 'bg-primary-100 text-primary-800',
      neutral: 'bg-gray-100 text-gray-800',
    },
  },
}

// === UTILITY FUNCTIONS ===
export const cn = (...classes: (string | undefined | null | false)[]) => {
  return classes.filter(Boolean).join(' ')
}

export const getButtonClasses = (
  size: keyof typeof buttonVariants.size = 'md',
  variant: keyof typeof buttonVariants.variant = 'primary'
) => {
  return cn(
    buttonVariants.base,
    buttonVariants.size[size],
    buttonVariants.variant[variant]
  )
}

export const getCardClasses = (
  padding: keyof typeof cardVariants.padding = 'md',
  interactive: boolean = false
) => {
  return cn(
    cardVariants.base,
    cardVariants.padding[padding],
    interactive && cardVariants.interactive
  )
}
