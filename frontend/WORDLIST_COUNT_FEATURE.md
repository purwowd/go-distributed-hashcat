# Wordlist Count Feature

## Overview
Fitur ini menampilkan total dictionary (jumlah kata) dari wordlist yang dipilih pada kolom Attack Mode di modal Job Creation.

## Perubahan yang Dibuat

### 1. HTML Changes (job-modal.html)
- Mengubah span yang menampilkan "0" menjadi dinamis menggunakan Alpine.js computed property
- Dari: `<span class="text-xs text-gray-500 bg-green-100 px-2 py-1 rounded">0</span>`
- Menjadi: `<span class="text-xs text-gray-500 bg-green-100 px-2 py-1 rounded" x-text="selectedWordlistCount"></span>`

### 2. JavaScript Changes (main.ts)
- Menambahkan computed property `selectedWordlistCount` yang otomatis menghitung jumlah kata dari wordlist yang dipilih
- Property ini reactive terhadap perubahan `jobForm.wordlist_id`

## Cara Kerja

### Computed Property `selectedWordlistCount`
```typescript
get selectedWordlistCount() {
    if (!this.jobForm.wordlist_id) {
        return '0'
    }
    
    const selectedWordlist = this.wordlists.find((w: any) => w.id === this.jobForm.wordlist_id)
    if (!selectedWordlist) {
        return '0'
    }
    
    // Return word count if available, otherwise return size in KB
    if (selectedWordlist.word_count && selectedWordlist.word_count > 0) {
        return selectedWordlist.word_count.toLocaleString()
    } else if (selectedWordlist.size) {
        const sizeKB = Math.round(selectedWordlist.size / 1024)
        return `${sizeKB} KB`
    }
    
    return '0'
}
```

### Logic
1. **Jika tidak ada wordlist yang dipilih**: menampilkan "0"
2. **Jika wordlist memiliki word_count**: menampilkan jumlah kata dengan format yang mudah dibaca (contoh: "14,344,384")
3. **Jika wordlist tidak memiliki word_count**: menampilkan ukuran file dalam KB (contoh: "50 KB")
4. **Fallback**: menampilkan "0" jika tidak ada data yang tersedia

## Data Structure
Wordlist object harus memiliki struktur:
```typescript
interface Wordlist {
    id: string
    name: string
    orig_name: string
    size: number
    word_count?: number  // Optional field for word count
    path?: string
    created_at: string
}
```

## Testing
File test tersedia di `test_wordlist_count.html` untuk memverifikasi fungsi berjalan dengan benar.

## Contoh Output
- **rockyou.txt**: "14,344,384" (jika word_count tersedia)
- **10k-most-common.txt**: "10,000" (jika word_count tersedia)
- **custom-wordlist.txt**: "50 KB" (jika hanya size yang tersedia)
- **Tidak ada wordlist**: "0"

## Keuntungan
1. **Real-time updates**: Jumlah kata otomatis berubah ketika wordlist berbeda dipilih
2. **Fallback handling**: Tetap menampilkan informasi berguna meskipun word_count tidak tersedia
3. **User-friendly**: Format angka yang mudah dibaca dengan pemisah ribuan
4. **Performance**: Menggunakan computed property Alpine.js yang efisien
