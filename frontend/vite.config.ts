import { defineConfig, loadEnv } from 'vite'
import path from 'path'
import fs from 'fs'
import { fileURLToPath } from 'url'

// Get __dirname equivalent for ES modules
const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

// Custom plugin to compile HTML templates
const htmlTemplatePlugin = () => {
  return {
    name: 'html-template',
    transformIndexHtml: {
      order: 'pre' as const,
      handler(html: string, context: any) {
        try {
          // Read component files
          const componentsPath = path.resolve(__dirname, 'src/components')
          const templatesPath = path.resolve(__dirname, 'src/templates')
          
          // Check if component system exists
          if (!fs.existsSync(componentsPath)) {
            console.log('Component system not yet migrated, using existing HTML')
            return html
          }

          // Load components if they exist
          const components: Record<string, string> = {}
          
                      const loadComponent = (name: string, componentPath: string) => {
             const fullPath = path.resolve(componentsPath, componentPath)
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
          const baseTemplatePath = path.resolve(templatesPath, 'base.html')
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
          console.warn('Template compilation warning:', error instanceof Error ? error.message : String(error))
          return html // fallback to original HTML if template compilation fails
        }
      }
    }
  }
}

export default defineConfig(({ command, mode }: { command: string; mode: string }) => {
  // Load env file based on `mode` in the current working directory.
  const env = loadEnv(mode, process.cwd(), '')
  
  return {
    // Root directory configuration
    root: path.resolve(__dirname, 'src'),
    // Development server configuration
    server: {
      port: 3000,
      host: true, // Allow external connections
      cors: true,
      proxy: {
        // Proxy API calls to backend during development
        '/api': {
          target: env.VITE_API_BASE_URL || 'http://localhost:1337',
          changeOrigin: true,
          secure: false,
        }
      }
    },
    
    // Build configuration
    build: {
      target: 'es2015',
      outDir: '../dist',
      assetsDir: 'assets',
      minify: 'terser' as const,
      sourcemap: mode === 'development',
      rollupOptions: {
        output: {
          manualChunks: {
            vendor: ['alpinejs'],
          },
          chunkFileNames: 'assets/[name]-[hash].js',
          entryFileNames: 'assets/[name]-[hash].js',
          assetFileNames: 'assets/[name]-[hash].[ext]'
        }
      },
      terserOptions: {
        compress: {
          drop_console: mode === 'production',
          drop_debugger: mode === 'production'
        }
      }
    },
    
    // Path resolution
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    
    // Environment variables
    define: {
      __APP_VERSION__: JSON.stringify(process.env.npm_package_version || '1.0.0'),
      __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
    },
    
    // CSS configuration
    css: {
      postcss: './postcss.config.mjs',
    },
    
    // Preview server (for production builds)
    preview: {
      port: 3000,
      host: true,
      cors: true
    },
    
    // Plugin configuration
    plugins: [htmlTemplatePlugin()],
    
    // Asset optimization
    assetsInclude: ['**/*.html'], // Include HTML files as assets
    
    // Dependency optimization
    optimizeDeps: {
      include: ['alpinejs'],
      exclude: []
    }
  }
}) 
