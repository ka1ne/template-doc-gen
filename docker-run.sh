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

# Clean up existing content in the output directory
echo "Cleaning up existing documentation in $OUTPUT_DIR..."
rm -rf "$OUTPUT_DIR"/* || { echo "Warning: Could not clean up output directory"; }

# Recreate the directory with proper permissions
mkdir -p "$OUTPUT_DIR" || { echo "Error creating output directory"; exit 1; }
chmod 777 "$OUTPUT_DIR" || { echo "Error setting directory permissions"; exit 1; }

# Get absolute paths for volume mounts
SOURCE_DIR=$(readlink -f "$SOURCE_DIR" || echo "templates")
OUTPUT_DIR=$(readlink -f "$OUTPUT_DIR" || echo "docs/output") 

echo "Starting Harness Template Documentation Generator..."
echo "Using Docker image: ka1ne/template-doc-gen:0.0.3-alpha"
echo "Source directory: $SOURCE_DIR"
echo "Output directory: $OUTPUT_DIR"

# Build the Docker command with explicit flags
DOCKER_CMD="docker run --rm"
DOCKER_CMD="$DOCKER_CMD -v \"$SOURCE_DIR:/app/templates\" -v \"$OUTPUT_DIR:/app/docs\""
DOCKER_CMD="$DOCKER_CMD --env-file .env"
DOCKER_CMD="$DOCKER_CMD ka1ne/template-doc-gen:0.0.3-alpha"

# Add explicit flags based on .env settings
if [ "$VERBOSE" = "true" ]; then
    DOCKER_CMD="$DOCKER_CMD --verbose"
fi

if [ "$VALIDATE_ONLY" = "true" ]; then
    DOCKER_CMD="$DOCKER_CMD --validate"
fi

# Run the Docker container with explicit parameters
echo "Running command: $DOCKER_CMD"
eval "$DOCKER_CMD"

# Check if the command was successful
if [ $? -eq 0 ]; then
    echo "Documentation generated successfully!"
    echo "Output available at: $OUTPUT_DIR"
    
    # Open the documentation in browser if possible
    if command -v xdg-open &> /dev/null; then
        xdg-open "$OUTPUT_DIR/index.html" &> /dev/null || true
    elif command -v open &> /dev/null; then
        open "$OUTPUT_DIR/index.html" &> /dev/null || true
    fi
else
    echo "Error: Documentation generation failed."
    exit 1
fi 