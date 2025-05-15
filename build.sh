#!/bin/bash

# Exit on error
set -e

# Get the current version from the module path
VERSION=$(go list -m -f '{{.Version}}' 2>/dev/null || echo "v0.0.0")
if [ "$VERSION" = "" ] || [ "$VERSION" = "none" ]; then
    VERSION="v0.0.0"
fi

# Clean up old builds
rm -f taskmasterra-*

# Build for macOS (amd64 and arm64)
for GOOS in darwin; do
    for GOARCH in amd64 arm64; do
        echo "Building for $GOOS/$GOARCH..."
        GOOS=$GOOS GOARCH=$GOARCH go build -o "taskmasterra-$GOOS-$GOARCH"
    done
done

# Build for Linux (amd64)
for GOOS in linux; do
    for GOARCH in amd64; do
        echo "Building for $GOOS/$GOARCH..."
        GOOS=$GOOS GOARCH=$GOARCH go build -o "taskmasterra-$GOOS-$GOARCH"
    done
done

echo "Build complete. The following binaries were created:"
ls -la taskmasterra-*
