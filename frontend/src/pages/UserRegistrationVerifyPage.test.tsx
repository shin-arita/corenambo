import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";
import UserRegistrationVerifyPage from "./UserRegistrationVerifyPage";

const VALID_GET_RESPONSE = {
  ok: true,
  json: async () => ({ code: "REGISTRATION_TOKEN_VALID", message: "本登録トークンは有効です" }),
} as unknown as Response;

function renderWithToken(token: string) {
  return render(
    <MemoryRouter initialEntries={[`/registration/verify?token=${token}`]}>
      <Routes>
        <Route path="/registration/verify" element={<UserRegistrationVerifyPage />} />
        <Route path="/registration/success" element={<div>成功ページ</div>} />
      </Routes>
    </MemoryRouter>
  );
}

function renderWithoutToken() {
  return render(
    <MemoryRouter initialEntries={["/registration/verify"]}>
      <Routes>
        <Route path="/registration/verify" element={<UserRegistrationVerifyPage />} />
      </Routes>
    </MemoryRouter>
  );
}

// -------- tokenなし --------

describe("UserRegistrationVerifyPage - tokenなし", () => {
  it("「本登録リンクが無効です」が表示される", () => {
    renderWithoutToken();
    expect(screen.getByText("本登録リンクが無効です")).toBeInTheDocument();
  });

  it("フォームが表示されない", () => {
    renderWithoutToken();
    expect(screen.queryByRole("button", { name: "本登録を完了する" })).not.toBeInTheDocument();
  });

  it("tokenがない場合はAPIを呼ばない", () => {
    vi.stubGlobal("fetch", vi.fn());
    renderWithoutToken();
    expect(fetch).not.toHaveBeenCalled();
    vi.unstubAllGlobals();
  });
});

// -------- 読み込み中 --------

describe("UserRegistrationVerifyPage - 読み込み中", () => {
  beforeEach(() => {
    vi.spyOn(window.history, "replaceState").mockImplementation(() => {});
    vi.stubGlobal("fetch", vi.fn());
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("初期表示中は「確認中...」が表示される", async () => {
    vi.mocked(fetch).mockImplementationOnce(() => new Promise(() => {}));
    renderWithToken("abc123");
    expect(await screen.findByText("確認中...")).toBeInTheDocument();
  });

  it("読み込み中はフォームが表示されない", async () => {
    vi.mocked(fetch).mockImplementationOnce(() => new Promise(() => {}));
    renderWithToken("abc123");
    await screen.findByText("確認中...");
    expect(screen.queryByRole("button", { name: "本登録を完了する" })).not.toBeInTheDocument();
  });
});

// -------- 初期表示（tokenあり） --------

describe("UserRegistrationVerifyPage - 初期表示（tokenあり）", () => {
  beforeEach(() => {
    vi.spyOn(window.history, "replaceState").mockImplementation(() => {});
    vi.stubGlobal("fetch", vi.fn());
    vi.mocked(fetch).mockResolvedValue(VALID_GET_RESPONSE);
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("タイトルが表示される", async () => {
    renderWithToken("abc123");
    expect(await screen.findByRole("heading", { name: "本会員登録" })).toBeInTheDocument();
  });

  it("本登録を完了するボタンが表示される", async () => {
    renderWithToken("abc123");
    expect(await screen.findByRole("button", { name: "本登録を完了する" })).toBeInTheDocument();
  });

  it("表示名の入力欄が表示される", async () => {
    renderWithToken("abc123");
    expect(await screen.findByPlaceholderText("例：タロウ")).toBeInTheDocument();
  });

  it("パスワードの入力欄が表示される", async () => {
    renderWithToken("abc123");
    expect(await screen.findByPlaceholderText("8文字以上、英字と数字を含む")).toBeInTheDocument();
  });

  it("パスワード確認の入力欄が表示される", async () => {
    renderWithToken("abc123");
    expect(await screen.findByPlaceholderText("パスワードを再入力")).toBeInTheDocument();
  });

  it("利用規約同意チェックボックスが表示される", async () => {
    renderWithToken("abc123");
    expect(await screen.findByRole("checkbox")).toBeInTheDocument();
  });

  it("マウント時にURLからtokenが除去される", () => {
    vi.mocked(fetch).mockImplementationOnce(() => new Promise(() => {}));
    renderWithToken("abc123");
    expect(window.history.replaceState).toHaveBeenCalledWith(null, "", expect.any(String));
  });

  it("tokenが画面上に表示されない", () => {
    vi.mocked(fetch).mockImplementationOnce(() => new Promise(() => {}));
    const { container } = renderWithToken("supersecrettoken");
    expect(container.textContent).not.toContain("supersecrettoken");
  });

  it("外部リンクにrel=\"noreferrer\"が付与されている", async () => {
    renderWithToken("abc123");
    await screen.findByRole("heading", { name: "本会員登録" });
    const links = screen.getAllByRole("link");
    links.forEach((link) => {
      if (link.getAttribute("target") === "_blank") {
        expect(link).toHaveAttribute("rel", "noreferrer");
      }
    });
  });

  it("console.logでtokenを出力しない", () => {
    vi.mocked(fetch).mockImplementationOnce(() => new Promise(() => {}));
    const spy = vi.spyOn(console, "log");
    renderWithToken("supersecrettoken");
    const output = spy.mock.calls.flat().join(" ");
    expect(output).not.toContain("supersecrettoken");
    spy.mockRestore();
  });

  it("GETにtokenがクエリパラメータとして送信される", async () => {
    renderWithToken("mytesttoken");
    await screen.findByRole("heading", { name: "本会員登録" });
    const getCall = (fetch as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(getCall[0]).toContain("token=mytesttoken");
    expect(getCall[1]).toBeUndefined();
  });
});

// -------- GETトークンチェックエラー --------

describe("UserRegistrationVerifyPage - GETトークンチェックエラー", () => {
  beforeEach(() => {
    vi.spyOn(window.history, "replaceState").mockImplementation(() => {});
    vi.stubGlobal("fetch", vi.fn());
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("INVALID_REGISTRATION_TOKENのとき、エラーメッセージが表示される", async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ code: "INVALID_REGISTRATION_TOKEN", message: "トークンが不正です" }),
    } as unknown as Response);

    renderWithToken("abc123");
    expect(await screen.findByText("トークンが不正です")).toBeInTheDocument();
  });

  it("EXPIRED_REGISTRATION_TOKENのとき、エラーメッセージが表示される", async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ code: "EXPIRED_REGISTRATION_TOKEN", message: "トークンの有効期限が切れています" }),
    } as unknown as Response);

    renderWithToken("abc123");
    expect(await screen.findByText("トークンの有効期限が切れています")).toBeInTheDocument();
  });

  it("USED_REGISTRATION_TOKENのとき、会員登録済み画面が表示される", async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ code: "USED_REGISTRATION_TOKEN", message: "既に登録が完了しています" }),
    } as unknown as Response);

    renderWithToken("abc123");
    expect(await screen.findByRole("heading", { name: "会員登録済み" })).toBeInTheDocument();
    expect(screen.getByText("このメールアドレスは既に登録されています")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ログインページへ" })).toBeInTheDocument();
  });

  it("USED_REGISTRATION_TOKEN後はフォームが表示されない", async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ code: "USED_REGISTRATION_TOKEN", message: "既に登録が完了しています" }),
    } as unknown as Response);

    renderWithToken("abc123");
    await screen.findByRole("heading", { name: "会員登録済み" });
    expect(screen.queryByRole("button", { name: "本登録を完了する" })).not.toBeInTheDocument();
  });

  it("USER_ALREADY_REGISTEREDのとき、会員登録済み画面が表示される", async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ code: "USER_ALREADY_REGISTERED", message: "入力されたメールアドレスは既に登録されています" }),
    } as unknown as Response);

    renderWithToken("abc123");
    expect(await screen.findByRole("heading", { name: "会員登録済み" })).toBeInTheDocument();
    expect(screen.getByText("このメールアドレスは既に登録されています")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ログインページへ" })).toBeInTheDocument();
  });

  it("USER_ALREADY_REGISTERED後はフォームが表示されない", async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ code: "USER_ALREADY_REGISTERED", message: "入力されたメールアドレスは既に登録されています" }),
    } as unknown as Response);

    renderWithToken("abc123");
    await screen.findByRole("heading", { name: "会員登録済み" });
    expect(screen.queryByRole("button", { name: "本登録を完了する" })).not.toBeInTheDocument();
  });

  it("エラーでmessageがない場合はデフォルトメッセージが表示される", async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ code: "INVALID_REGISTRATION_TOKEN" }),
    } as unknown as Response);

    renderWithToken("abc123");
    expect(await screen.findByText("本登録リンクが無効です")).toBeInTheDocument();
  });

  it("GETエラー後はフォームが表示されない", async () => {
    vi.mocked(fetch).mockResolvedValueOnce({
      ok: false,
      json: async () => ({ code: "EXPIRED_REGISTRATION_TOKEN", message: "トークンの有効期限が切れています" }),
    } as unknown as Response);

    renderWithToken("abc123");
    await screen.findByText("トークンの有効期限が切れています");
    expect(screen.queryByRole("button", { name: "本登録を完了する" })).not.toBeInTheDocument();
  });

  it("通信エラー時にtokenErrorが表示される", async () => {
    vi.mocked(fetch).mockRejectedValueOnce(new Error("network error"));

    renderWithToken("abc123");
    expect(await screen.findByText("通信エラーが発生しました。しばらく経ってから再度お試しください")).toBeInTheDocument();
  });

  it("通信エラー後はフォームが表示されない", async () => {
    vi.mocked(fetch).mockRejectedValueOnce(new Error("network error"));

    renderWithToken("abc123");
    await screen.findByText("通信エラーが発生しました。しばらく経ってから再度お試しください");
    expect(screen.queryByRole("button", { name: "本登録を完了する" })).not.toBeInTheDocument();
  });
});

// -------- 正常系 --------

describe("UserRegistrationVerifyPage - 正常系", () => {
  beforeEach(() => {
    vi.spyOn(window.history, "replaceState").mockImplementation(() => {});
    vi.stubGlobal("fetch", vi.fn());
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("成功時に /registration/success へ遷移する", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ code: "USER_REGISTRATION_VERIFIED", message: "本登録が完了しました" }),
      } as unknown as Response);

    renderWithToken("abc123");
    const submitBtn = await screen.findByRole("button", { name: "本登録を完了する" });
    await userEvent.type(screen.getByPlaceholderText("例：タロウ"), "テストユーザー");
    await userEvent.type(screen.getByPlaceholderText("8文字以上、英字と数字を含む"), "password123");
    await userEvent.type(screen.getByPlaceholderText("パスワードを再入力"), "password123");
    await userEvent.click(screen.getByRole("checkbox"));
    await userEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText("成功ページ")).toBeInTheDocument();
    });
  });

  it("POSTにtokenがクエリパラメータとして送信される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ code: "USER_REGISTRATION_VERIFIED", message: "本登録が完了しました" }),
      } as unknown as Response);

    renderWithToken("mytesttoken");
    const submitBtn = await screen.findByRole("button", { name: "本登録を完了する" });
    await userEvent.type(screen.getByPlaceholderText("例：タロウ"), "テストユーザー");
    await userEvent.type(screen.getByPlaceholderText("8文字以上、英字と数字を含む"), "password123");
    await userEvent.type(screen.getByPlaceholderText("パスワードを再入力"), "password123");
    await userEvent.click(screen.getByRole("checkbox"));
    await userEvent.click(submitBtn);

    await waitFor(() => {
      const postCall = (fetch as ReturnType<typeof vi.fn>).mock.calls[1];
      expect(postCall[0]).toContain("token=mytesttoken");
    });
  });

  it("bodyにdisplay_name/password/password_confirmation/agreed_to_termsが含まれる", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ code: "USER_REGISTRATION_VERIFIED", message: "本登録が完了しました" }),
      } as unknown as Response);

    renderWithToken("abc123");
    const submitBtn = await screen.findByRole("button", { name: "本登録を完了する" });
    await userEvent.type(screen.getByPlaceholderText("例：タロウ"), "テストユーザー");
    await userEvent.type(screen.getByPlaceholderText("8文字以上、英字と数字を含む"), "password123");
    await userEvent.type(screen.getByPlaceholderText("パスワードを再入力"), "password123");
    await userEvent.click(screen.getByRole("checkbox"));
    await userEvent.click(submitBtn);

    await waitFor(() => {
      const body = JSON.parse((fetch as ReturnType<typeof vi.fn>).mock.calls[1][1].body);
      expect(body.display_name).toBe("テストユーザー");
      expect(body.password).toBe("password123");
      expect(body.password_confirmation).toBe("password123");
      expect(body.agreed_to_terms).toBe(true);
    });
  });

  it("送信中はボタンが無効化される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockImplementationOnce(
        () => new Promise((resolve) =>
          setTimeout(() => resolve({ ok: true, json: async () => ({}) } as unknown as Response), 100)
        )
      );

    renderWithToken("abc123");
    const submitBtn = await screen.findByRole("button", { name: "本登録を完了する" });
    await userEvent.type(screen.getByPlaceholderText("例：タロウ"), "テストユーザー");
    await userEvent.type(screen.getByPlaceholderText("8文字以上、英字と数字を含む"), "password123");
    await userEvent.type(screen.getByPlaceholderText("パスワードを再入力"), "password123");
    await userEvent.click(screen.getByRole("checkbox"));
    await userEvent.click(submitBtn);

    expect(await screen.findByRole("button", { name: "登録中..." })).toBeDisabled();
  });
});

// -------- APIエラー分岐（POSTトークン致命的エラー） --------

describe("UserRegistrationVerifyPage - POSTトークン致命的エラー", () => {
  beforeEach(() => {
    vi.spyOn(window.history, "replaceState").mockImplementation(() => {});
    vi.stubGlobal("fetch", vi.fn());
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("INVALID_REGISTRATION_TOKENのとき、エラーメッセージが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({ code: "INVALID_REGISTRATION_TOKEN", message: "トークンが不正です" }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("トークンが不正です")).toBeInTheDocument();
  });

  it("EXPIRED_REGISTRATION_TOKENのとき、エラーメッセージが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({ code: "EXPIRED_REGISTRATION_TOKEN", message: "トークンの有効期限が切れています" }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("トークンの有効期限が切れています")).toBeInTheDocument();
  });

  it("USED_REGISTRATION_TOKENのとき、会員登録済み画面が表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({ code: "USED_REGISTRATION_TOKEN", message: "既に登録が完了しています" }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByRole("heading", { name: "会員登録済み" })).toBeInTheDocument();
    expect(screen.getByText("このメールアドレスは既に登録されています")).toBeInTheDocument();
  });

  it("USER_ALREADY_REGISTEREDのとき、会員登録済み画面が表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({ code: "USER_ALREADY_REGISTERED", message: "入力されたメールアドレスは既に登録されています" }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByRole("heading", { name: "会員登録済み" })).toBeInTheDocument();
    expect(screen.getByText("このメールアドレスは既に登録されています")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ログインページへ" })).toBeInTheDocument();
  });

  it("致命的エラーでmessageがない場合はデフォルトメッセージが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({ code: "INVALID_REGISTRATION_TOKEN" }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("本登録リンクが無効です")).toBeInTheDocument();
  });

  it("致命的エラー後はフォームが非表示になる", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({ code: "USED_REGISTRATION_TOKEN", message: "既に登録が完了しています" }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    await waitFor(() => {
      expect(screen.queryByRole("button", { name: "本登録を完了する" })).not.toBeInTheDocument();
    });
  });
});

// -------- VALIDATION_ERROR（フィールドエラー） --------

describe("UserRegistrationVerifyPage - VALIDATION_ERROR", () => {
  beforeEach(() => {
    vi.spyOn(window.history, "replaceState").mockImplementation(() => {});
    vi.stubGlobal("fetch", vi.fn());
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("display_nameフィールドエラーが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          code: "VALIDATION_ERROR",
          errors: { display_name: [{ code: "DISPLAY_NAME_REQUIRED", message: "ユーザ名を入力してください" }] },
        }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("ユーザ名を入力してください")).toBeInTheDocument();
  });

  it("passwordフィールドエラーが表示される（未入力）", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          code: "VALIDATION_ERROR",
          errors: { password: [{ code: "PASSWORD_REQUIRED", message: "パスワードを入力してください" }] },
        }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("パスワードを入力してください")).toBeInTheDocument();
  });

  it("PASSWORD_TOO_WEAKが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          code: "VALIDATION_ERROR",
          errors: { password: [{ code: "PASSWORD_TOO_WEAK", message: "パスワードは8文字以上で、英字と数字をそれぞれ1文字以上含めてください" }] },
        }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("パスワードは8文字以上で、英字と数字をそれぞれ1文字以上含めてください")).toBeInTheDocument();
  });

  it("password_confirmationフィールドエラーが表示される（不一致）", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          code: "VALIDATION_ERROR",
          errors: { password_confirmation: [{ code: "PASSWORD_CONFIRMATION_NOT_MATCH", message: "パスワードが一致しません" }] },
        }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("パスワードが一致しません")).toBeInTheDocument();
  });

  it("password_confirmationフィールドエラーが表示される（未入力）", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          code: "VALIDATION_ERROR",
          errors: { password_confirmation: [{ code: "PASSWORD_CONFIRMATION_REQUIRED", message: "確認用パスワードを入力してください" }] },
        }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("確認用パスワードを入力してください")).toBeInTheDocument();
  });

  it("agreed_to_termsフィールドエラーが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          code: "VALIDATION_ERROR",
          errors: { agreed_to_terms: [{ code: "AGREED_TO_TERMS_REQUIRED", message: "利用規約への同意が必要です" }] },
        }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("利用規約への同意が必要です")).toBeInTheDocument();
  });

  it("VALIDATION_ERRORだが既知フィールドエラーがない場合はフォームエラーが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          code: "VALIDATION_ERROR",
          errors: {},
          message: "バリデーションエラーが発生しました",
        }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("バリデーションエラーが発生しました")).toBeInTheDocument();
  });

  it("VALIDATION_ERRORかつerrorsがnullの場合はフォームエラーが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          code: "VALIDATION_ERROR",
          errors: null,
          message: "バリデーションエラーが発生しました",
        }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("バリデーションエラーが発生しました")).toBeInTheDocument();
  });
});

// -------- その他APIエラー --------

describe("UserRegistrationVerifyPage - その他APIエラー", () => {
  beforeEach(() => {
    vi.spyOn(window.history, "replaceState").mockImplementation(() => {});
    vi.stubGlobal("fetch", vi.fn());
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("TOO_MANY_REQUESTSのとき、フォームエラーが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({ code: "TOO_MANY_REQUESTS", message: "リクエストが多すぎます。しばらく待ってから再試行してください" }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("リクエストが多すぎます。しばらく待ってから再試行してください")).toBeInTheDocument();
  });

  it("INTERNAL_SERVER_ERRORのとき、フォームエラーが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({ code: "INTERNAL_SERVER_ERROR", message: "システムエラーが発生しました" }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("システムエラーが発生しました")).toBeInTheDocument();
  });

  it("APIエラーにmessageがない場合はデフォルトエラーメッセージが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockResolvedValueOnce({
        ok: false,
        json: async () => ({ code: "UNKNOWN_ERROR" }),
      } as unknown as Response);

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("エラーが発生しました。しばらく経ってから再度お試しください")).toBeInTheDocument();
  });

  it("通信エラー時にフォームエラーメッセージが表示される", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(VALID_GET_RESPONSE)
      .mockRejectedValueOnce(new Error("network error"));

    renderWithToken("abc123");
    await userEvent.click(await screen.findByRole("button", { name: "本登録を完了する" }));

    expect(await screen.findByText("通信エラーが発生しました。しばらく経ってから再度お試しください")).toBeInTheDocument();
  });
});
