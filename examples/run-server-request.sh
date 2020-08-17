#!/bin/bash
response=$(curl -s -XPOST \
	-H 'Content-Type: application/json' \
	-d "@admission-review.example.json" \
	'http://localhost:8443/mutate' | jq . )

echo "Response: "
echo "$response"
echo ""
jsonpatch=$(echo "$response" | jq -r .response.patch | base64 -d | jq .)
echo "Decoded jsonPatch Response: "
echo "$jsonpatch"

