# corenambo

## セットアップ

```bash
git clone https://github.com/shin-arita/corenambo.git
cd corenambo

cp .env.example .env

docker compose up -d --build

起動確認
Frontend: http://localhost:5173
API: http://localhost:8080/health
Mail: http://localhost:8025

