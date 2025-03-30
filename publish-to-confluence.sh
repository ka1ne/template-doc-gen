#!/bin/bash
# Script to simplify running the template documentation generator with Confluence publishing

# Load environment variables from .env if it exists
if [ -f .env ]; then
    echo "Loading configuration from .env file..."
    export $(grep -v '^#' .env | xargs)
else
    echo "No .env file found. Please create one based on .env.example"
    exit 1
fi

# Check required Confluence variables
if [ -z "$CONFLUENCE_URL" ] || [ -z "$CONFLUENCE_USERNAME" ] || [ -z "$CONFLUENCE_API_TOKEN" ] || [ -z "$CONFLUENCE_SPACE_KEY" ] || [ -z "$CONFLUENCE_PARENT_PAGE_ID" ]; then
    echo "Error: Missing required Confluence configuration!"
    echo "Please ensure your .env file contains:"
    echo "  CONFLUENCE_URL"
    echo "  CONFLUENCE_USERNAME"
    echo "  CONFLUENCE_API_TOKEN"
    echo "  CONFLUENCE_SPACE_KEY"
    echo "  CONFLUENCE_PARENT_PAGE_ID"
    exit 1
fi

# Default values for optional parameters
SOURCE_DIR=${SOURCE_DIR:-templates}
OUTPUT_DIR=${OUTPUT_DIR:-docs/output}
FORMAT=${FORMAT:-html}
VERBOSE=${VERBOSE:-true}
PUBLISH=${PUBLISH:-true}
VALIDATE_ONLY=${VALIDATE_ONLY:-false}

# Build the command
CMD="python process_template.py --source $SOURCE_DIR --output $OUTPUT_DIR --format $FORMAT"

# Add optional flags
if [ "$VERBOSE" = "true" ]; then
    CMD="$CMD --verbose"
fi

if [ "$PUBLISH" = "true" ]; then
    CMD="$CMD --publish"
    CMD="$CMD --confluence-url $CONFLUENCE_URL"
    CMD="$CMD --confluence-username $CONFLUENCE_USERNAME"
    CMD="$CMD --confluence-token $CONFLUENCE_API_TOKEN"
    CMD="$CMD --confluence-space $CONFLUENCE_SPACE_KEY"
    CMD="$CMD --confluence-parent-id $CONFLUENCE_PARENT_PAGE_ID"
fi

if [ "$VALIDATE_ONLY" = "true" ]; then
    CMD="$CMD --validate"
fi

# Create output directory if it doesn't exist
mkdir -p $OUTPUT_DIR

# Execute the command
echo "Starting documentation generation..."
echo "Command: $CMD"
eval $CMD

# Check execution status
if [ $? -eq 0 ]; then
    echo "Documentation generation completed successfully!"
    if [ "$PUBLISH" = "true" ]; then
        echo "Published to Confluence at: $CONFLUENCE_URL/spaces/$CONFLUENCE_SPACE_KEY"
    else
        echo "Documentation available at: $OUTPUT_DIR"
    fi
else
    echo "Documentation generation failed. See errors above."
    exit 1
fi 