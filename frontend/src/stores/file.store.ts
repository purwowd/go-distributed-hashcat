// File Store for Hash Files State Management
import { apiService, type HashFile } from '@/services/api.service'

interface FileState {
    hashFiles: HashFile[]
    loading: boolean
    error: string | null
    lastUpdated: Date | null
}

class FileStore {
    private state: FileState = {
        hashFiles: [],
        loading: false,
        error: null,
        lastUpdated: null
    }

    private listeners: Set<() => void> = new Set()

    public getState(): FileState {
        return { ...this.state }
    }

    public subscribe(listener: () => void): () => void {
        this.listeners.add(listener)
        return () => this.listeners.delete(listener)
    }

    private notify(): void {
        this.listeners.forEach(listener => listener())
    }

    private setState(updates: Partial<FileState>): void {
        this.state = { ...this.state, ...updates }
        this.notify()
    }

    public actions = {
        fetchHashFiles: async (): Promise<void> => {
            this.setState({ loading: true, error: null })
            
            try {
                const hashFiles = await apiService.getHashFiles()
                this.setState({ 
                    hashFiles, 
                    loading: false, 
                    lastUpdated: new Date(),
                    error: null 
                })
            } catch (error) {
                this.setState({ 
                    loading: false, 
                    error: error instanceof Error ? error.message : 'Failed to fetch hash files' 
                })
            }
        },

        fetchHashFile: async (id: string): Promise<HashFile | null> => {
            try {
                const hashFile = await apiService.getHashFile(id)
                if (hashFile) {
                    const updatedFiles = this.state.hashFiles.map(f => 
                        f.id === hashFile.id ? hashFile : f
                    )
                    if (!updatedFiles.find(f => f.id === hashFile.id)) {
                        updatedFiles.push(hashFile)
                    }
                    this.setState({ hashFiles: updatedFiles })
                }
                return hashFile
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to fetch hash file' 
                })
                return null
            }
        },

        uploadHashFile: async (file: File): Promise<HashFile | null> => {
            this.setState({ loading: true, error: null })
            
            try {
                const newHashFile = await apiService.uploadHashFile(file)
                if (newHashFile) {
                    this.setState({ 
                        hashFiles: [...this.state.hashFiles, newHashFile],
                        loading: false 
                    })
                }
                return newHashFile
            } catch (error) {
                this.setState({ 
                    loading: false, 
                    error: error instanceof Error ? error.message : 'Failed to upload hash file' 
                })
                return null
            }
        },

        deleteHashFile: async (id: string): Promise<boolean> => {
            try {
                const success = await apiService.deleteHashFile(id)
                if (success) {
                    const updatedFiles = this.state.hashFiles.filter(file => file.id !== id)
                    this.setState({ hashFiles: updatedFiles })
                }
                return success
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to delete hash file' 
                })
                return false
            }
        },

        clearError: (): void => {
            this.setState({ error: null })
        },

        reset: (): void => {
            this.setState({
                hashFiles: [],
                loading: false,
                error: null,
                lastUpdated: null
            })
        }
    }

    public getters = {
        getHashFileById: (id: string): HashFile | undefined => {
            return this.state.hashFiles.find(file => file.id === id)
        },

        getHashFilesByName: (name: string): HashFile[] => {
            return this.state.hashFiles.filter(file => 
                file.name.toLowerCase().includes(name.toLowerCase())
            )
        },

        getTotalCount: (): number => {
            return this.state.hashFiles.length
        },

        getTotalSize: (): number => {
            return this.state.hashFiles.reduce((total, file) => total + file.size, 0)
        },

        getTotalHashCount: (): number => {
            // Note: hash_count field not available in current API response
            return 0
        }
    }
}

export const fileStore = new FileStore()
export type { FileState } 
