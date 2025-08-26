// SOHOAAS Authentication Store with Firebase Auth
// External state management using Zustand - outside React rendering cycle
// Following the SOHOAAS multi-user architecture

import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import { signInWithPopup, signOut, onAuthStateChanged, GoogleAuthProvider, type User as FirebaseUser } from 'firebase/auth'
import { doc, getDoc, updateDoc, serverTimestamp } from 'firebase/firestore'
import { auth, googleProvider, db } from '../config/firebase'
import type { AuthState, User } from '../types'

// Firestore-based user access control
const checkUserAccess = async (email: string, uid: string): Promise<boolean> => {
  try {
    // Check if user exists in allowed_users collection
    const userDoc = await getDoc(doc(db, 'allowed_users', email))
    
    if (userDoc.exists()) {
      const userData = userDoc.data()
      
      // Update user's UID and last login if not set
      if (!userData.uid || userData.uid !== uid) {
        await updateDoc(doc(db, 'allowed_users', email), {
          uid: uid,
          lastLogin: serverTimestamp()
        })
      } else {
        // Just update last login
        await updateDoc(doc(db, 'allowed_users', email), {
          lastLogin: serverTimestamp()
        })
      }
      
      return userData.active !== false // Default to true if not specified
    }
    
    return false // User not in allowed list
  } catch (error) {
    console.error('Error checking user access:', error)
    return false
  }
}


interface AuthStore extends AuthState {
  // Actions
  checkAuth: () => Promise<void>
  login: () => Promise<void>
  logout: () => Promise<void>
  refreshToken: () => Promise<void>
  setUser: (user: User) => void
  clearError: () => void
  firebaseUser: FirebaseUser | null
  googleAccessToken: string | null
  setGoogleAccessToken: (token: string) => void
}

export const useAuthStore = create<AuthStore>()(
  subscribeWithSelector((set, get) => ({
    // Initial state
    isAuthenticated: false,
    token: null,
    user: null,
    loading: false,
    error: null,
    firebaseUser: null,
    googleAccessToken: null,

    // Actions
    checkAuth: async () => {
      set({ loading: true, error: null })
      
      // Firebase auth state is handled by onAuthStateChanged listener
      // This method is kept for compatibility but auth state is managed by Firebase
      set({ loading: false })
    },

    login: async () => {
      set({ loading: true, error: null })
      
      try {
        const result = await signInWithPopup(auth, googleProvider)
        const user = result.user
        
        // Check user access in Firestore
        const hasAccess = await checkUserAccess(user.email || '', user.uid)
        if (!hasAccess) {
          await signOut(auth)
          set({ 
            loading: false,
            error: 'UNAUTHORIZED_ACCESS'
          })
          return
        }

        // Get the ID token for backend authentication
        const idToken = await user.getIdToken()
        
        // Store Google access token securely in backend after successful login
        const credential = GoogleAuthProvider.credentialFromResult(result)
        if (credential?.accessToken) {
          try {
            // Proxy-first pattern: try nginx proxy (same-origin) then direct backend URL
            const PROXY_BASE_URL = import.meta.env.VITE_PROXY_URL || (typeof window !== 'undefined' ? window.location.origin : 'http://localhost:3000')
            const BACKEND_BASE_URL = import.meta.env.VITE_BACKEND_URL || 'http://localhost:8081'
            const urls = [
              `${PROXY_BASE_URL}/api/v1/auth/store-google-token`,
              `${BACKEND_BASE_URL}/api/v1/auth/store-google-token`,
            ]

            let stored = false
            for (const url of urls) {
              try {
                const response = await fetch(url, {
                  method: 'POST',
                  headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${await user.getIdToken()}`,
                  },
                  body: JSON.stringify({ google_access_token: credential.accessToken })
                })
                if (response.ok) {
                  // Do not log sensitive data
                  console.log('Google access token stored securely in backend via', url)
                  stored = true
                  break
                }
              } catch (e) {
                console.warn('Store Google token failed via', url, e)
                continue
              }
            }

            if (stored) {
              // Remove from frontend state for security
              set({ googleAccessToken: null })
            } else {
              // Fallback: keep in frontend temporarily
              set({ googleAccessToken: credential.accessToken })
            }
          } catch (error) {
            console.error('Error storing Google token in backend:', error)
            // Fallback: keep in frontend temporarily
            set({ googleAccessToken: credential.accessToken })
          }
        } else {
          console.warn('No Google access token received from Firebase Auth')
        }
        
        // Create user object for SOHOAAS backend
        const sohoaasUser: User = {
          user_id: user.uid,
          email: user.email || '',
          name: user.displayName || '',
          oauth_tokens: {
            google: {
              access_token: '', // Token now stored securely in backend
              token_type: 'Bearer'
            }
          },
          connected_services: ['gmail', 'calendar', 'docs', 'drive']
        }

        set({ 
          isAuthenticated: true, 
          token: { access_token: idToken, token_type: 'Bearer', expires_in: 3600, valid: true },
          user: sohoaasUser,
          firebaseUser: user,
          googleAccessToken: null, // Token now stored securely in backend
          loading: false 
        })

        // Load workflows after successful authentication
        try {
          const { useWorkflowStore } = await import('./workflowStore')
          await useWorkflowStore.getState().loadWorkflows()
        } catch (error) {
          console.error('Failed to load workflows after authentication:', error)
        }

      } catch (error) {
        console.error('Login failed:', error)
        set({ 
          loading: false,
          error: error instanceof Error ? error.message : 'Login failed'
        })
      }
    },

    logout: async () => {
      try {
        await signOut(auth)
        set({ 
          isAuthenticated: false, 
          token: null, 
          user: null,
          firebaseUser: null,
          loading: false,
          error: null
        })
      } catch (error) {
        console.error('Logout failed:', error)
        set({ error: error instanceof Error ? error.message : 'Logout failed' })
      }
    },

    refreshToken: async () => {
      const { firebaseUser } = get()
      if (firebaseUser) {
        try {
          const idToken = await firebaseUser.getIdToken(true) // Force refresh
          set({ 
            token: { access_token: idToken, token_type: 'Bearer', expires_in: 3600, valid: true }
          })
        } catch (error) {
          console.error('Token refresh failed:', error)
          set({ error: error instanceof Error ? error.message : 'Token refresh failed' })
        }
      }
    },

    setUser: (user: User) => {
      set({ user })
    },

    clearError: () => {
      set({ error: null })
    },

    setGoogleAccessToken: (token: string) => {
      set({ googleAccessToken: token })
    }
  }))
)

// Firebase auth state listener
onAuthStateChanged(auth, async (firebaseUser: FirebaseUser | null) => {
  // const store = useAuthStore.getState() // Removed unused variable
  const set = (updates: Partial<AuthStore>) => {
    useAuthStore.setState(updates)
  }
  
  if (firebaseUser) {
    // Check user access in Firestore
    const hasAccess = await checkUserAccess(firebaseUser.email || '', firebaseUser.uid)
    if (!hasAccess) {
      await signOut(auth)
      set({ 
        isAuthenticated: false,
        token: null,
        user: null,
        firebaseUser: null,
        error: 'UNAUTHORIZED_ACCESS'
      })
      return
    }

    try {
      const idToken = await firebaseUser.getIdToken()
      
      const sohoaasUser: User = {
        user_id: firebaseUser.uid,
        email: firebaseUser.email || '',
        name: firebaseUser.displayName || '',
        oauth_tokens: {
          google: {
            access_token: idToken,
            token_type: 'Bearer'
          }
        },
        connected_services: ['gmail', 'calendar', 'docs', 'drive']
      }

      set({ 
        isAuthenticated: true, 
        token: { access_token: idToken, token_type: 'Bearer', expires_in: 3600, valid: true },
        user: sohoaasUser,
        firebaseUser: firebaseUser,
        loading: false,
        error: null
      })

      // After auto sign-in, verify backend has Google token; if missing, auto-logout
      try {
        const PROXY_BASE_URL = import.meta.env.VITE_PROXY_URL || (typeof window !== 'undefined' ? window.location.origin : 'http://localhost:3000')
        const BACKEND_BASE_URL = import.meta.env.VITE_BACKEND_URL || 'http://localhost:8081'
        const urls = [
          `${PROXY_BASE_URL}/api/v1/auth/token-info`,
          `${BACKEND_BASE_URL}/api/v1/auth/token-info`,
        ]

        let ok = false
        for (const url of urls) {
          try {
            const resp = await fetch(url, {
              headers: {
                'Authorization': `Bearer ${await firebaseUser.getIdToken()}`,
              }
            })
            if (resp.ok) { ok = true; break }
            if (resp.status === 404 || resp.status === 401) { ok = false; break }
          } catch (_) {
            continue
          }
        }

        if (!ok) {
          // Missing Google token in backend; sign out to force user to re-consent and store token
          await signOut(auth)
          set({ 
            isAuthenticated: false,
            token: null,
            user: null,
            firebaseUser: null,
            loading: false,
            error: 'MISSING_GOOGLE_TOKEN'
          })
          return
        }
      } catch (e) {
        // Non-fatal; keep user signed in
        console.warn('Token info check failed:', e)
      }

      // Load workflows after successful authentication
      try {
        const { useWorkflowStore } = await import('./workflowStore')
        await useWorkflowStore.getState().loadWorkflows()
      } catch (error) {
        console.error('Failed to load workflows after authentication:', error)
      }
    } catch (error) {
      console.error('Failed to get ID token:', error)
      set({ 
        isAuthenticated: false,
        token: null,
        user: null,
        firebaseUser: null,
        error: error instanceof Error ? error.message : 'Authentication failed'
      })
    }
  } else {
    set({ 
      isAuthenticated: false, 
      token: null, 
      user: null,
      firebaseUser: null,
      loading: false,
      error: null
    })
  }
})

// Auto-refresh token every 30 minutes
setInterval(() => {
  const { isAuthenticated, refreshToken } = useAuthStore.getState()
  if (isAuthenticated) {
    refreshToken()
  }
}, 30 * 60 * 1000)

// Subscribe to auth changes for logging
useAuthStore.subscribe(
  (state) => state.isAuthenticated,
  (isAuthenticated) => {
    console.log('Auth status changed:', isAuthenticated ? 'Authenticated' : 'Not authenticated')
  }
)
