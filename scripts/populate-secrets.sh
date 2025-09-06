#!/bin/sh

echo "=========================================="
echo "Populating Secrets Manager with individual secrets"
echo "=========================================="

SECRETS_FILE="/secrets/localstack-secrets.json"
ENDPOINT_URL="http://stori-challenge-localstack:4566"
REGION="us-east-1"

# Check if secrets file exists
if [ ! -f "$SECRETS_FILE" ]; then
    echo "Error: Secrets file not found at $SECRETS_FILE"
    exit 1
fi

# Check if jq is available for JSON parsing
if ! command -v jq > /dev/null 2>&1; then
    echo "Error: jq is not installed. Please install jq for JSON parsing."
    exit 1
fi

echo "Reading secrets from $SECRETS_FILE..."

# Validate JSON format
if ! jq empty "$SECRETS_FILE" > /dev/null 2>&1; then
    echo "Error: Invalid JSON format in $SECRETS_FILE"
    exit 1
fi

echo "Creating individual secrets dynamically..."

# Get all keys from JSON file
KEYS=$(jq -r 'keys[]' "$SECRETS_FILE")

# Counter for tracking progress
total_secrets=$(echo "$KEYS" | wc -l)
current=0

echo "Found $total_secrets secrets to process..."
echo ""

# Process each key dynamically
for key in $KEYS; do
    current=$((current + 1))
    
    # Extract value for this key
    value=$(jq -r ".\"$key\"" "$SECRETS_FILE")
    
    echo "[$current/$total_secrets] Updating secret: $key"
    
    # Try to update the secret
    aws --endpoint-url="$ENDPOINT_URL" secretsmanager update-secret \
        --secret-id "$key" \
        --secret-string "$value" \
        --region "$REGION" \
        > /dev/null 2>&1
    
    update_result=$?
    
    if [ $update_result -eq 0 ]; then
        echo "   $key secret updated successfully!"
    else
        echo "   $key secret update failed, trying to create it..."
        
        # If update failed, try to create the secret
        aws --endpoint-url="$ENDPOINT_URL" secretsmanager create-secret \
            --name "$key" \
            --secret-string "$value" \
            --region "$REGION" \
            > /dev/null 2>&1
        
        create_result=$?
        
        if [ $create_result -eq 0 ]; then
            echo "   $key secret created successfully!"
        else
            echo "   $key secret creation failed (might already exist or need CloudFormation deployment)"
        fi
    fi
    
    echo ""
done

echo "=========================================="
echo "   Dynamic secrets population completed!"
echo "   Processed $total_secrets secrets from JSON file"
echo "=========================================="
