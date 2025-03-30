#!/bin/bash
# Script to run the template documentation generator in Docker

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed or not in your PATH"
    exit 1
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo "No .env file found. Creating one from .env.example..."
    if [ -f .env.example ]; then
        cp .env.example .env
        echo "Created .env file. Please edit it with your settings before continuing."
        exit 1
    else
        echo "Error: .env.example file not found. Cannot create .env file."
        exit 1
    fi
fi

# Load environment variables from .env file
source .env || { echo "Error loading .env file"; exit 1; }

# Ensure output directory exists with proper permissions
mkdir -p "$OUTPUT_DIR" || { echo "Error creating output directory"; exit 1; }
chmod 777 "$OUTPUT_DIR" || { echo "Error setting directory permissions"; exit 1; }

# Get absolute paths for volume mounts
SOURCE_DIR=$(readlink -f "$SOURCE_DIR" || echo "templates")
OUTPUT_DIR=$(readlink -f "$OUTPUT_DIR" || echo "docs/output") 

echo "Starting Harness Template Documentation Generator..."
echo "Using Docker image: ka1ne/template-doc-gen:0.0.1-alpha"
echo "Source directory: $SOURCE_DIR"
echo "Output directory: $OUTPUT_DIR"
echo "Publishing to Confluence: $PUBLISH"
echo "Space Key: $CONFLUENCE_SPACE_KEY"
echo "Parent Page ID: $CONFLUENCE_PARENT_PAGE_ID"

# Build the Docker command with explicit flags
DOCKER_CMD="docker run --rm"
DOCKER_CMD="$DOCKER_CMD -v \"$SOURCE_DIR:/app/templates\" -v \"$OUTPUT_DIR:/app/docs\""
DOCKER_CMD="$DOCKER_CMD --env-file .env"
DOCKER_CMD="$DOCKER_CMD ka1ne/template-doc-gen:0.0.1-alpha"

# Add explicit flags based on .env settings
if [ "$VERBOSE" = "true" ]; then
    DOCKER_CMD="$DOCKER_CMD --verbose"
fi

if [ "$PUBLISH" = "true" ]; then
    DOCKER_CMD="$DOCKER_CMD --publish"
    DOCKER_CMD="$DOCKER_CMD --confluence-url \"$CONFLUENCE_URL\""
    DOCKER_CMD="$DOCKER_CMD --confluence-username \"$CONFLUENCE_USERNAME\""
    DOCKER_CMD="$DOCKER_CMD --confluence-token \"$CONFLUENCE_API_TOKEN\""
    DOCKER_CMD="$DOCKER_CMD --confluence-space \"$CONFLUENCE_SPACE_KEY\""
    DOCKER_CMD="$DOCKER_CMD --confluence-parent-id \"$CONFLUENCE_PARENT_PAGE_ID\""
fi

# Run the Docker container with explicit parameters
echo "Running command: $DOCKER_CMD"
eval "$DOCKER_CMD"

# Check if the command was successful
if [ $? -eq 0 ]; then
    echo "Documentation generated successfully!"
    echo "Output available at: $OUTPUT_DIR"
    
    # If publishing was enabled, show confirmation
    if [ "$PUBLISH" = "true" ]; then
        echo "Published to Confluence: $CONFLUENCE_URL/wiki/spaces/$CONFLUENCE_SPACE_KEY/pages/$CONFLUENCE_PARENT_PAGE_ID"
        echo "Check your Confluence page to see the updated documentation."
    fi
else
    echo "Error: Documentation generation failed."
    exit 1
fi 