import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Routes, Route, useLocation } from "react-router-dom";
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";
import UserRegistrationPage from "./UserRegistrationPage";

function renderPage() {
  return render(
    <MemoryRouter>
      <UserRegistrationPage />
    </MemoryRouter>
  );
}

function CompleteStatePage() {
  const { state } = useLocation();
  return <div data-testid="complete-state">{JSON.stringify(state)}</div>;
}

function renderWithRoutes() {
  return render(
    <MemoryRouter initialEntries={["/registration"]}>
      <Routes>
        <Route path="/registration" element={<UserRegistrationPage />} />
        <Route path="/registration/complete" element={<CompleteStatePage />} />
      </Routes>
    </MemoryRouter>
  );
}

describe("UserRegistrationPage", () => {
  describe("初期表示", () => {
    it("タイトルと説明文が表示される", () => {
      renderPage();
      expect(screen.getByRole("heading", { name: "仮会員登録" })).toBeInTheDocument();
      expect(screen.getByText("メールアドレスを入力すると、本登録用のリンクをお送りします")).toBeInTheDocument();
    });

    it("メールアドレスと確認用のinputが表示される", () => {
      renderPage();
      const inputs = screen.getAllByPlaceholderText("example@mail.com");
      expect(inputs).toHaveLength(2);
    });

    it("送信ボタンが表示される", () => {
      renderPage();
      expect(screen.getByRole("button", { name: "登録メールを送信する" })).toBeInTheDocument();
    });

    it("ログインリンクが表示される", () => {
      renderPage();
      expect(screen.getByRole("link", { name: "ログイン" })).toBeInTheDocument();
    });
  });

  describe("クライアントサイドバリデーション", () => {
    it("空のまま送信するとメールアドレスのエラーが表示される", async () => {
      renderPage();
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));
      expect(await screen.findByText("メールアドレスを入力してください")).toBeInTheDocument();
    });

    it("無効なメールアドレスを入力するとエラーが表示される", async () => {
      renderPage();
      const inputs = screen.getAllByPlaceholderText("example@mail.com");
      await userEvent.type(inputs[0], "invalid-email");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));
      expect(await screen.findByText("正しいメールアドレス形式で入力してください")).toBeInTheDocument();
    });

    it("確認用メールアドレスが空のときエラーが表示される", async () => {
      renderPage();
      const inputs = screen.getAllByPlaceholderText("example@mail.com");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));
      expect(await screen.findByText("メールアドレス（確認）を入力してください")).toBeInTheDocument();
    });

    it("確認用メールアドレスが一致しない場合エラーが表示される", async () => {
      renderPage();
      const inputs = screen.getAllByPlaceholderText("example@mail.com");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.type(inputs[1], "other@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));
      expect(await screen.findByText("確認用メールアドレスが一致しません")).toBeInTheDocument();
    });
  });

  describe("API通信", () => {
    beforeEach(() => {
      vi.stubGlobal("fetch", vi.fn());
    });

    afterEach(() => {
      vi.unstubAllGlobals();
    });

    it("成功時に /registration/complete へ遷移する", async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ expires_minutes: 60 }),
      } as unknown as Response);

      const { container } = renderWithRoutes();
      const inputs = container.querySelectorAll("input[type='email']");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.type(inputs[1], "test@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));

      await waitFor(() => {
        expect(screen.getByTestId("complete-state")).toBeInTheDocument();
      });
    });

    it("成功時に state.email に入力メールアドレスが渡される", async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ expires_minutes: 60 }),
      } as unknown as Response);

      const { container } = renderWithRoutes();
      const inputs = container.querySelectorAll("input[type='email']");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.type(inputs[1], "test@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));

      await waitFor(() => screen.getByTestId("complete-state"));
      const state = JSON.parse(screen.getByTestId("complete-state").textContent ?? "{}");
      expect(state.email).toBe("test@example.com");
    });

    it("成功時に state.expiresMinutes に APIの expires_minutes が渡される", async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ expires_minutes: 60 }),
      } as unknown as Response);

      const { container } = renderWithRoutes();
      const inputs = container.querySelectorAll("input[type='email']");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.type(inputs[1], "test@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));

      await waitFor(() => screen.getByTestId("complete-state"));
      const state = JSON.parse(screen.getByTestId("complete-state").textContent ?? "{}");
      expect(state.expiresMinutes).toBe(60);
    });

    it("expires_minutes が 30 の場合 state.expiresMinutes が 30 になる", async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ expires_minutes: 30 }),
      } as unknown as Response);

      const { container } = renderWithRoutes();
      const inputs = container.querySelectorAll("input[type='email']");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.type(inputs[1], "test@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));

      await waitFor(() => screen.getByTestId("complete-state"));
      const state = JSON.parse(screen.getByTestId("complete-state").textContent ?? "{}");
      expect(state.expiresMinutes).toBe(30);
    });

    it("VALIDATION_ERRORのとき、フィールドエラーが表示される", async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          code: "VALIDATION_ERROR",
          errors: {
            email: [{ message: "このメールアドレスはすでに使用されています" }],
          },
        }),
      } as unknown as Response);

      renderPage();
      const inputs = screen.getAllByPlaceholderText("example@mail.com");
      await userEvent.type(inputs[0], "used@example.com");
      await userEvent.type(inputs[1], "used@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));

      expect(await screen.findByText("このメールアドレスはすでに使用されています")).toBeInTheDocument();
    });

    it("その他のAPIエラー時にフォームエラーが表示される", async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: false,
        json: async () => ({ message: "サーバーエラーが発生しました" }),
      } as unknown as Response);

      renderPage();
      const inputs = screen.getAllByPlaceholderText("example@mail.com");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.type(inputs[1], "test@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));

      expect(await screen.findByText("サーバーエラーが発生しました")).toBeInTheDocument();
    });

    it("VALIDATION_ERRORのとき、email_confirmationフィールドエラーが表示される", async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          code: "VALIDATION_ERROR",
          errors: {
            email_confirmation: [{ message: "確認用メールアドレスが一致しません" }],
          },
        }),
      } as unknown as Response);

      renderPage();
      const inputs = screen.getAllByPlaceholderText("example@mail.com");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.type(inputs[1], "test@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));

      expect(await screen.findByText("確認用メールアドレスが一致しません")).toBeInTheDocument();
    });

    it("VALIDATION_ERRORだが既知フィールドエラーがない場合はフォームエラーが表示される", async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: false,
        json: async () => ({
          code: "VALIDATION_ERROR",
          errors: {},
          message: "バリデーションエラーが発生しました",
        }),
      } as unknown as Response);

      renderPage();
      const inputs = screen.getAllByPlaceholderText("example@mail.com");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.type(inputs[1], "test@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));

      expect(await screen.findByText("バリデーションエラーが発生しました")).toBeInTheDocument();
    });

    it("APIエラーにmessageがない場合はデフォルトエラーメッセージが表示される", async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: false,
        json: async () => ({ code: "INTERNAL_ERROR" }),
      } as unknown as Response);

      renderPage();
      const inputs = screen.getAllByPlaceholderText("example@mail.com");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.type(inputs[1], "test@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));

      expect(await screen.findByText("エラーが発生しました。しばらく経ってから再度お試しください")).toBeInTheDocument();
    });

    it("通信エラー時に通信エラーメッセージが表示される", async () => {
      vi.mocked(fetch).mockRejectedValueOnce(new Error("network error"));

      renderPage();
      const inputs = screen.getAllByPlaceholderText("example@mail.com");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.type(inputs[1], "test@example.com");
      await userEvent.click(screen.getByRole("button", { name: "登録メールを送信する" }));

      expect(await screen.findByText("通信エラーが発生しました。しばらく経ってから再度お試しください")).toBeInTheDocument();
    });

    it("送信中はボタンが無効化される", async () => {
      vi.mocked(fetch).mockImplementationOnce(
        () => new Promise((resolve) => setTimeout(() => resolve({ ok: true, json: async () => ({ expires_minutes: 60 }) } as unknown as Response), 100))
      );

      renderPage();
      const inputs = screen.getAllByPlaceholderText("example@mail.com");
      await userEvent.type(inputs[0], "test@example.com");
      await userEvent.type(inputs[1], "test@example.com");

      const button = screen.getByRole("button", { name: "登録メールを送信する" });
      await userEvent.click(button);

      expect(await screen.findByRole("button", { name: "送信中..." })).toBeDisabled();
    });
  });
});
