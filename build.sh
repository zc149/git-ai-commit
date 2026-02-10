#!/bin/bash

# Git AI Commit - Cross-platform Build Script
# This script builds binaries for multiple platforms

set -e

VERSION=${VERSION:-"1.4.0"}
BUILD_DIR="dist"
BINARY_NAME="git-ai-commit"

echo "Building Git AI Commit v${VERSION}"
echo "====================================="

# Create dist directory
mkdir -p ${BUILD_DIR}

# Build for macOS (AMD64)
echo "Building for macOS (AMD64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 main.go
chmod +x ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64

# Build for macOS (ARM64/Apple Silicon)
echo "Building for macOS (ARM64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ${BUILD_DIR}/${BINARY_NAME}-darwin-arm64 main.go
chmod +x ${BUILD_DIR}/${BINARY_NAME}-darwin-arm64

# Build for Windows (AMD64)
echo "Building for Windows (AMD64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ${BUILD_DIR}/${BINARY_NAME}-windows-amd64.exe main.go

# Build for Linux (AMD64)
echo "Building for Linux (AMD64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 main.go
chmod +x ${BUILD_DIR}/${BINARY_NAME}-linux-amd64

echo ""
echo "Build complete! Binaries are in ${BUILD_DIR}/"
echo "====================================="
ls -lh ${BUILD_DIR}/