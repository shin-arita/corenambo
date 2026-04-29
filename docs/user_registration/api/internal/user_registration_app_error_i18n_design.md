# ユーザ仮登録 app_error / i18n 設計

## 1. 概要

本設計は、ユーザ仮登録機能における app_error / i18n の責務分割を定義する。

APIレスポンスは `code + message` 形式で統一し、
表示文言とエラー構造を分離する。

---

## 2. 全体方針

- エラー判定は `code` ベースで行う
- 表示文言は i18n で解決する
- service は message を持たない
- handler が最終レスポンスを構築する

---

## 3. レスポンス仕様

### 正常系

```json
{
  "code": "USER_REGISTRATION_REQUEST_CREATED",
  "message": "仮登録メールを送信しました。メールをご確認ください。"
}
```

---

### 異常系

```json
{
  "code": "BAD_REQUEST",
  "message": "リクエストが不正です。"
}
```

---

### バリデーション

```json
{
  "code": "VALIDATION_ERROR",
  "message": "入力内容に誤りがあります。",
  "errors": {
    "email": [
      {
        "code": "EMAIL_REQUIRED",
        "message": "メールアドレスを入力してください。"
      }
    ]
  }
}
```

---

## 4. 注意事項（実装準拠）

- USER_ALREADY_REGISTERED は APIレスポンスでは使用しない
- メール存在は外部に漏らさない
- 常に成功レスポンスを返す

---

## 5. i18n

- ja / en をサポート
- 未指定は ja
- 未定義 code は code を返す

---

## 6. まとめ

- code 主体
- message は i18n
- service は message を持たない
