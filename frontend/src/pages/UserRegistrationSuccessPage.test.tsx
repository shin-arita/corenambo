import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect } from "vitest";
import UserRegistrationSuccessPage from "./UserRegistrationSuccessPage";

function renderPage() {
  return render(
    <MemoryRouter>
      <UserRegistrationSuccessPage />
    </MemoryRouter>
  );
}

describe("UserRegistrationSuccessPage", () => {
  it("ロゴが表示される", () => {
    renderPage();
    expect(screen.getByText("コレナンボ↓オークション")).toBeInTheDocument();
  });

  it("タイトルが表示される", () => {
    renderPage();
    expect(screen.getByRole("heading", { name: "本登録が完了しました" })).toBeInTheDocument();
  });

  it("お礼メッセージが表示される", () => {
    renderPage();
    expect(screen.getByText("ご登録いただきありがとうございます")).toBeInTheDocument();
  });

  it("ログインリンクが表示される", () => {
    renderPage();
    expect(screen.getByRole("link", { name: "ログイン" })).toBeInTheDocument();
  });
});
