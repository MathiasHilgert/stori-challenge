#!/bin/sh

echo "=========================================="
echo "Installing AWS CLI in Alpine Linux..."
echo "=========================================="

# Update package index
echo "Updating package index..."
apk update

# Install required packages
echo "Installing required packages..."
apk add --no-cache curl unzip python3 py3-pip jq

# Install AWS CLI using virtual environment (workaround for externally managed Python)
echo "Installing AWS CLI using virtual environment..."
python3 -m venv /opt/aws-cli-venv
. /opt/aws-cli-venv/bin/activate
pip install awscli

# Create symlink to make aws command available globally
ln -sf /opt/aws-cli-venv/bin/aws /usr/local/bin/aws

# Configure AWS CLI for LocalStack
echo "Configuring AWS CLI for LocalStack..."
/opt/aws-cli-venv/bin/aws configure set aws_access_key_id test
/opt/aws-cli-venv/bin/aws configure set aws_secret_access_key test
/opt/aws-cli-venv/bin/aws configure set default.region us-east-1

# Display AWS CLI version
echo "=========================================="
echo "AWS CLI installed successfully!"
/opt/aws-cli-venv/bin/aws --version
echo "=========================================="

# Display usage instructions
echo "Container ready!"
echo ""
echo "To use AWS CLI with LocalStack:"
echo "  aws --endpoint-url=http://stori-challenge-localstack:4566 s3 ls"
echo ""
echo "To access this container:"
echo "  docker exec -it stori-challenge-terminal sh"
echo "=========================================="

# Populate secrets if the file exists and CloudFormation is ready
if [ -f "/secrets/localstack-secrets.json" ]; then
    echo "Secrets file found, will populate after CloudFormation deployment..."
fi

# Create a marker file to indicate setup is complete
touch /tmp/aws-setup-complete

# Keep container running
tail -f /dev/null
