#! /bin/bash

CLUSTER_DOMAIN=$(kubectl get ingresses.config.openshift.io cluster -o jsonpath='{.spec.domain}')
HOST="https://maas.${CLUSTER_DOMAIN}"

TOKEN_RESPONSE=$(curl -sSk \
  -H "Authorization: Bearer $(oc whoami -t)" \
  -H "Content-Type: application/json" \
  -X POST \
  -d '{"expiration": "10m"}' \
  "${HOST}/maas-api/v1/tokens")

TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r .token)

# Check if TOKEN is valid
if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
  echo "ERROR: Failed to obtain authentication token."
  echo "Token response: $TOKEN_RESPONSE"
  exit 1
fi

echo "Successfully obtained authentication token."

MODELS=$(curl -sSk "${HOST}/maas-api/v1/models" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" | jq -r .)

echo "$MODELS" | jq .

# Prompt user for index
TOTAL_MODELS=$(echo "$MODELS" | jq '.data | length')

# Check if we got valid models
if [[ -z "$TOTAL_MODELS" || "$TOTAL_MODELS" == "null" || "$TOTAL_MODELS" == "0" ]]; then
  echo "ERROR: No models available or failed to retrieve models."
  exit 1
fi

echo "There are ${TOTAL_MODELS} models available."

read -rp "Enter the array index to use (0 - $((TOTAL_MODELS-1))): " INDEX

# Validate numeric input
if ! [[ "$INDEX" =~ ^[0-9]+$ ]]; then
  echo "Invalid input. Must be a number."
  exit 1
fi

# Validate range
if (( INDEX < 0 || INDEX >= TOTAL_MODELS )); then
  echo "Index out of range."
  exit 1
fi

MODEL_NAME=$(echo "$MODELS" | jq -r ".data[$INDEX].id")
MODEL_URL=$(echo "$MODELS" | jq -r ".data[$INDEX].url")

echo "Selected Model Index: $INDEX"
echo "Model Name: $MODEL_NAME"
echo "Model URL:  $MODEL_URL"

# Attempt inference with the URL as provided
echo ""
echo "Testing inference with provided URL..."
RESPONSE=$(curl -sSk -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"model\": \"${MODEL_NAME}\", \"prompt\": \"Hello\", \"max_tokens\": 50}" \
  "${MODEL_URL}/v1/completions")

# Check if response is empty or contains connection errors
if [[ -z "$RESPONSE" || "$RESPONSE" == *"Could not resolve host"* || "$RESPONSE" == *"Connection refused"* || "$RESPONSE" == *"Empty reply"* ]]; then
  echo "⚠️  Empty or failed response detected."
  
  # Check if URL starts with http:// (not https://)
  if [[ "$MODEL_URL" == http://* ]]; then
    echo "📋 The model URL returned HTTP but testing with HTTPS as a fallback..."
    
    # Convert HTTP to HTTPS
    HTTPS_URL="${MODEL_URL/http:/https:}"
    echo "Retrying with: $HTTPS_URL"
    
    # Retry with HTTPS
    RESPONSE=$(curl -sSk -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"model\": \"${MODEL_NAME}\", \"prompt\": \"Hello\", \"max_tokens\": 50}" \
      "${HTTPS_URL}/v1/completions")
    
    if [[ -n "$RESPONSE" && "$RESPONSE" != *"Could not resolve host"* ]]; then
      echo ""
      echo "✅ Success with HTTPS! Response:"
      echo "$RESPONSE" | jq .
    else
      echo ""
      echo "❌ ERROR: Both HTTP and HTTPS requests failed."
      echo "Response: $RESPONSE"
      exit 1
    fi
  else
    echo ""
    echo "❌ ERROR: Request failed and URL is already HTTPS."
    echo "Response: $RESPONSE"
    exit 1
  fi
else
  echo ""
  echo "✅ Success! Response:"
  echo "$RESPONSE" | jq .
fi

