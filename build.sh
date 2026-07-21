#!/bin/bash

set -e

echo "🔨 Building ytac (Youtube Audio in Console)..."

# Create bin directory
mkdir -p bin

# Build the binary
CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/ytac ./cmd/ytac

echo "✅ Build complete!"
echo "📦 Binary: bin/ytac"
echo ""
echo "Usage:"
echo "  bin/ytac <search terms>    Search and play"
echo "  bin/ytac <url>             Play from URL"
echo "  bin/ytac --help            Show help"
