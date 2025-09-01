#!/bin/bash
# Usage: ./wait-sa.sh <sa_id> <timeout_seconds> <interval_seconds> <expected_json>

SA_ID="$1"
TIMEOUT="$2"
INTERVAL="$3"
EXPECTED="$4"

# Convert timeout and interval to seconds (remove 's' suffix if present)
TIMEOUT_SEC=$(echo "$TIMEOUT" | sed 's/s$//')
INTERVAL_SEC=$(echo "$INTERVAL" | sed 's/s$//')

end_time=$(($(date +%s) + TIMEOUT_SEC))
last_response=""

while [ $(date +%s) -lt $end_time ]; do
    if response=$(kubectl exec -n "$NS" "$DEPLOYMENT" -- \
        curl -sSf -u "$USER:$PASS" "http://localhost:3000/api/serviceaccounts/$SA_ID" 2>/dev/null); then

        last_response="$response"

        # Check if all expected fields match
        if echo "$response" | jq --argjson exp "$EXPECTED" '
            . as $curr | $exp | to_entries |
            all(.key as $k | .value == $curr[$k])' | grep -q true; then
            echo "$response"
            exit 0
        fi
    fi
    sleep "$INTERVAL_SEC"
done

echo "Timeout waiting for SA $SA_ID to match $EXPECTED" >&2
if [ -n "$last_response" ]; then
    echo "Last state:" >&2
    echo "$last_response" | jq . >&2
fi
exit 1
