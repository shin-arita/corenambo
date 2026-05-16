import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";

/* c8 ignore next */
const API_BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

function validateEmail(value: string): string | null {
  if (!value.trim()) return "メールアドレスを入力してください";
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim())) return "正しいメールアドレス形式で入力してください";
  return null;
}

function validateEmailConfirmation(email: string, confirmation: string): string | null {
  if (!confirmation.trim()) return "メールアドレス（確認）を入力してください";
  if (email.trim() !== confirmation.trim()) return "確認用メールアドレスが一致しません";
  return null;
}

interface FieldErrors {
  email?: string;
  emailConfirmation?: string;
}

interface ApiValidationErrors {
  email?: Array<{ message: string }>;
  email_confirmation?: Array<{ message: string }>;
}

interface ApiResponse {
  code?: string;
  message?: string;
  expires_minutes?: number;
  errors?: ApiValidationErrors;
}

export default function UserRegistrationPage() {
  const navigate = useNavigate();

  const [email, setEmail] = useState("");
  const [emailConfirmation, setEmailConfirmation] = useState("");
  const [errors, setErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const newErrors: FieldErrors = {};
    const emailError = validateEmail(email);
    if (emailError) newErrors.email = emailError;

    const confirmError = validateEmailConfirmation(email, emailConfirmation);
    if (confirmError) newErrors.emailConfirmation = confirmError;

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    setErrors({});
    setFormError(null);
    setSubmitting(true);

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/user-registration-requests`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email: email.trim(),
          email_confirmation: emailConfirmation.trim(),
        }),
      });

      const data: ApiResponse = await response.json();

      if (response.ok) {
        navigate("/registration/complete", {
          state: { email: email.trim(), expiresMinutes: data.expires_minutes },
        });
        return;
      }

      if (data.code === "VALIDATION_ERROR" && data.errors) {
        const apiErrors: FieldErrors = {};
        if (data.errors.email?.[0]) apiErrors.email = data.errors.email[0].message;
        if (data.errors.email_confirmation?.[0]) apiErrors.emailConfirmation = data.errors.email_confirmation[0].message;
        if (Object.keys(apiErrors).length > 0) {
          setErrors(apiErrors);
          return;
        }
      }

      setFormError(data.message ?? "エラーが発生しました。しばらく経ってから再度お試しください");
    } catch {
      setFormError("通信エラーが発生しました。しばらく経ってから再度お試しください");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <p className="text-2xl font-bold tracking-tight text-slate-800">
            コレナンボ↓オークション
          </p>
        </div>

        <div className="bg-white rounded-2xl border border-slate-200 shadow-sm px-8 py-10">
          <h1 className="text-xl font-bold text-slate-900 mb-2">仮会員登録</h1>
          <p className="text-sm text-slate-500 mb-8">
            メールアドレスを入力すると、本登録用のリンクをお送りします
          </p>

          {formError && (
            <div className="mb-6 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-600">
              {formError}
            </div>
          )}

          <form onSubmit={handleSubmit} noValidate>
            <div className="mb-5">
              <label className="block text-sm font-medium text-slate-700 mb-1">
                メールアドレス
                <span className="ml-1.5 text-xs font-normal text-red-400">必須</span>
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="example@mail.com"
                className={`w-full rounded-lg border px-3 py-2.5 text-sm outline-none transition-colors ${
                  errors.email
                    ? "border-red-400 focus:border-red-400"
                    : "border-slate-300 focus:border-blue-500"
                }`}
              />
              {errors.email && (
                <p className="mt-1 text-xs text-red-500">{errors.email}</p>
              )}
            </div>

            <div className="mb-6">
              <label className="block text-sm font-medium text-slate-700 mb-1">
                メールアドレス（確認）
                <span className="ml-1.5 text-xs font-normal text-red-400">必須</span>
              </label>
              <input
                type="email"
                value={emailConfirmation}
                onChange={(e) => setEmailConfirmation(e.target.value)}
                placeholder="example@mail.com"
                className={`w-full rounded-lg border px-3 py-2.5 text-sm outline-none transition-colors ${
                  errors.emailConfirmation
                    ? "border-red-400 focus:border-red-400"
                    : "border-slate-300 focus:border-blue-500"
                }`}
              />
              {errors.emailConfirmation && (
                <p className="mt-1 text-xs text-red-500">{errors.emailConfirmation}</p>
              )}
            </div>

            <p className="mb-6 text-xs text-slate-400">
              ※メールが届かない場合は、迷惑メールフォルダをご確認ください
            </p>

            <button
              type="submit"
              disabled={submitting}
              className="w-full rounded-lg bg-blue-600 px-4 py-3 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {submitting ? "送信中..." : "登録メールを送信する"}
            </button>
          </form>

          <p className="mt-6 text-xs text-slate-400 text-center">
            登録することで、
            <Link to="/terms" className="text-blue-600 hover:underline">利用規約</Link>
            および
            <Link to="/privacy" className="text-blue-600 hover:underline">プライバシーポリシー</Link>
            に同意したものとみなします
          </p>

          <p className="mt-4 text-sm text-center text-slate-500">
            すでに会員の方は{" "}
            <Link to="/login" className="text-blue-600 hover:underline font-medium">
              ログイン
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
