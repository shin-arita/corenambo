import { Link } from "react-router-dom";

export default function UserRegistrationSuccessPage() {
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
            本登録が完了しました
          </h1>
          <p className="text-sm text-slate-500 mb-6">
            ご登録いただきありがとうございます
          </p>
          <p className="text-sm text-center text-slate-500 mt-4">
            <Link to="/login" className="text-blue-600 hover:underline font-medium">
              ログイン
            </Link>
            してご利用ください
          </p>
        </div>
      </div>
    </div>
  );
}
