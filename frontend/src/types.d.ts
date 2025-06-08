// Alpine.js type declarations
declare module 'alpinejs' {
  interface Alpine {
    data(name: string, callback: () => any): void
    start(): void
  }
  
  const Alpine: Alpine
  export default Alpine
}

// Vite environment variables
interface ImportMetaEnv {
  readonly VITE_API_URL?: string
  readonly VITE_APP_TITLE?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
} 
