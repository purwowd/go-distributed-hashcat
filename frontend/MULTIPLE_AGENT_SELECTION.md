# Multiple Agent Selection Feature

## Overview
Fitur ini memungkinkan user untuk memilih multiple agents pada form Create New Job, dengan opsi "Select All" untuk memilih semua agents yang tersedia. **Sekarang dengan scrolling yang optimal untuk handling banyak data.**

## Fitur yang Ditambahkan

### 1. Multiple Selection UI
- **Checkbox "Select All"**: Checkbox di atas daftar agents untuk memilih/deselect semua agents
- **Individual Checkboxes**: Setiap agent memiliki checkbox sendiri untuk selection individual
- **Visual Feedback**: Menampilkan jumlah agents yang dipilih
- **Compact Scrolling**: Area selection dengan fixed height dan smooth scrolling

### 2. Data Structure Changes
- **jobForm.agent_ids**: Array of strings untuk menyimpan multiple agent IDs
- **Validation**: Validasi bahwa minimal satu agent dipilih
- **Agent Status Check**: Memastikan semua selected agents online

### 3. Functions Added
- **toggleSelectAllAgents(checked)**: Toggle select/deselect semua agents
- **areAllAgentsSelected()**: Check apakah semua agents sudah dipilih

### 4. Scrolling Improvements
- **Fixed Height**: Area selection dibatasi maksimal 160px (max-h-40)
- **Custom Scrollbar**: Thin scrollbar dengan styling yang modern
- **Compact Layout**: Setiap agent item memiliki padding dan border yang optimal
- **Text Truncation**: Mencegah overflow text pada nama dan IP address

## Implementation Details

### HTML Changes (job-modal.html)
```html
<!-- Select All Checkbox -->
<div class="mb-2">
    <label class="flex items-center space-x-2 cursor-pointer">
        <input type="checkbox" 
               @change="toggleSelectAllAgents($event.target.checked)"
               :checked="areAllAgentsSelected()"
               class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 focus:ring-2">
        <span class="text-sm font-medium text-gray-700">Select All Available Agents</span>
    </label>
</div>

<!-- Multiple Agent Selection with Scrolling -->
<div class="max-h-40 overflow-y-auto border border-gray-200 rounded-lg bg-white scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100">
    <template x-for="agent in onlineAgents" :key="agent.id">
        <label class="flex items-center space-x-3 py-2 px-3 hover:bg-gray-50 rounded cursor-pointer border-b border-gray-100 last:border-b-0">
            <input type="checkbox" 
                   :value="agent.id"
                   x-model="jobForm.agent_ids"
                   class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 focus:ring-2 flex-shrink-0">
            <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between">
                    <span class="font-medium text-gray-900 truncate" x-text="agent.name"></span>
                    <span class="text-xs text-gray-500 bg-blue-100 px-2 py-1 rounded flex-shrink-0 ml-2" x-text="agent.capabilities || 'Unknown GPU'"></span>
                </div>
                <div class="text-sm text-gray-600 truncate" x-text="agent.ip_address"></div>
            </div>
        </label>
    </template>
</div>
```

### CSS Changes (main.css)
```css
/* Custom scrollbar styles */
@layer utilities {
  .scrollbar-thin {
    scrollbar-width: thin;
    scrollbar-color: #d1d5db #f3f4f6;
  }
  
  .scrollbar-thin::-webkit-scrollbar {
    width: 6px;
  }
  
  .scrollbar-thin::-webkit-scrollbar-track {
    background: #f3f4f6;
    border-radius: 3px;
  }
  
  .scrollbar-thin::-webkit-scrollbar-thumb {
    background: #d1d5db;
    border-radius: 3px;
  }
  
  .scrollbar-thin::-webkit-scrollbar-thumb:hover {
    background: #9ca3af;
  }
}
```

### JavaScript Changes (main.ts)
```typescript
// Data structure change
jobForm: { 
    name: '', 
    hash_file_id: '', 
    wordlist_id: '', 
    agent_ids: [] as string[], // Changed from agent_id: string
    hash_type: '', 
    attack_mode: '' 
}

// New functions
toggleSelectAllAgents(checked: boolean) {
    if (checked) {
        this.jobForm.agent_ids = this.onlineAgents.map((agent: any) => agent.id)
    } else {
        this.jobForm.agent_ids = []
    }
},

areAllAgentsSelected(): boolean {
    return this.onlineAgents.length > 0 && 
           this.jobForm.agent_ids.length === this.onlineAgents.length &&
           this.onlineAgents.every((agent: any) => this.jobForm.agent_ids.includes(agent.id))
}
```

## Scrolling Features

### 1. **Fixed Height Container**
- **max-h-40**: Maksimal tinggi 160px untuk area selection
- **overflow-y-auto**: Scrollbar otomatis muncul ketika konten melebihi tinggi
- **Responsive**: Tetap compact di berbagai ukuran layar

### 2. **Custom Scrollbar**
- **Thin Design**: Scrollbar tipis (6px) untuk tampilan yang clean
- **Hover Effects**: Scrollbar berubah warna saat hover
- **Cross-browser**: Support untuk WebKit dan Firefox

### 3. **Compact Layout**
- **Border Separation**: Setiap agent item dipisah dengan border tipis
- **Padding Optimization**: Padding yang optimal untuk readability
- **Text Truncation**: Mencegah overflow pada nama dan IP yang panjang

## Usage

### 1. Select All Agents
- Klik checkbox "Select All Available Agents" untuk memilih semua agents
- Checkbox akan otomatis checked ketika semua agents dipilih

### 2. Individual Selection
- Klik checkbox individual untuk memilih/deselect agent tertentu
- Checkbox "Select All" akan unchecked jika ada agent yang tidak dipilih

### 3. Scrolling
- Gunakan scrollbar atau mouse wheel untuk navigasi
- Area selection tetap compact meskipun ada banyak agents
- Smooth scrolling experience

### 4. Validation
- Minimal satu agent harus dipilih untuk membuat job
- Semua selected agents harus online
- Form tidak bisa di-submit tanpa agent selection

## Benefits

1. **Efficiency**: User bisa memilih multiple agents sekaligus
2. **Flexibility**: Bisa memilih subset dari available agents
3. **User Experience**: Interface yang intuitif dengan visual feedback
4. **Scalability**: Mendukung distributed computing dengan multiple agents
5. **Space Optimization**: Scrolling area yang compact untuk handling banyak data
6. **Modern UI**: Custom scrollbar dengan styling yang modern

## Testing

- `frontend/test_multiple_selection.html` - Basic multiple selection test
- `frontend/test_scrollable_agents.html` - Scrolling test dengan 25+ agents

## Backend Considerations

Pastikan backend API mendukung multiple agent assignment dengan field `agent_ids` array.

## Performance Notes

- Scrolling area dibatasi maksimal 160px untuk optimal performance
- Text truncation mencegah layout shift
- Custom scrollbar tidak mempengaruhi performance rendering
