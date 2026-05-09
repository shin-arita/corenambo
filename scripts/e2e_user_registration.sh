#!/usr/bin/env bash
set -euo pipefail

# ================================================================
# E2E: ユーザ仮登録 → 本登録
# ================================================================
# 使用方法: bash scripts/e2e_user_registration.sh
# 前提: docker compose up 済み (db / api / redis / worker / mail)
#
# トークン取得戦略:
#   worker が mail_outboxes を処理してメール送信後に payload を
#   '{}'に上書きするため、Mailpit API からメール本文を取得する
# ================================================================

API_URL="${API_URL:-http://localhost:8080}"
MAILPIT_URL="${MAILPIT_URL:-http://localhost:8025}"
DB_USER="${POSTGRES_USER:-app_user}"
DB_NAME="${POSTGRES_DB:-app_db}"

PASS_COUNT=0
FAIL_COUNT=0

# ----------------------------------------------------------------
# ユーティリティ
# ----------------------------------------------------------------
pass() { echo "  [PASS] $1"; PASS_COUNT=$((PASS_COUNT + 1)); }
fail() { echo "  [FAIL] $1"; FAIL_COUNT=$((FAIL_COUNT + 1)); exit 1; }

db_query() {
  docker compose exec -T db psql \
    -U "$DB_USER" -d "$DB_NAME" -tAc "$1"
}

# HTTP ステータスを先頭行に、レスポンスボディを残りに出力する (-f なし)
post_json() {
  local path="$1"
  local body="$2"
  local resp_file
  resp_file=$(mktemp)
  local status
  status=$(curl -s -o "$resp_file" -w "%{http_code}" \
    -X POST "${API_URL}${path}" \
    -H "Content-Type: application/json" \
    -H "Accept-Language: ja" \
    -d "$body")
  printf '%s\n%s' "$status" "$(cat "$resp_file")"
  rm -f "$resp_file"
}

# Mailpit からトークンを取得 (最大 10 秒リトライ)
get_token_from_mailpit() {
  local email="$1"
  # @ を %40 にエンコード
  local encoded_email
  encoded_email=$(echo "$email" | sed 's/@/%40/g')

  local attempt=0
  while [ "$attempt" -lt 10 ]; do
    local msg_id
    msg_id=$(curl -s "${MAILPIT_URL}/api/v1/messages?query=to%3A${encoded_email}" | \
      grep -o '"ID":"[^"]*"' | head -1 | cut -d'"' -f4)

    if [ -n "$msg_id" ]; then
      # Text フィールドから token= 以降を抽出
      local token
      token=$(curl -s "${MAILPIT_URL}/api/v1/message/${msg_id}" | \
        grep -o '"Text":"[^"]*"' | \
        grep -o 'token=[^\\]*' | \
        head -1 | \
        sed 's/token=//')
      if [ -n "$token" ]; then
        echo "$token"
        return 0
      fi
    fi

    attempt=$((attempt + 1))
    sleep 1
  done
  return 1
}

# テスト後のデータクリーンアップ
cleanup_email() {
  local email="$1"
  db_query "
    DELETE FROM user_credentials
      WHERE user_id IN (SELECT user_id FROM user_emails WHERE email = LOWER('${email}'));
    DELETE FROM user_emails WHERE email = LOWER('${email}');
    DELETE FROM users
      WHERE id NOT IN (SELECT user_id FROM user_emails);
    DELETE FROM user_registration_requests WHERE email = '${email}';
    DELETE FROM mail_outboxes WHERE to_email = '${email}';
  " > /dev/null
}

# ================================================================
# API 起動確認 (TCP 接続チェック)
# ================================================================
echo "== API 起動確認 =="
API_HOST=$(echo "$API_URL" | sed 's|http://||' | cut -d: -f1)
API_PORT=$(echo "$API_URL" | sed 's|.*:||')
attempt=0
until nc -z "$API_HOST" "$API_PORT" 2>/dev/null; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 30 ]; then
    echo "  APIが起動しませんでした (30秒タイムアウト)"
    exit 1
  fi
  sleep 1
done
echo "  API 起動確認 OK (${API_HOST}:${API_PORT})"

# ================================================================
# 事前クリーンアップ
# ================================================================
echo ""
echo "== 事前クリーンアップ =="

# Redis のレートリミットキーを削除
DELETED=$(docker compose exec -T redis redis-cli --no-auth-warning \
  EVAL "local keys = redis.call('keys', 'rate_limit:*'); if #keys > 0 then return redis.call('del', unpack(keys)) else return 0 end" 0 2>/dev/null || echo "0")
echo "  Redis rate_limit:* 削除キー数: ${DELETED}"

# Mailpit のメッセージを全削除
curl -s -X DELETE "${MAILPIT_URL}/api/v1/messages" > /dev/null
echo "  Mailpit メッセージを削除"

# ================================================================
# テスト 1: 正常系 (仮登録 → 本登録 → DB検証)
# ================================================================
echo ""
echo "=== テスト 1: 正常系 ==="

TS=$(date +%Y%m%d%H%M%S)
EMAIL="e2e_${TS}@example.com"
cleanup_email "$EMAIL" 2>/dev/null || true

# Step 1: 仮登録
echo "  -- Step 1: 仮登録 --"
RESULT=$(post_json "/api/v1/user-registration-requests" \
  "{\"email\":\"${EMAIL}\",\"email_confirmation\":\"${EMAIL}\"}")
STATUS=$(echo "$RESULT" | head -1)
BODY=$(echo "$RESULT" | tail -n +2)

if [ "$STATUS" != "201" ]; then
  fail "仮登録 HTTP 201 期待 → $STATUS (body: $BODY)"
fi
pass "仮登録 HTTP 201"

if ! echo "$BODY" | grep -q "USER_REGISTRATION_REQUEST_CREATED"; then
  fail "仮登録レスポンスコード: USER_REGISTRATION_REQUEST_CREATED 期待 → $BODY"
fi
pass "仮登録 code=USER_REGISTRATION_REQUEST_CREATED"

COUNT=$(db_query "SELECT COUNT(*) FROM user_registration_requests WHERE email = '${EMAIL}';")
if [ "$COUNT" != "1" ]; then
  fail "user_registration_requests: 1件期待 → ${COUNT}件"
fi
pass "user_registration_requests: 1件作成"

# Step 2: Mailpit からトークン取得
echo "  -- Step 2: トークン取得 (Mailpit API) --"
TOKEN=$(get_token_from_mailpit "$EMAIL") || fail "Mailpit からトークンを取得できませんでした"
pass "トークン取得成功 (${TOKEN:0:8}...)"

# Step 3: 本登録
echo "  -- Step 3: 本登録 --"
RESULT2=$(post_json "/api/v1/user-registrations/verify" \
  "{
    \"token\":\"${TOKEN}\",
    \"display_name\":\"E2Eテストユーザー\",
    \"password\":\"Password123!\",
    \"password_confirmation\":\"Password123!\",
    \"agreed_to_terms\":true
  }")
STATUS2=$(echo "$RESULT2" | head -1)
BODY2=$(echo "$RESULT2" | tail -n +2)

if [ "$STATUS2" != "201" ]; then
  fail "本登録 HTTP 201 期待 → $STATUS2 (body: $BODY2)"
fi
pass "本登録 HTTP 201"

if ! echo "$BODY2" | grep -q "USER_REGISTRATION_VERIFIED"; then
  fail "本登録レスポンスコード: USER_REGISTRATION_VERIFIED 期待 → $BODY2"
fi
pass "本登録 code=USER_REGISTRATION_VERIFIED"

# Step 4: DB検証
echo "  -- Step 4: DB検証 --"

USER_COUNT=$(db_query \
  "SELECT COUNT(*) FROM users u
     JOIN user_emails ue ON u.id = ue.user_id
   WHERE ue.email = LOWER('${EMAIL}');")
if [ "$USER_COUNT" != "1" ]; then
  fail "users: 1件期待 → ${USER_COUNT}件"
fi
pass "users: 1件作成"

IS_PRIMARY=$(db_query "SELECT is_primary FROM user_emails WHERE email = LOWER('${EMAIL}');")
if [ "$IS_PRIMARY" != "t" ]; then
  fail "user_emails.is_primary: true期待 → $IS_PRIMARY"
fi
pass "user_emails.is_primary=true"

HAS_VERIFIED_AT=$(db_query \
  "SELECT verified_at IS NOT NULL FROM user_emails WHERE email = LOWER('${EMAIL}');")
if [ "$HAS_VERIFIED_AT" != "t" ]; then
  fail "user_emails.verified_at: non-null期待"
fi
pass "user_emails.verified_at がセットされている"

CRED_COUNT=$(db_query \
  "SELECT COUNT(*) FROM user_credentials uc
     JOIN user_emails ue ON uc.user_id = ue.user_id
   WHERE ue.email = LOWER('${EMAIL}');")
if [ "$CRED_COUNT" != "1" ]; then
  fail "user_credentials: 1件期待 → ${CRED_COUNT}件"
fi
pass "user_credentials: 1件作成"

HAS_HASH=$(db_query \
  "SELECT password_hash <> '' FROM user_credentials uc
     JOIN user_emails ue ON uc.user_id = ue.user_id
   WHERE ue.email = LOWER('${EMAIL}');")
if [ "$HAS_HASH" != "t" ]; then
  fail "user_credentials.password_hash: 非空期待"
fi
pass "user_credentials.password_hash がセットされている"

REQ_VERIFIED=$(db_query \
  "SELECT verified_at IS NOT NULL FROM user_registration_requests WHERE email = '${EMAIL}';")
if [ "$REQ_VERIFIED" != "t" ]; then
  fail "user_registration_requests.verified_at: non-null期待"
fi
pass "user_registration_requests.verified_at がセットされている"

cleanup_email "$EMAIL"

# ================================================================
# テスト 2: 異常系 - トークン不正
# ================================================================
echo ""
echo "=== テスト 2: 異常系 - トークン不正 ==="

RESULT=$(post_json "/api/v1/user-registrations/verify" \
  "{
    \"token\":\"invalid-token-xyz-000\",
    \"display_name\":\"テストユーザー\",
    \"password\":\"Password123!\",
    \"password_confirmation\":\"Password123!\",
    \"agreed_to_terms\":true
  }")
STATUS=$(echo "$RESULT" | head -1)
BODY=$(echo "$RESULT" | tail -n +2)

if [ "$STATUS" != "400" ]; then
  fail "トークン不正 HTTP 400 期待 → $STATUS (body: $BODY)"
fi
pass "トークン不正 HTTP 400"

if ! echo "$BODY" | grep -q "INVALID_REGISTRATION_TOKEN"; then
  fail "トークン不正レスポンスコード: INVALID_REGISTRATION_TOKEN 期待 → $BODY"
fi
pass "トークン不正 code=INVALID_REGISTRATION_TOKEN"

# ================================================================
# テスト 3: 異常系 - トークン期限切れ
# ================================================================
echo ""
echo "=== テスト 3: 異常系 - トークン期限切れ ==="

# レートリミット解除
docker compose exec -T redis redis-cli --no-auth-warning \
  EVAL "local keys = redis.call('keys', 'rate_limit:*'); if #keys > 0 then return redis.call('del', unpack(keys)) else return 0 end" 0 > /dev/null 2>&1 || true

# Mailpit クリア (前テストのメッセージと混在させない)
curl -s -X DELETE "${MAILPIT_URL}/api/v1/messages" > /dev/null

TS=$(date +%Y%m%d%H%M%S)
EXP_EMAIL="e2e_exp_${TS}@example.com"
cleanup_email "$EXP_EMAIL" 2>/dev/null || true

RESULT=$(post_json "/api/v1/user-registration-requests" \
  "{\"email\":\"${EXP_EMAIL}\",\"email_confirmation\":\"${EXP_EMAIL}\"}")
STATUS=$(echo "$RESULT" | head -1)
if [ "$STATUS" != "201" ]; then
  fail "期限切れテスト 仮登録 HTTP 201 期待 → $STATUS"
fi

EXP_TOKEN=$(get_token_from_mailpit "$EXP_EMAIL") || fail "期限切れテスト: Mailpit からトークン取得失敗"

# DB で created_at / expires_at を過去に書き換える
# (expires_at > created_at 制約を満たすため両方更新)
db_query "UPDATE user_registration_requests
            SET created_at  = NOW() - INTERVAL '2 hours',
                last_sent_at = NOW() - INTERVAL '2 hours',
                expires_at  = NOW() - INTERVAL '1 hour'
          WHERE email = '${EXP_EMAIL}';" > /dev/null
pass "created_at/expires_at を過去に更新 (期限切れ状態)"

RESULT=$(post_json "/api/v1/user-registrations/verify" \
  "{
    \"token\":\"${EXP_TOKEN}\",
    \"display_name\":\"テストユーザー\",
    \"password\":\"Password123!\",
    \"password_confirmation\":\"Password123!\",
    \"agreed_to_terms\":true
  }")
STATUS=$(echo "$RESULT" | head -1)
BODY=$(echo "$RESULT" | tail -n +2)

if [ "$STATUS" != "400" ]; then
  fail "期限切れ HTTP 400 期待 → $STATUS (body: $BODY)"
fi
pass "期限切れ HTTP 400"

if ! echo "$BODY" | grep -q "EXPIRED_REGISTRATION_TOKEN"; then
  fail "期限切れレスポンスコード: EXPIRED_REGISTRATION_TOKEN 期待 → $BODY"
fi
pass "期限切れ code=EXPIRED_REGISTRATION_TOKEN"

cleanup_email "$EXP_EMAIL"

# ================================================================
# テスト 4: 異常系 - トークン使用済み
# ================================================================
echo ""
echo "=== テスト 4: 異常系 - トークン使用済み ==="

# レートリミット解除
docker compose exec -T redis redis-cli --no-auth-warning \
  EVAL "local keys = redis.call('keys', 'rate_limit:*'); if #keys > 0 then return redis.call('del', unpack(keys)) else return 0 end" 0 > /dev/null 2>&1 || true

# Mailpit クリア
curl -s -X DELETE "${MAILPIT_URL}/api/v1/messages" > /dev/null

TS=$(date +%Y%m%d%H%M%S)
USED_EMAIL="e2e_used_${TS}@example.com"
cleanup_email "$USED_EMAIL" 2>/dev/null || true

RESULT=$(post_json "/api/v1/user-registration-requests" \
  "{\"email\":\"${USED_EMAIL}\",\"email_confirmation\":\"${USED_EMAIL}\"}")
STATUS=$(echo "$RESULT" | head -1)
if [ "$STATUS" != "201" ]; then
  fail "使用済みテスト 仮登録 HTTP 201 期待 → $STATUS"
fi

USED_TOKEN=$(get_token_from_mailpit "$USED_EMAIL") || fail "使用済みテスト: Mailpit からトークン取得失敗"

VERIFY_BODY="{
  \"token\":\"${USED_TOKEN}\",
  \"display_name\":\"テストユーザー\",
  \"password\":\"Password123!\",
  \"password_confirmation\":\"Password123!\",
  \"agreed_to_terms\":true
}"

# 1回目の本登録 (成功)
RESULT=$(post_json "/api/v1/user-registrations/verify" "$VERIFY_BODY")
STATUS=$(echo "$RESULT" | head -1)
BODY=$(echo "$RESULT" | tail -n +2)
if [ "$STATUS" != "201" ]; then
  fail "使用済みテスト 1回目 HTTP 201 期待 → $STATUS (body: $BODY)"
fi
pass "使用済みテスト 1回目本登録 HTTP 201"

# 2回目の本登録 (使用済みエラー期待)
RESULT=$(post_json "/api/v1/user-registrations/verify" "$VERIFY_BODY")
STATUS=$(echo "$RESULT" | head -1)
BODY=$(echo "$RESULT" | tail -n +2)

if [ "$STATUS" != "409" ]; then
  fail "使用済み HTTP 409 期待 → $STATUS (body: $BODY)"
fi
pass "使用済み HTTP 409"

if ! echo "$BODY" | grep -q "USED_REGISTRATION_TOKEN"; then
  fail "使用済みレスポンスコード: USED_REGISTRATION_TOKEN 期待 → $BODY"
fi
pass "使用済み code=USED_REGISTRATION_TOKEN"

cleanup_email "$USED_EMAIL"

# ================================================================
# 結果サマリー
# ================================================================
echo ""
echo "================================================="
echo "  結果: PASS=${PASS_COUNT} FAIL=${FAIL_COUNT}"
echo "================================================="

if [ "$FAIL_COUNT" -ne 0 ]; then
  exit 1
fi

echo "  E2E 全テスト通過"
