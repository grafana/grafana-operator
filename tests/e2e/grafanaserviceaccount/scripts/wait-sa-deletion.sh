#!/bin/bash
# Usage: ./wait-sa-deletion.sh <sa_id> <timeout_seconds> <interval_seconds>

SA_ID="$1"
TIMEOUT="$2"
INTERVAL="$3"

# Convert timeout and interval to seconds (remove 's' suffix if present)
TIMEOUT_SEC=$(echo "$TIMEOUT" | sed 's/s$//')
INTERVAL_SEC=$(echo "$INTERVAL" | sed 's/s$//')

end_time=$(($(date +%s) + TIMEOUT_SEC))

while [ $(date +%s) -lt $end_time ]; do
    # Try to fetch the service account
    if ! kubectl exec -n "$NS" "$DEPLOYMENT" -- \
        curl -sSf -u "$USER:$PASS" "http://localhost:3000/api/serviceaccounts/$SA_ID" 2>/dev/null; then
        # Service account not found (404), which is what we want
        echo "Service account $SA_ID has been deleted from Grafana"
        exit 0
    fi
    sleep "$INTERVAL_SEC"
done

echo "Timeout waiting for SA $SA_ID to be deleted from Grafana" >&2
exit 1
