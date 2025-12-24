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

MODELS=$(curl -sSk "${HOST}/maas-api/v1/models" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" | jq -r .)

echo "$MODELS" | jq .

# Prompt user for index
TOTAL_MODELS=$(echo "$MODELS" | jq '.data | length')
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

curl -sSk -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"model\": \"${MODEL_NAME}\", \"prompt\": \"Hello\", \"max_tokens\": 50}" \
  "${MODEL_URL}/v1/completions"

