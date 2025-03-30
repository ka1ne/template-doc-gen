#!/bin/bash
# Script to generate HTML documentation from Harness templates and serve it locally

# Load environment variables from .env if it exists
if [ -f .env ]; then
    echo "Loading configuration from .env file..."
    export $(grep -v '^#' .env | xargs)
else
    echo "No .env file found. Using default settings."
    # Set default values
    SOURCE_DIR="templates"
    OUTPUT_DIR="docs/output"
    FORMAT="html"
    VERBOSE="false"
    VALIDATE_ONLY="false"
fi

# Create output directory if it doesn't exist
mkdir -p $OUTPUT_DIR

# Generate documentation
echo "Generating documentation from templates in $SOURCE_DIR..."
CMD="python process_template.py --source $SOURCE_DIR --output $OUTPUT_DIR --format $FORMAT"

if [ "$VERBOSE" = "true" ]; then
    CMD="$CMD --verbose"
fi

if [ "$VALIDATE_ONLY" = "true" ]; then
    CMD="$CMD --validate"
fi

# Execute the command
echo "Executing: $CMD"
eval $CMD

# Check execution status
if [ $? -eq 0 ]; then
    echo "Documentation generation completed successfully!"
    echo "Documentation available at: $OUTPUT_DIR"
    
    # Serve the documentation locally with Python's HTTP server
    echo "Starting local server to preview documentation..."
    echo "Open your browser and navigate to http://localhost:8000/"
    
    # Check if we should serve the files
    if [ "$VALIDATE_ONLY" != "true" ]; then
        # Change to output directory and start server
        cd $OUTPUT_DIR
        python -m http.server 8000
    fi
else
    echo "Documentation generation failed. See errors above."
    exit 1
fi 