#!/bin/bash
# Usage: ./wait-token-deletion.sh <sa_id> <token_id> <timeout_seconds> <interval_seconds>

SA_ID="$1"
TOKEN_ID="$2"
TIMEOUT="$3"
INTERVAL="$4"

# Convert timeout and interval to seconds (remove 's' suffix if present)
TIMEOUT_SEC=$(echo "$TIMEOUT" | sed 's/s$//')
INTERVAL_SEC=$(echo "$INTERVAL" | sed 's/s$//')

end_time=$(($(date +%s) + TIMEOUT_SEC))

while [ $(date +%s) -lt $end_time ]; do
    # List all tokens for the service account and check if our token still exists
    response=$(kubectl exec -n "$NS" "$DEPLOYMENT" -- \
        curl -sSf -u "$USER:$PASS" "http://localhost:3000/api/serviceaccounts/$SA_ID/tokens" 2>/dev/null)

    if [ $? -eq 0 ]; then
        # Check if the token with TOKEN_ID exists in the response
        if ! echo "$response" | jq -e ".[] | select(.id == $TOKEN_ID)" > /dev/null 2>&1; then
            # Token not found in the list, which is what we want
            echo "Token $TOKEN_ID has been deleted from service account $SA_ID"
            exit 0
        fi
    else
        # Failed to get tokens list, but this might mean the SA itself was deleted
        echo "Warning: Failed to list tokens for SA $SA_ID" >&2
    fi

    sleep "$INTERVAL_SEC"
done

echo "Timeout waiting for token $TOKEN_ID to be deleted from SA $SA_ID" >&2
exit 1
