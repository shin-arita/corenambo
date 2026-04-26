up:
	docker compose up -d --build

down:
	docker compose down

destroy:
	docker compose down -v

logs:
	docker compose logs -f

ps:
	docker compose ps

db:
	docker compose exec db psql -U app_user -d app_db

api:
	docker compose exec api sh

front:
	docker compose exec frontend sh

migrate-up:
	docker compose exec api sh -lc 'migrate -path /db/migrations -database "$$DATABASE_URL" up'
