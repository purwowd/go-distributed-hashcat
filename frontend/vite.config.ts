import { defineConfig } from 'vite'
import { resolve } from 'path'
import fs from 'fs'

// Custom plugin to compile HTML templates
const htmlTemplatePlugin = () => {
  return {
    name: 'html-template',
    transformIndexHtml: {
      order: 'pre' as const,
      handler(html: string, context: any) {
        try {
          // Read component files
          const componentsPath = resolve(__dirname, 'src/components')
          const templatesPath = resolve(__dirname, 'src/templates')
          
          // Check if component system exists
          if (!fs.existsSync(componentsPath)) {
            console.log('Component system not yet migrated, using existing HTML')
            return html
          }

          // Load components if they exist
          const components: Record<string, string> = {}
          
          const loadComponent = (name: string, path: string) => {
            const fullPath = resolve(componentsPath, path)
            if (fs.existsSync(fullPath)) {
              components[name] = fs.readFileSync(fullPath, 'utf-8')
            } else {
              components[name] = `<!-- Component ${name} not found -->`
            }
          }

          // Load all components
          loadComponent('navigation', 'layout/navigation.html')
          loadComponent('overview', 'tabs/overview.html')
          loadComponent('agents', 'tabs/agents.html')
          loadComponent('jobs', 'tabs/jobs.html')
          loadComponent('files', 'tabs/files.html')
          loadComponent('wordlists', 'tabs/wordlists.html')
          loadComponent('docs', 'tabs/docs.html')
          loadComponent('modals', 'modals/all-modals.html')
          loadComponent('notifications', 'ui/notifications.html')
          loadComponent('loading', 'ui/loading.html')
          
          // Combine all tab content
          const content = `
            ${components.overview}
            ${components.agents}
            ${components.jobs}
            ${components.files}
            ${components.wordlists}
            ${components.docs}
          `
          
          // Load base template if it exists
          const baseTemplatePath = resolve(templatesPath, 'base.html')
          if (fs.existsSync(baseTemplatePath)) {
            const baseTemplate = fs.readFileSync(baseTemplatePath, 'utf-8')
            
            // Replace template variables
            return baseTemplate
              .replace('{{ navigation | safe }}', components.navigation)
              .replace('{{ content | safe }}', content)
              .replace('{{ modals | safe }}', components.modals)
              .replace('{{ notifications | safe }}', components.notifications)
              .replace('{{ loading | safe }}', components.loading)
              .replace('{{ title | default(\'Distributed Hashcat Dashboard\') }}', 'Distributed Hashcat Dashboard')
          }
          
          return html // fallback to original HTML
        } catch (error) {
          console.warn('Template compilation warning:', error.message)
          return html // fallback to original HTML if template compilation fails
        }
      }
    }
  }
}

export default defineConfig({
  plugins: [htmlTemplatePlugin()],
  root: 'src',
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    // Production optimizations
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true
      }
    },
    // Rollup options for input and chunking
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'src/index.html')
      },
      output: {
        manualChunks: {
          vendor: ['alpinejs'],
          utils: ['./src/utils/component-loader.ts', './src/config/build.config.ts']
        }
      }
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:1337',
        changeOrigin: true
      },
      '/health': {
        target: 'http://localhost:1337', 
        changeOrigin: true
      }
    }
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@components': resolve(__dirname, 'src/components'),
      '@templates': resolve(__dirname, 'src/templates'),
      '@styles': resolve(__dirname, 'src/styles'),
      '@utils': resolve(__dirname, 'src/utils'),
      '@config': resolve(__dirname, 'src/config'),
      '@services': resolve(__dirname, 'src/services'),
      '@stores': resolve(__dirname, 'src/stores')
    }
  },
  optimizeDeps: {
    include: ['alpinejs']
  },
  // Environment variables
  define: {
    __APP_VERSION__: JSON.stringify(process.env.npm_package_version || '1.0.0')
  }
}) 
