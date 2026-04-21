import { useEffect, useMemo, useState } from "react";

const API_BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export default function App() {
  const [email, setEmail] = useState("test@example.com");
  const [password, setPassword] = useState("password123");
  const [accessToken, setAccessToken] = useState("");
  const [me, setMe] = useState(null);
  const [searchWord, setSearchWord] = useState("防水");
  const [products, setProducts] = useState([]);
  const [message, setMessage] = useState("");

  const loggedIn = useMemo(() => accessToken !== "", [accessToken]);

  useEffect(() => {
    void refreshTokenOnLoad();
  }, []);

  useEffect(() => {
    if (!accessToken) {
      setMe(null);
      return;
    }
    void fetchMe(accessToken);
  }, [accessToken]);

  async function refreshTokenOnLoad() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
        method: "POST",
        credentials: "include"
      });

      if (!response.ok) {
        return;
      }

      const json = await response.json();
      setAccessToken(json.data.accessToken);
      setMessage("リフレッシュトークンから再ログインしました。\n");
    } catch {
      setMessage("");
    }
  }

  async function login(event) {
    event.preventDefault();
    setMessage("");

    const response = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ email, password })
    });

    const json = await response.json();

    if (!response.ok) {
      setMessage(json.error ?? "ログインに失敗しました。\n");
      return;
    }

    setAccessToken(json.data.accessToken);
    setMessage("ログインしました。\n");
  }

  async function fetchMe(token) {
    const response = await fetch(`${API_BASE_URL}/api/v1/me`, {
      method: "GET",
      credentials: "include",
      headers: {
        Authorization: `Bearer ${token}`
      }
    });

    const json = await response.json();

    if (!response.ok) {
      setMessage(json.error ?? "ユーザー情報の取得に失敗しました。\n");
      return;
    }

    setMe(json.data);
  }

  async function searchProducts(event) {
    event.preventDefault();
    setMessage("");

    const response = await fetch(
      `${API_BASE_URL}/api/v1/products/search?q=${encodeURIComponent(searchWord)}`,
      {
        method: "GET",
        credentials: "include"
      }
    );

    const json = await response.json();

    if (!response.ok) {
      setMessage(json.error ?? "商品検索に失敗しました。\n");
      return;
    }

    setProducts(json.data);
    setMessage(`${json.data.length}件取得しました。\n`);
  }

  async function logout() {
    await fetch(`${API_BASE_URL}/api/v1/auth/logout`, {
      method: "POST",
      credentials: "include"
    });

    setAccessToken("");
    setMe(null);
    setProducts([]);
    setMessage("ログアウトしました。\n");
  }

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <div className="mx-auto max-w-5xl px-4 py-10">
        <header className="mb-8 rounded-3xl bg-white p-6 shadow-sm ring-1 ring-slate-200">
          <h1 className="text-3xl font-bold">corenambo</h1>
          <p className="mt-2 text-sm text-slate-600">
            React + Tailwind + Go + PostgreSQL(PGroonga) の最小完成版です。
          </p>
        </header>

        <div className="grid gap-6 lg:grid-cols-2">
          <section className="rounded-3xl bg-white p-6 shadow-sm ring-1 ring-slate-200">
            <h2 className="mb-4 text-xl font-semibold">ログイン</h2>

            <form className="space-y-4" onSubmit={login}>
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-700">
                  メールアドレス
                </label>
                <input
                  className="w-full rounded-xl border border-slate-300 px-4 py-3 outline-none ring-0 focus:border-slate-500"
                  type="email"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                />
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium text-slate-700">
                  パスワード
                </label>
                <input
                  className="w-full rounded-xl border border-slate-300 px-4 py-3 outline-none ring-0 focus:border-slate-500"
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                />
              </div>

              <div className="flex gap-3">
                <button
                  className="rounded-xl bg-slate-900 px-5 py-3 text-sm font-semibold text-white hover:bg-slate-700"
                  type="submit"
                >
                  ログイン
                </button>

                <button
                  className="rounded-xl border border-slate-300 px-5 py-3 text-sm font-semibold text-slate-700 hover:bg-slate-100"
                  type="button"
                  onClick={logout}
                >
                  ログアウト
                </button>
              </div>
            </form>

            <div className="mt-6 rounded-2xl bg-slate-50 p-4 text-sm">
              <div className="font-semibold text-slate-800">動作確認用アカウント</div>
              <div className="mt-2 text-slate-600">email: test@example.com</div>
              <div className="text-slate-600">password: password123</div>
            </div>
          </section>

          <section className="rounded-3xl bg-white p-6 shadow-sm ring-1 ring-slate-200">
            <h2 className="mb-4 text-xl font-semibold">ログイン中ユーザー</h2>

            {!loggedIn && (
              <p className="text-sm text-slate-500">未ログインです。</p>
            )}

            {loggedIn && me && (
              <div className="space-y-2 text-sm">
                <div>
                  <span className="font-semibold">ID:</span> {me.id}
                </div>
                <div>
                  <span className="font-semibold">名前:</span> {me.name}
                </div>
                <div>
                  <span className="font-semibold">メール:</span> {me.email}
                </div>
              </div>
            )}

            {message && (
              <div className="mt-6 whitespace-pre-line rounded-2xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-700">
                {message}
              </div>
            )}
          </section>
        </div>

        <section className="mt-6 rounded-3xl bg-white p-6 shadow-sm ring-1 ring-slate-200">
          <h2 className="mb-4 text-xl font-semibold">商品検索</h2>

          <form className="flex flex-col gap-3 sm:flex-row" onSubmit={searchProducts}>
            <input
              className="flex-1 rounded-xl border border-slate-300 px-4 py-3 outline-none ring-0 focus:border-slate-500"
              type="text"
              value={searchWord}
              onChange={(event) => setSearchWord(event.target.value)}
              placeholder="例: 防水 軽量"
            />
            <button
              className="rounded-xl bg-blue-600 px-5 py-3 text-sm font-semibold text-white hover:bg-blue-500"
              type="submit"
            >
              検索
            </button>
          </form>

          <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {products.map((product) => (
              <article
                key={product.id}
                className="rounded-2xl border border-slate-200 bg-slate-50 p-4"
              >
                <div className="text-xs font-semibold tracking-wide text-slate-500">
                  {product.sku}
                </div>
                <h3 className="mt-2 text-lg font-semibold">{product.name}</h3>
                <p className="mt-3 text-sm text-slate-600">
                  ¥{Number(product.price).toLocaleString("ja-JP")}
                </p>
              </article>
            ))}
          </div>
        </section>
      </div>
    </div>
  );
}
