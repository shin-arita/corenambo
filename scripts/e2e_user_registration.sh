#!/bin/sh
set -eu

API_URL="${API_URL:-http://localhost:8080}"
EMAIL="e2e_$(date +%Y%m%d%H%M%S)@example.com"

echo "== docker compose up =="
docker compose up -d db mail api

echo "== wait api =="
i=0
while [ "$i" -lt 30 ]; do
  if curl -s -o /dev/null "$API_URL"; then
    break
  fi
  i=$((i + 1))
  sleep 1
done

echo "== request =="
BODY_FILE="$(mktemp)"

STATUS=$(
  curl -s -o "$BODY_FILE" -w "%{http_code}" \
    -X POST "$API_URL/api/v1/user-registration-requests" \
    -H "Content-Type: application/json" \
    -H "Accept-Language: ja" \
    -d "{
      \"email\":\"$EMAIL\",
      \"email_confirmation\":\"$EMAIL\"
    }"
)

cat "$BODY_FILE"
echo

if [ "$STATUS" != "201" ]; then
  echo "NG: expected HTTP 201, got $STATUS"
  exit 1
fi

if ! grep -q "USER_REGISTRATION_REQUEST_CREATED" "$BODY_FILE"; then
  echo "NG: response code mismatch"
  exit 1
fi

echo "== db check =="
COUNT=$(
  docker compose exec -T db psql \
    -U "${POSTGRES_USER:-app_user}" \
    -d "${POSTGRES_DB:-app_db}" \
    -tAc "SELECT COUNT(*) FROM user_registration_requests WHERE email = '$EMAIL';"
)

if [ "$COUNT" != "1" ]; then
  echo "NG: DB row not found. count=$COUNT"
  exit 1
fi

echo "OK: E2E passed"
echo "email=$EMAIL"
