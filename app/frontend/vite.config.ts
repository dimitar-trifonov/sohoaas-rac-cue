import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  
  // Environment variable configuration for Docker deployment
  define: {
    // Expose environment variables to the client
    __VITE_PROXY_URL__: JSON.stringify(process.env.VITE_PROXY_URL || 'http://localhost:3000'),
    __VITE_MCP_URL__: JSON.stringify(process.env.VITE_MCP_URL || 'http://localhost:8080'),
    __VITE_BACKEND_URL__: JSON.stringify(process.env.VITE_BACKEND_URL || 'http://localhost:8081'),
  },
  
  // Server configuration for development
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      // Proxy API calls during development
      '/api': {
        target: process.env.VITE_PROXY_URL || 'http://localhost:3000',
        changeOrigin: true,
        secure: false,
      }
    }
  },
  
  // Build configuration
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          stores: ['zustand'],
        }
      }
    }
  }
})
