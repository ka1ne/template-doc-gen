#!/bin/bash
# Script to test Confluence API access by creating a simple page

# Load variables from .env file
source .env

# Build the JSON payload
read -r -d '' PAYLOAD << EOM
{
    "type": "page",
    "title": "Test API Page - $(date '+%Y-%m-%d %H:%M:%S')",
    "space": {
        "key": "$CONFLUENCE_SPACE_KEY"
    },
    "body": {
        "storage": {
            "value": "<h1>Test API Access</h1><p>This is a test page created via the Confluence REST API at $(date '+%Y-%m-%d %H:%M:%S')</p>",
            "representation": "storage"
        }
    },
    "ancestors": [
        {
            "id": "$CONFLUENCE_PARENT_PAGE_ID"
        }
    ]
}
EOM

echo "Testing Confluence API access..."
echo "URL: $CONFLUENCE_URL"
echo "Username: $CONFLUENCE_USERNAME"
echo "Space Key: $CONFLUENCE_SPACE_KEY"
echo "Parent Page ID: $CONFLUENCE_PARENT_PAGE_ID"
echo

# Make the API call
curl -s -X POST \
  -u "$CONFLUENCE_USERNAME:$CONFLUENCE_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD" \
  "$CONFLUENCE_URL/wiki/rest/api/content" | python3 -m json.tool

echo
echo "If you see a JSON response with an ID, the page was created successfully."
echo "Check your Confluence space to see the new page." 