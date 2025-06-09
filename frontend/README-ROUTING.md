# 🔗 Frontend Routing System

Client-side routing implementation untuk Hashcat Dashboard menggunakan Hash-based routing dengan Alpine.js integration.

## 📍 **Available Routes**

| Route | URL | Description |
|-------|-----|-------------|
| Overview | `/` or `/#overview` | Dashboard overview dengan statistics |
| Agents | `/#agents` | Manage dan monitor agents |
| Jobs | `/#jobs` | Create dan monitor cracking jobs |
| Files | `/#files` | Upload dan manage hash files |
| Wordlists | `/#wordlists` | Manage wordlists dan dictionaries |
| API Docs | `/#docs` | REST API documentation |

## 🏗️ **Architecture**

### **Router Class** (`utils/router.ts`)
- Singleton pattern untuk global state management
- Hash-based routing menggunakan `window.location.hash`
- Browser back/forward navigation support
- Event-driven route changes dengan subscription pattern

### **Alpine.js Integration** (`main.ts`)
- Router state sync dengan Alpine data
- Automatic tab switching berdasarkan URL
- Route-aware navigation components

### **Navigation Components**
- **Navigation Bar**: Smart active states berdasarkan current route
- **Breadcrumbs**: Contextual navigation dengan quick actions
- **SEO Integration**: Dynamic page titles dan meta tags

## 🎯 **Features**

### ✅ **URL Sharing & Bookmarking**
```javascript
// Users dapat bookmark atau share URL langsung
https://dashboard.local/#agents  // Langsung ke Agents page
https://dashboard.local/#jobs    // Langsung ke Jobs page
```

### ✅ **Browser Navigation**
- ✅ Back/Forward button support
- ✅ Refresh-safe routing (maintains current route)
- ✅ Deep linking support

### ✅ **SEO Enhancement**
```javascript
// Dynamic page titles
"Overview - Hashcat Dashboard"
"Agents - Hashcat Dashboard"
"Jobs - Hashcat Dashboard"

// Dynamic meta descriptions
// Automatic OpenGraph updates
```

### ✅ **User Experience**
- ✅ Active navigation indicators
- ✅ Breadcrumb navigation dengan contextual info
- ✅ Smooth transitions antar routes
- ✅ Copy URL functionality

## 🔧 **Implementation**

### **Router Usage**
```typescript
import { router } from './utils/router'

// Navigate programmatically
router.navigate('agents')

// Get current route
const currentRoute = router.getCurrentRoute()

// Subscribe to route changes
router.subscribe((route: string) => {
    console.log('Route changed to:', route)
})

// Check if route is current
if (router.isCurrentRoute('jobs')) {
    // Do something for jobs page
}
```

### **Alpine.js Integration**
```html
<!-- Navigation dengan proper URLs -->
<a :href="getRouteUrl('agents')" 
   @click.prevent="switchTab('agents')"
   :class="isCurrentRoute('agents') ? 'nav-item-active' : 'nav-item'">
    Agents
</a>

<!-- Route-specific content -->
<div x-show="currentTab === 'agents'">
    <!-- Agents content -->
</div>
```

### **Breadcrumb Components**
```html
<!-- Automatic breadcrumb dengan contextual info -->
<nav x-show="currentTab !== 'overview'">
    <a :href="getRouteUrl('overview')">Dashboard</a>
    → <span x-text="currentTab"></span>
    
    <!-- Dynamic stats per route -->
    <span x-text="`${agents.length} agents, ${onlineAgents.length} online`"></span>
</nav>
```

## 📊 **Benefits**

### **User Experience**
- ✅ **Shareable URLs**: Users dapat share direct links ke specific pages
- ✅ **Bookmarkable**: Browser bookmarks work properly  
- ✅ **Navigation Intuitive**: Back/forward buttons work as expected
- ✅ **Deep Linking**: External links dapat point ke specific tabs

### **SEO & Accessibility**
- ✅ **Search Engine Friendly**: Proper page titles dan meta tags
- ✅ **Screen Reader Support**: Semantic navigation structure
- ✅ **Progressive Enhancement**: Works tanpa JavaScript (fallback)

### **Developer Experience**
- ✅ **Type Safe**: Full TypeScript support
- ✅ **Event Driven**: Clean subscription pattern
- ✅ **Modular**: Easy to extend dengan new routes
- ✅ **Performance**: Hash routing = no server requests

## 🎨 **Styling**

Router menggunakan modern CSS dengan:
- ✅ Smooth transitions antar routes
- ✅ Active state indicators
- ✅ Glassmorphism design untuk breadcrumbs
- ✅ Responsive design untuk mobile/desktop

## 🔮 **Future Enhancements**

Potensial improvements:
- [ ] **Query Parameters**: Support untuk filters dan pagination
- [ ] **Route Guards**: Authentication checks per route
- [ ] **Lazy Loading**: Dynamic component loading per route
- [ ] **Analytics**: Route change tracking
- [ ] **History API**: Upgrade dari hash routing ke proper History API

## 🚀 **Usage Examples**

```typescript
// Programmatic navigation
switchTab('agents')  // Updates URL to /#agents

// URL sharing
copyToClipboard(window.location.href)  // Copy current URL

// Route-specific logic
if (isCurrentRoute('jobs')) {
    // Load job-specific data
    await loadJobs()
}

// Browser integration
window.addEventListener('popstate', () => {
    // Automatically handled by router
})
```

---

**Result**: Sekarang setiap menu punya URL tersendiri, users dapat bookmark, share, dan navigate dengan browser buttons! 🎉 
