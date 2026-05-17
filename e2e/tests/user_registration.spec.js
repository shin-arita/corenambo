// @ts-check
const { test, expect } = require('@playwright/test');
const { randomUUID } = require('crypto');

const MAILPIT_API = 'http://mail:8025/api/v1';
const POLL_INTERVAL_MS = 1000;
const POLL_TIMEOUT_MS = 30000;

/**
 * @param {import('@playwright/test').APIRequestContext} apiContext
 * @param {string} toEmail
 * @param {number} since - Unix ms（Created フィールドがこの値以上のメールのみ対象）
 * @returns {Promise<{ID: string, Date: string} | null>}
 */
async function pollForEmail(apiContext, toEmail, since) {
  const deadline = Date.now() + POLL_TIMEOUT_MS;
  while (Date.now() < deadline) {
    const res = await apiContext.get(
      `${MAILPIT_API}/search?query=${encodeURIComponent('to:' + toEmail)}&limit=10`
    );
    if (res.ok()) {
      const body = await res.json();
      if (body.messages && body.messages.length > 0) {
        const msg = body.messages.find(
          (m) => new Date(m.Created).getTime() >= since
        );
        if (msg) return msg;
      }
    }
    await new Promise((r) => setTimeout(r, POLL_INTERVAL_MS));
  }
  return null;
}

test('仮登録正常系：フォーム送信 → 完了画面遷移 → メール到着 → token確認', async ({ page, request }) => {
  const testEmail = `e2e-${Date.now()}-${randomUUID()}@example.com`;
  // 5秒バッファを持たせ、テスト開始前の残留メールを誤検出しない
  const testStartedAt = Date.now() - 5000;

  await page.goto('/registration');

  await page.locator('#email').fill(testEmail);
  await page.locator('#emailConfirmation').fill(testEmail);
  await page.getByRole('button', { name: '登録メールを送信する' }).click();

  await expect(
    page.getByRole('heading', { name: '仮会員登録メールを送信しました' })
  ).toBeVisible();

  const message = await pollForEmail(request, testEmail, testStartedAt);
  expect(message, 'Mailpit にメールが届いていること').not.toBeNull();

  const msgRes = await request.get(`${MAILPIT_API}/message/${message.ID}`);
  expect(msgRes.ok(), 'Mailpit メッセージ詳細取得が成功すること').toBeTruthy();

  const msgData = await msgRes.json();
  const bodyText = msgData.Text ?? '';

  expect(bodyText, 'メール本文に verify URL が含まれること').toContain('/registration/verify');
  expect(bodyText, 'メール本文に token が含まれること').toContain('token=');
});
