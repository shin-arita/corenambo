import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";

export default function UserRegistrationCompletePage() {
  const { state } = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    if (!state) {
      navigate("/registration", { replace: true });
    }
  }, [state, navigate]);

  if (!state) return null;

  const { email, expiresMinutes } = state;

  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <p className="text-2xl font-bold tracking-tight text-slate-800">
            コレナンボ↓オークション
          </p>
        </div>

        <div className="bg-white rounded-2xl border border-slate-200 shadow-sm px-8 py-10">
          <h1 className="text-xl font-bold text-slate-900 mb-2">
            仮会員登録メールを送信しました
          </h1>
          <p className="text-sm text-slate-500 mb-6">
            <span className="font-medium text-slate-700">{email}</span>{" "}
            に本登録用のリンクを送信しました。
          </p>

          <div className="rounded-lg bg-slate-50 border border-slate-200 px-4 py-4 text-sm text-slate-600 space-y-2">
            <p>メール内のリンクをクリックして、会員登録を完了してください。</p>
            {expiresMinutes != null && (
              <p className="text-xs text-slate-400">
                このリンクの有効期限は{expiresMinutes}分です。
              </p>
            )}
            <p className="text-xs text-slate-400">
              ※メールが届かない場合は、迷惑メールフォルダをご確認ください。
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
