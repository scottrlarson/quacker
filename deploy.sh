#!/bin/bash

# Define variables
LOCAL_BINARY="./quacker"
REMOTE_USER="mreider"
REMOTE_HOST="quacker.eu"
REMOTE_BINARY="/usr/local/bin/quacker"
REMOTE_SERVICE="quacker"

# Step 1: Build the binary
echo "Building the binary..."
if ! GOOS=linux GOARCH=amd64 go build -o "$LOCAL_BINARY"; then
    echo "Error: Failed to build the binary. Aborting deployment."
    exit 1
fi
echo "Build succeeded."

# Step 2: Upload the binary
echo "Uploading binary to $REMOTE_USER@$REMOTE_HOST..."
if ! scp "$LOCAL_BINARY" "$REMOTE_USER@$REMOTE_HOST:/tmp/quacker"; then
    echo "Error: Failed to upload the binary. Aborting deployment."
    exit 1
fi

# Step 3: Move the binary and restart the service
echo "Moving binary to $REMOTE_BINARY and restarting the service..."
if ! ssh "$REMOTE_USER@$REMOTE_HOST" << EOF
    sudo mv /tmp/quacker $REMOTE_BINARY
    sudo chmod +x $REMOTE_BINARY
    sudo systemctl restart $REMOTE_SERVICE
EOF
then
    echo "Error: Failed to deploy or restart the service. Aborting."
    exit 1
fi

echo "Deployment and service restart successful!"
