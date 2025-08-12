// SOHOAAS UI Store
// External state management using Zustand - outside React rendering cycle
// Following the SOHOAAS 5-agent PoC system architecture

import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import type { UIState, Notification } from '../types'

interface UIStore extends UIState {
  // Actions
  setActiveTab: (tab: 'create' | 'workflows' | 'agents') => void
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  addNotification: (notification: Omit<Notification, 'id' | 'timestamp' | 'read'>) => void
  markNotificationRead: (id: string) => void
  removeNotification: (id: string) => void
  clearNotifications: () => void
}

export const useUIStore = create<UIStore>()(
  subscribeWithSelector((set, get) => ({
    // Initial state
    activeTab: 'create',
    sidebarOpen: false,
    notifications: [],

    // Actions
    setActiveTab: (tab: 'create' | 'workflows' | 'agents') => {
      set({ activeTab: tab })
    },

    toggleSidebar: () => {
      set((state) => ({ sidebarOpen: !state.sidebarOpen }))
    },

    setSidebarOpen: (open: boolean) => {
      set({ sidebarOpen: open })
    },

    addNotification: (notification: Omit<Notification, 'id' | 'timestamp' | 'read'>) => {
      const newNotification: Notification = {
        ...notification,
        id: `notification_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        timestamp: new Date().toISOString(),
        read: false
      }

      set((state) => ({
        notifications: [newNotification, ...state.notifications]
      }))

      // Auto-remove success notifications after 5 seconds
      if (notification.type === 'success') {
        setTimeout(() => {
          get().removeNotification(newNotification.id)
        }, 5000)
      }
    },

    markNotificationRead: (id: string) => {
      set((state) => ({
        notifications: state.notifications.map(notification =>
          notification.id === id
            ? { ...notification, read: true }
            : notification
        )
      }))
    },

    removeNotification: (id: string) => {
      set((state) => ({
        notifications: state.notifications.filter(notification => notification.id !== id)
      }))
    },

    clearNotifications: () => {
      set({ notifications: [] })
    }
  }))
)

// Subscribe to notifications for logging
useUIStore.subscribe(
  (state) => state.notifications.length,
  (notificationCount) => {
    console.log('Notification count changed:', notificationCount)
  }
)
