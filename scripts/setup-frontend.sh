#!/bin/bash

# Frontend Setup Script for Distributed Hashcat Dashboard
# This script sets up the modern TypeScript frontend

set -e

FRONTEND_DIR="frontend"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🌐 Setting up Hashcat Dashboard Frontend..."
echo "📁 Frontend directory: $PROJECT_ROOT/$FRONTEND_DIR"

# Check if we're in the right directory
if [ ! -d "$PROJECT_ROOT/$FRONTEND_DIR" ]; then
    echo "❌ Frontend directory not found. Are you in the project root?"
    exit 1
fi

cd "$PROJECT_ROOT/$FRONTEND_DIR"

# Check Node.js version
echo "🔍 Checking Node.js version..."
if ! command -v node &> /dev/null; then
    echo "❌ Node.js not found. Please install Node.js 18+ first."
    echo "   Visit: https://nodejs.org/"
    exit 1
fi

NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo "❌ Node.js $NODE_VERSION found, but Node.js 18+ is required."
    echo "   Current version: $(node -v)"
    echo "   Please upgrade: https://nodejs.org/"
    exit 1
fi

echo "✅ Node.js $(node -v) detected"

# Check npm
if ! command -v npm &> /dev/null; then
    echo "❌ npm not found. Please install npm first."
    exit 1
fi

echo "✅ npm $(npm -v) detected"

# Install dependencies
echo "📦 Installing dependencies..."
npm install

# Check if backend is running
echo "🔍 Checking backend connectivity..."
if curl -s http://localhost:1337/health > /dev/null; then
    echo "✅ Backend is running on port 1337"
    BACKEND_RUNNING=true
else
    echo "⚠️  Backend not running on port 1337"
    echo "   Start backend: make run-server (or ./bin/server)"
    BACKEND_RUNNING=false
fi

# Build for production
echo "🔨 Building for production..."
npm run build

echo ""
echo "🎯 Setup Complete!"
echo "=================="
echo ""

if [ "$BACKEND_RUNNING" = true ]; then
    echo "🚀 Ready to start development:"
    echo "   npm run dev"
    echo ""
    echo "🌐 Access points:"
    echo "   - Frontend: http://localhost:3000"
    echo "   - Backend API: http://localhost:1337"
    echo "   - Health check: http://localhost:1337/health"
else
    echo "🔧 Start backend first:"
    echo "   cd .. && make run-server"
    echo ""
    echo "🚀 Then start frontend:"
    echo "   npm run dev"
fi

echo ""
echo "📊 Build output:"
ls -la dist/ | grep -E '\.(js|css|html)$' | awk '{print "   - "$9": "$5" bytes"}'

echo ""
echo "🛠️ Available commands:"
echo "   npm run dev      - Start development server"
echo "   npm run build    - Build for production"
echo "   npm run preview  - Preview production build"
echo ""
echo "📖 Documentation:"
echo "   - Frontend README: ./README.md"
echo "   - Main README: ../README.md"
echo ""
echo "✅ Frontend setup completed successfully!" 
