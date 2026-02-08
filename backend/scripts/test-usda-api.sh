#!/bin/bash
# Script to test USDA API key validity

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if API key is set
if [ -z "$MACROLENS_USDA_API_KEY" ]; then
    echo "❌ ERROR: MACROLENS_USDA_API_KEY is not set in .env file"
    exit 1
fi

echo "🔑 Testing USDA API key: ${MACROLENS_USDA_API_KEY:0:8}..."
echo ""

# Test API with a simple search
API_KEY="$MACROLENS_USDA_API_KEY"
BASE_URL="${MACROLENS_USDA_BASE_URL:-https://api.nal.usda.gov/fdc}"

echo "🌐 Making test request to USDA API..."
echo "   URL: $BASE_URL/v1/foods/search?query=milk&api_key=***"
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/foods/search?query=milk&api_key=$API_KEY&pageSize=1")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "📊 Response Code: $HTTP_CODE"
echo ""

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ SUCCESS! API key is valid and working"
    echo ""
    echo "Sample response:"
    echo "$BODY" | head -20
elif [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "401" ]; then
    echo "❌ FAILED! API key is invalid or unauthorized"
    echo ""
    echo "Error response:"
    echo "$BODY"
    exit 1
elif [ "$HTTP_CODE" = "429" ]; then
    echo "⚠️  WARNING! Rate limit exceeded"
    echo ""
    echo "You've hit the USDA API rate limit (1000 requests/hour)"
    echo "Wait a bit and try again"
    exit 1
else
    echo "❌ FAILED! Unexpected error (HTTP $HTTP_CODE)"
    echo ""
    echo "Error response:"
    echo "$BODY"
    exit 1
fi
