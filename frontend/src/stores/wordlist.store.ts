// Wordlist Store for State Management
import { apiService, type Wordlist } from '@/services/api.service'

interface WordlistState {
    wordlists: Wordlist[]
    loading: boolean
    error: string | null
    lastUpdated: Date | null
}

class WordlistStore {
    private state: WordlistState = {
        wordlists: [],
        loading: false,
        error: null,
        lastUpdated: null
    }

    private listeners: Set<() => void> = new Set()

    public getState(): WordlistState {
        return { ...this.state }
    }

    public subscribe(listener: () => void): () => void {
        this.listeners.add(listener)
        return () => this.listeners.delete(listener)
    }

    private notify(): void {
        this.listeners.forEach(listener => listener())
    }

    private setState(updates: Partial<WordlistState>): void {
        this.state = { ...this.state, ...updates }
        this.notify()
    }

    public actions = {
        fetchWordlists: async (): Promise<void> => {
            this.setState({ loading: true, error: null })
            
            try {
                const wordlists = await apiService.getWordlists()
                this.setState({ 
                    wordlists, 
                    loading: false, 
                    lastUpdated: new Date(),
                    error: null 
                })
            } catch (error) {
                this.setState({ 
                    loading: false, 
                    error: error instanceof Error ? error.message : 'Failed to fetch wordlists' 
                })
            }
        },

        fetchWordlist: async (id: string): Promise<Wordlist | null> => {
            try {
                const wordlist = await apiService.getWordlist(id)
                if (wordlist) {
                    const updatedWordlists = this.state.wordlists.map(w => 
                        w.id === wordlist.id ? wordlist : w
                    )
                    if (!updatedWordlists.find(w => w.id === wordlist.id)) {
                        updatedWordlists.push(wordlist)
                    }
                    this.setState({ wordlists: updatedWordlists })
                }
                return wordlist
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to fetch wordlist' 
                })
                return null
            }
        },

        uploadWordlist: async (file: File): Promise<Wordlist | null> => {
            this.setState({ loading: true, error: null })
            
            try {
                const newWordlist = await apiService.uploadWordlist(file)
                if (newWordlist) {
                    this.setState({ 
                        wordlists: [...this.state.wordlists, newWordlist],
                        loading: false 
                    })
                }
                return newWordlist
            } catch (error) {
                this.setState({ 
                    loading: false, 
                    error: error instanceof Error ? error.message : 'Failed to upload wordlist' 
                })
                return null
            }
        },

        deleteWordlist: async (id: string): Promise<boolean> => {
            try {
                const success = await apiService.deleteWordlist(id)
                if (success) {
                    const updatedWordlists = this.state.wordlists.filter(wordlist => wordlist.id !== id)
                    this.setState({ wordlists: updatedWordlists })
                }
                return success
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to delete wordlist' 
                })
                return false
            }
        },

        clearError: (): void => {
            this.setState({ error: null })
        },

        reset: (): void => {
            this.setState({
                wordlists: [],
                loading: false,
                error: null,
                lastUpdated: null
            })
        }
    }

    public getters = {
        getWordlistById: (id: string): Wordlist | undefined => {
            return this.state.wordlists.find(wordlist => wordlist.id === id)
        },

        getWordlistsByName: (name: string): Wordlist[] => {
            return this.state.wordlists.filter(wordlist => 
                wordlist.name.toLowerCase().includes(name.toLowerCase())
            )
        },

        getTotalCount: (): number => {
            return this.state.wordlists.length
        },

        getTotalSize: (): number => {
            return this.state.wordlists.reduce((total, wordlist) => total + wordlist.size, 0)
        },

        getTotalWordCount: (): number => {
            return this.state.wordlists.reduce((total, wordlist) => total + (wordlist.word_count || 0), 0)
        }
    }
}

export const wordlistStore = new WordlistStore()
export type { WordlistState } 
