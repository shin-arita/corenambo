import { useEffect, useRef, useState } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";

/* c8 ignore next */
const API_BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

const TOKEN_FATAL_CODES = new Set([
  "INVALID_REGISTRATION_TOKEN",
  "EXPIRED_REGISTRATION_TOKEN",
  "USED_REGISTRATION_TOKEN",
  "USER_ALREADY_REGISTERED",
]);

interface FieldErrors {
  displayName?: string;
  password?: string;
  passwordConfirmation?: string;
  agreedToTerms?: string;
}

interface ApiValidationErrors {
  display_name?: Array<{ message: string }>;
  password?: Array<{ message: string }>;
  password_confirmation?: Array<{ message: string }>;
  agreed_to_terms?: Array<{ message: string }>;
}

interface ApiResponse {
  code?: string;
  message?: string;
  errors?: ApiValidationErrors;
}

export default function UserRegistrationVerifyPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const [token] = useState(() => searchParams.get("token") ?? "");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const [passwordConfirmation, setPasswordConfirmation] = useState("");
  const [agreedToTerms, setAgreedToTerms] = useState(false);
  const [errors, setErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [tokenError, setTokenError] = useState<string | null>(null);
  const [alreadyRegistered, setAlreadyRegistered] = useState(false);
  const [tokenCheckLoading, setTokenCheckLoading] = useState(token !== "");
  const [submitting, setSubmitting] = useState(false);
  const fetchedRef = useRef(false);

  useEffect(() => {
    if (token) {
      window.history.replaceState(null, "", window.location.pathname);
    }

    if (!token) return;
    if (fetchedRef.current) return;
    fetchedRef.current = true;

    fetch(
      `${API_BASE_URL}/api/v1/user-registrations/verify?token=${encodeURIComponent(token)}`
    )
      .then(async (res) => {
        if (res.ok) {
          setTokenCheckLoading(false);
          return;
        }
        const data: ApiResponse = await res.json();
        if (data.code === "USER_ALREADY_REGISTERED" || data.code === "USED_REGISTRATION_TOKEN") {
          setAlreadyRegistered(true);
        } else {
          setTokenError(data.message ?? "本登録リンクが無効です");
        }
        setTokenCheckLoading(false);
      })
      .catch(() => {
        setTokenError("通信エラーが発生しました。しばらく経ってから再度お試しください");
        setTokenCheckLoading(false);
      });
  }, [token]);

  if (!token) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          <div className="text-center mb-8">
            <p className="text-2xl font-bold tracking-tight text-slate-800">
              コレナンボ↓オークション
            </p>
          </div>
          <div className="bg-white rounded-2xl border border-slate-200 shadow-sm px-8 py-10 text-center">
            <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-600 text-center">
              本登録リンクが無効です
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (tokenCheckLoading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          <div className="text-center mb-8">
            <p className="text-2xl font-bold tracking-tight text-slate-800">
              コレナンボ↓オークション
            </p>
          </div>
          <div className="bg-white rounded-2xl border border-slate-200 shadow-sm px-8 py-10 text-center">
            <p className="text-sm text-slate-500 text-center">確認中...</p>
          </div>
        </div>
      </div>
    );
  }

  if (alreadyRegistered) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          <div className="text-center mb-8">
            <p className="text-2xl font-bold tracking-tight text-slate-800">
              コレナンボ↓オークション
            </p>
          </div>
          <div className="bg-white rounded-2xl border border-slate-200 shadow-sm px-8 py-10 text-center">
            <h1 className="text-xl font-bold text-slate-900 mb-2">会員登録済み</h1>
            <p className="text-sm text-slate-500 mb-8">
              このメールアドレスは既に登録されています
            </p>
            <a
              href="/login"
              className="block w-full rounded-lg bg-blue-600 px-4 py-3 text-center text-sm font-semibold text-white hover:bg-blue-700 transition-colors"
            >
              ログインページへ
            </a>
          </div>
        </div>
      </div>
    );
  }

  if (tokenError) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          <div className="text-center mb-8">
            <p className="text-2xl font-bold tracking-tight text-slate-800">
              コレナンボ↓オークション
            </p>
          </div>
          <div className="bg-white rounded-2xl border border-slate-200 shadow-sm px-8 py-10 text-center">
            <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-600 text-center">
              {tokenError}
            </div>
          </div>
        </div>
      </div>
    );
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrors({});
    setFormError(null);
    setSubmitting(true);

    try {
      const response = await fetch(
        `${API_BASE_URL}/api/v1/user-registrations/verify?token=${encodeURIComponent(token)}`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            display_name: displayName,
            password,
            password_confirmation: passwordConfirmation,
            agreed_to_terms: agreedToTerms,
          }),
        }
      );

      const data: ApiResponse = await response.json();

      if (response.ok) {
        navigate("/registration/success");
        return;
      }

      const code = data.code;

      if (code === "USER_ALREADY_REGISTERED" || code === "USED_REGISTRATION_TOKEN") {
        setAlreadyRegistered(true);
        return;
      }

      if (code !== undefined && TOKEN_FATAL_CODES.has(code)) {
        setTokenError(data.message ?? "本登録リンクが無効です");
        return;
      }

      if (code === "VALIDATION_ERROR" && data.errors) {
        const apiErrors: FieldErrors = {};
        if (data.errors.display_name?.[0]) apiErrors.displayName = data.errors.display_name[0].message;
        if (data.errors.password?.[0]) apiErrors.password = data.errors.password[0].message;
        if (data.errors.password_confirmation?.[0]) apiErrors.passwordConfirmation = data.errors.password_confirmation[0].message;
        if (data.errors.agreed_to_terms?.[0]) apiErrors.agreedToTerms = data.errors.agreed_to_terms[0].message;
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
          <h1 className="text-xl font-bold text-slate-900 mb-2 text-center">本会員登録</h1>
          <p className="text-sm text-slate-500 mb-8 text-center">
            以下の情報を入力して、本登録を完了してください
          </p>

          {formError && (
            <div className="mb-6 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-600 text-center">
              {formError}
            </div>
          )}

          <form onSubmit={handleSubmit} noValidate>
            <div className="mb-5">
              <label className="block text-sm font-medium text-slate-700 mb-1">
                表示名
                <span className="ml-1.5 text-xs font-normal text-red-400">必須</span>
              </label>
              <input
                type="text"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="例：タロウ"
                className={`w-full rounded-lg border px-3 py-2.5 text-sm outline-none transition-colors ${
                  errors.displayName
                    ? "border-red-400 focus:border-red-400"
                    : "border-slate-300 focus:border-blue-500"
                }`}
              />
              {errors.displayName && (
                <p className="mt-1 text-xs text-red-500">{errors.displayName}</p>
              )}
            </div>

            <div className="mb-5">
              <label className="block text-sm font-medium text-slate-700 mb-1">
                パスワード
                <span className="ml-1.5 text-xs font-normal text-red-400">必須</span>
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="8文字以上、英字と数字を含む"
                className={`w-full rounded-lg border px-3 py-2.5 text-sm outline-none transition-colors ${
                  errors.password
                    ? "border-red-400 focus:border-red-400"
                    : "border-slate-300 focus:border-blue-500"
                }`}
              />
              {errors.password && (
                <p className="mt-1 text-xs text-red-500">{errors.password}</p>
              )}
            </div>

            <div className="mb-5">
              <label className="block text-sm font-medium text-slate-700 mb-1">
                パスワード（確認）
                <span className="ml-1.5 text-xs font-normal text-red-400">必須</span>
              </label>
              <input
                type="password"
                value={passwordConfirmation}
                onChange={(e) => setPasswordConfirmation(e.target.value)}
                placeholder="パスワードを再入力"
                className={`w-full rounded-lg border px-3 py-2.5 text-sm outline-none transition-colors ${
                  errors.passwordConfirmation
                    ? "border-red-400 focus:border-red-400"
                    : "border-slate-300 focus:border-blue-500"
                }`}
              />
              {errors.passwordConfirmation && (
                <p className="mt-1 text-xs text-red-500">{errors.passwordConfirmation}</p>
              )}
            </div>

            <div className="mb-6">
              <label className="flex items-start gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={agreedToTerms}
                  onChange={(e) => setAgreedToTerms(e.target.checked)}
                  className="mt-0.5"
                />
                <span className="text-sm text-slate-600">
                  <a href="/terms" target="_blank" rel="noreferrer" className="text-blue-600 hover:underline">
                    利用規約
                  </a>
                  および
                  <a href="/privacy" target="_blank" rel="noreferrer" className="text-blue-600 hover:underline">
                    プライバシーポリシー
                  </a>
                  に同意する
                  <span className="ml-1.5 text-xs font-normal text-red-400">必須</span>
                </span>
              </label>
              {errors.agreedToTerms && (
                <p className="mt-1 text-xs text-red-500">{errors.agreedToTerms}</p>
              )}
            </div>

            <button
              type="submit"
              disabled={submitting}
              className="w-full rounded-lg bg-blue-600 px-4 py-3 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {submitting ? "登録中..." : "本登録を完了する"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
