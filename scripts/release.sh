#!/bin/bash
set -e

# Get the current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$DIR/.."

# Default values
VERSION=${1:-$(grep 'Version =' internal/version/version.go | cut -d '"' -f 2)}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "==============================================="
echo "Harness Template Documentation Generator Release"
echo "==============================================="
echo "Version: $VERSION"
echo "Commit: $COMMIT"
echo "Build Date: $BUILD_DATE"
echo "==============================================="
echo ""

# Confirm with user
read -p "Do you want to proceed with this release? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Release cancelled."
    exit 1
fi

# 1. Build local binaries
echo "Building local binaries..."
make clean
make build
echo "Local build completed."
echo ""

# 2. Run tests
echo "Running tests..."
make test
echo "Tests completed."
echo ""

# 3. Generate documentation
echo "Generating documentation..."
make generate
echo "Documentation generated."
echo ""

# 4. Build Docker images for all architectures
echo "Building multi-architecture Docker images..."
echo "This will push images to Docker Hub. Make sure you're logged in with 'docker login'."
read -p "Continue with Docker image build and push? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Docker build skipped."
else
    # Update the version in the Makefile if needed
    MAKEFILE_VERSION=$(grep 'DOCKER_TAG=' Makefile | cut -d'=' -f2)
    if [ "$MAKEFILE_VERSION" != "$VERSION" ]; then
        echo "Updating DOCKER_TAG in Makefile from $MAKEFILE_VERSION to $VERSION"
        sed -i "s/DOCKER_TAG=.*/DOCKER_TAG=$VERSION/" Makefile
    fi
    
    # Build and push all Docker images
    make docker-release-all
    echo "All Docker images built and pushed."
fi
echo ""

# 5. Create tag in git
echo "Creating git tag v$VERSION..."
read -p "Create and push git tag? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Git tag creation skipped."
else
    git tag -a "v$VERSION" -m "Release version $VERSION"
    git push origin "v$VERSION"
    echo "Git tag v$VERSION created and pushed."
fi

echo ""
echo "==============================================="
echo "Release $VERSION completed!"
echo "===============================================" 