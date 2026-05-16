import { render, screen } from "@testing-library/react";
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";
import App from "./App";

const VALID_GET_RESPONSE = {
  ok: true,
  json: async () => ({ code: "REGISTRATION_TOKEN_VALID", message: "本登録トークンは有効です" }),
} as unknown as Response;

// BrowserRouter はマウント時に window.location を読む。
// pushState で URL をセットしてから render することでルーティングをテストする。

describe("App ルーティング", () => {
  beforeEach(() => {
    // VerifyPage が mount 時に replaceState を呼ぶため、テストに影響しないようモックする
    vi.spyOn(window.history, "replaceState").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
    window.history.pushState({}, "", "/");
  });

  it("/registration/verify?token=xxx でverify画面が表示される", async () => {
    vi.stubGlobal("fetch", vi.fn());
    vi.mocked(fetch).mockResolvedValueOnce(VALID_GET_RESPONSE);
    window.history.pushState({}, "", "/registration/verify?token=testtoken");
    render(<App />);
    expect(await screen.findByRole("heading", { name: "本会員登録" })).toBeInTheDocument();
    vi.unstubAllGlobals();
  });

  it("/registration/verify でtokenなしの場合「本登録リンクが無効です」が表示される", () => {
    window.history.pushState({}, "", "/registration/verify");
    render(<App />);
    expect(screen.getByText("本登録リンクが無効です")).toBeInTheDocument();
  });

  it("/registration で仮会員登録ページが表示される（既存ルートの回帰確認）", () => {
    window.history.pushState({}, "", "/registration");
    render(<App />);
    expect(screen.getByRole("heading", { name: "仮会員登録" })).toBeInTheDocument();
  });

  it("/registration/success で成功ページが表示される", () => {
    window.history.pushState({}, "", "/registration/success");
    render(<App />);
    expect(screen.getByRole("heading", { name: "本登録が完了しました" })).toBeInTheDocument();
  });
});
