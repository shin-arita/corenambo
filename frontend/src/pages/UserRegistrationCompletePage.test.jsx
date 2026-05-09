import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, it, expect } from "vitest";
import UserRegistrationCompletePage from "./UserRegistrationCompletePage";

function renderWithState(state) {
  return render(
    <MemoryRouter initialEntries={[{ pathname: "/registration/complete", state }]}>
      <Routes>
        <Route path="/registration/complete" element={<UserRegistrationCompletePage />} />
        <Route path="/registration" element={<div>仮会員登録ページ</div>} />
      </Routes>
    </MemoryRouter>
  );
}

describe("UserRegistrationCompletePage", () => {
  describe("初期表示", () => {
    it("タイトルが表示される", () => {
      renderWithState({ email: "test@example.com", expiresMinutes: 60 });
      expect(screen.getByRole("heading", { name: "仮会員登録メールを送信しました" })).toBeInTheDocument();
    });

    it("ロゴが表示される", () => {
      renderWithState({ email: "test@example.com", expiresMinutes: 60 });
      expect(screen.getByText("コレナンボ↓オークション")).toBeInTheDocument();
    });

    it("ログインボタンは表示されない", () => {
      renderWithState({ email: "test@example.com", expiresMinutes: 60 });
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
    });
  });

  describe("メールアドレスの表示", () => {
    it("メールアドレスが表示される", () => {
      renderWithState({ email: "test@example.com", expiresMinutes: 60 });
      expect(screen.getByText("test@example.com")).toBeInTheDocument();
    });

    it("本登録リンク送信メッセージが表示される", () => {
      renderWithState({ email: "test@example.com", expiresMinutes: 60 });
      expect(screen.getByText("に本登録用のリンクを送信しました。", { exact: false })).toBeInTheDocument();
    });
  });

  describe("有効期限の表示", () => {
    it("expiresMinutesがある場合、有効期限が表示される", () => {
      renderWithState({ email: "test@example.com", expiresMinutes: 60 });
      expect(screen.getByText("このリンクの有効期限は60分です。")).toBeInTheDocument();
    });

    it("expiresMinutesがnullの場合、有効期限は表示されない", () => {
      renderWithState({ email: "test@example.com", expiresMinutes: null });
      expect(screen.queryByText(/有効期限/)).not.toBeInTheDocument();
    });

    it("expiresMinutes が 30 の場合、有効期限文言に「30分」が含まれる", () => {
      renderWithState({ email: "test@example.com", expiresMinutes: 30 });
      expect(screen.getByText("このリンクの有効期限は30分です。")).toBeInTheDocument();
    });
  });

  describe("注意事項", () => {
    it("迷惑メールフォルダの注意書きが表示される", () => {
      renderWithState({ email: "test@example.com", expiresMinutes: 60 });
      expect(screen.getByText("※メールが届かない場合は、迷惑メールフォルダをご確認ください。")).toBeInTheDocument();
    });

    it("メイン指示文が表示される", () => {
      renderWithState({ email: "test@example.com", expiresMinutes: 60 });
      expect(screen.getByText("メール内のリンクをクリックして、会員登録を完了してください。")).toBeInTheDocument();
    });
  });

  describe("直接アクセス防止", () => {
    it("stateがない場合、/registration にリダイレクトされる", () => {
      renderWithState(null);
      expect(screen.getByText("仮会員登録ページ")).toBeInTheDocument();
    });

    it("stateがない場合、完了画面は表示されない", () => {
      renderWithState(null);
      expect(screen.queryByRole("heading", { name: "仮会員登録メールを送信しました" })).not.toBeInTheDocument();
    });
  });
});
