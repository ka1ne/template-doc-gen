#!/bin/bash
set -e

# Load environment variables from .env file if it exists
if [ -f .env ]; then
  echo "Loading environment variables from .env file"
  export $(grep -v '^#' .env | xargs)
fi

# Default values
SOURCE_DIR=${SOURCE_DIR:-"templates"}
OUTPUT_DIR=${OUTPUT_DIR:-"docs/templates"}
FORMAT=${FORMAT:-"html"}
VERBOSE=${VERBOSE:-"true"}

# Check for required Confluence variables if publishing
if [ "$PUBLISH" = "true" ]; then
  if [ -z "$CONFLUENCE_URL" ] || [ -z "$CONFLUENCE_USERNAME" ] || [ -z "$CONFLUENCE_API_TOKEN" ] || [ -z "$CONFLUENCE_SPACE_KEY" ] || [ -z "$CONFLUENCE_PARENT_PAGE_ID" ]; then
    echo "Error: Missing required Confluence configuration. Please set the following environment variables:"
    echo "  CONFLUENCE_URL"
    echo "  CONFLUENCE_USERNAME"
    echo "  CONFLUENCE_API_TOKEN"
    echo "  CONFLUENCE_SPACE_KEY"
    echo "  CONFLUENCE_PARENT_PAGE_ID"
    exit 1
  fi
fi

# Build command
CMD="python process_template.py --source $SOURCE_DIR --output $OUTPUT_DIR --format $FORMAT"

# Add optional flags
if [ "$VERBOSE" = "true" ]; then
  CMD="$CMD --verbose"
fi

if [ "$VALIDATE_ONLY" = "true" ]; then
  CMD="$CMD --validate"
fi

if [ "$PUBLISH" = "true" ]; then
  CMD="$CMD --publish \
    --confluence-url $CONFLUENCE_URL \
    --confluence-username $CONFLUENCE_USERNAME \
    --confluence-token $CONFLUENCE_API_TOKEN \
    --confluence-space $CONFLUENCE_SPACE_KEY \
    --confluence-parent-id $CONFLUENCE_PARENT_PAGE_ID"
fi

# Execute the command
echo "Running: $CMD"
eval $CMD

echo "Documentation generation complete!" 