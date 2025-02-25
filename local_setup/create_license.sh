#!/bin/bash

# Configuration variables
KEYCLOAK_ADMIN_URL="http://localhost/keycloak/realms/prism-admin/protocol/openid-connect/token"
KEYCLOAK_USER_URL="http://localhost/keycloak/realms/prism-user/protocol/openid-connect/token"

LICENSE_URL="http://localhost/api/v1/license/create"
ACTIVATE_LICENSE_URL="http://localhost/api/v1/report/activate-license"

# Admin credentials for license creation
ADMIN_USERNAME="adminUser" # To Change
ADMIN_PASSWORD="password" # To Change
ADMIN_CLIENT_ID="prism-admin-login-client"

# User credentials for license activation
USER_USERNAME="pratik" # To Change
USER_PASSWORD="Password" # To Change
USER_CLIENT_ID="prism-user-login-client"

# License details
LICENSE_NAME="test-license"
LICENSE_EXPIRATION="2025-11-11T20:37:49.004638Z"

### Step 1: Obtain authentication token for Admin User (License Creation)
echo "Fetching authentication token for admin user..."
ADMIN_TOKEN_RESPONSE=$(curl -s -X POST "$KEYCLOAK_ADMIN_URL" \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=password" \
     -d "client_id=$ADMIN_CLIENT_ID" \
     -d "username=$ADMIN_USERNAME" \
     -d "password=$ADMIN_PASSWORD" \
     -d "scope=openid profile email roles")

echo "Response: $ADMIN_TOKEN_RESPONSE"

# Extract the access token
ADMIN_ACCESS_TOKEN=$(echo "$ADMIN_TOKEN_RESPONSE" | jq -r '.access_token')

if [ -z "$ADMIN_ACCESS_TOKEN" ] || [ "$ADMIN_ACCESS_TOKEN" == "null" ]; then
    echo "Error: Failed to retrieve admin access token!"
    exit 1
fi

echo "Admin access token obtained successfully."

### Step 2: Create a new license
echo "Creating a new license..."
LICENSE_RESPONSE=$(curl -s -X POST "$LICENSE_URL" \
     -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
     -H "Content-Type: application/json" \
     -d "{
         \"Name\": \"$LICENSE_NAME\",
         \"Expiration\": \"$LICENSE_EXPIRATION\"
     }")

# Extract License Key
LICENSE_KEY=$(echo "$LICENSE_RESPONSE" | jq -r '.License')

if [ -z "$LICENSE_KEY" ] || [ "$LICENSE_KEY" == "null" ]; then
    echo "Error: Failed to create license!"
    echo "Response: $LICENSE_RESPONSE"
    exit 1
fi

echo "License created successfully. License Key: $LICENSE_KEY"

### Step 3: Obtain authentication token for User (License Activation)
echo "Fetching authentication token for user: $USER_USERNAME..."
USER_TOKEN_RESPONSE=$(curl -s -X POST "$KEYCLOAK_USER_URL" \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=password" \
     -d "client_id=$USER_CLIENT_ID" \
     -d "username=$USER_USERNAME" \
     -d "password=$USER_PASSWORD" \
     -d "scope=openid profile email roles")

# Extract the access token
USER_ACCESS_TOKEN=$(echo "$USER_TOKEN_RESPONSE" | jq -r '.access_token')

if [ -z "$USER_ACCESS_TOKEN" ] || [ "$USER_ACCESS_TOKEN" == "null" ]; then
    echo "Error: Failed to retrieve user access token!"
    echo "Response: $USER_TOKEN_RESPONSE"
    exit 1
fi

echo "User access token obtained successfully."

### Step 4: Activate the License using User Credentials
echo "Activating the license..."
ACTIVATION_RESPONSE=$(curl -s -X POST "$ACTIVATE_LICENSE_URL" \
     -H "Authorization: Bearer $USER_ACCESS_TOKEN" \
     -H "Content-Type: application/json" \
     -d "{
           \"License\": \"$LICENSE_KEY\"
         }")


echo "License activated successfully."
