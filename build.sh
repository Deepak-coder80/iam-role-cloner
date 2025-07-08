#!/bin/bash

# Build script for IAM Role Cloner
# This script builds binaries for multiple platforms

set -e

APP_NAME="iam-role-cloner"
VERSION=${VERSION:-"1.0.0"}
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "dev")}
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION=$(go version | awk '{print $3}')

# Create build directory
BUILD_DIR="build"
mkdir -p $BUILD_DIR

echo "🚀 Building $APP_NAME v$VERSION"
echo "=================================="
echo "Git Commit: $GIT_COMMIT"
echo "Build Date: $BUILD_DATE"
echo "Go Version: $GO_VERSION"
echo ""

# Set ldflags for version information
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X 'iam-role-cloner/cmd.Version=$VERSION'"
LDFLAGS="$LDFLAGS -X 'iam-role-cloner/cmd.GitCommit=$GIT_COMMIT'"
LDFLAGS="$LDFLAGS -X 'iam-role-cloner/cmd.BuildDate=$BUILD_DATE'"

# Build for different platforms
echo "📦 Building binaries..."

# Linux AMD64
echo "  Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o $BUILD_DIR/${APP_NAME}-linux-amd64 .

# Linux ARM64
echo "  Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o $BUILD_DIR/${APP_NAME}-linux-arm64 .

# macOS AMD64
echo "  Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o $BUILD_DIR/${APP_NAME}-darwin-amd64 .

# macOS ARM64 (Apple Silicon)
echo "  Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o $BUILD_DIR/${APP_NAME}-darwin-arm64 .

# Windows AMD64
echo "  Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o $BUILD_DIR/${APP_NAME}-windows-amd64.exe .

# Windows ARM64
echo "  Building for Windows (arm64)..."
GOOS=windows GOARCH=arm64 go build -ldflags "$LDFLAGS" -o $BUILD_DIR/${APP_NAME}-windows-arm64.exe .

echo ""
echo "✅ Build completed! Binaries created in $BUILD_DIR/"
echo ""

# List the built files with sizes
echo "📊 Build artifacts:"
echo "=================="
ls -lh $BUILD_DIR/

echo ""
echo "🎯 Test your build:"
echo "=================="
echo "# Test local build:"
echo "./$BUILD_DIR/${APP_NAME}-$(go env GOOS)-$(go env GOARCH) version --detailed"
echo ""
echo "# Or test the current platform binary:"
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    echo "./$BUILD_DIR/${APP_NAME}-windows-amd64.exe version"
else
    echo "./$BUILD_DIR/${APP_NAME}-$(go env GOOS)-$(go env GOARCH) version"
fi
echo ""
echo "📤 Ready for distribution!"
