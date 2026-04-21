#!/bin/bash

set -e

echo "Creating project structure..."

mkdir -p corenambo/{frontend,api/cmd/server,db/init}

# ----------------------------
# docker-compose.yml
# ----------------------------
cat << 'EOT' > corenambo/docker-compose.yml
services:
  frontend:
    build: ./frontend
    ports:
      - "5173:5173"
    volumes:
      - ./frontend:/app
      - node_modules:/app/node_modules
    working_dir: /app
    command: sh -c "npm ci && npm run dev -- --host 0.0.0.0"

  api:
    build: ./api
    ports:
      - "8080:8080"
    volumes:
      - ./api:/app
    working_dir: /app
    command: sh -c "go mod init app-api || true && go run cmd/server/main.go"
    depends_on:
      - db

  db:
    image: postgres:16
    environment:
      POSTGRES_DB: app_db
      POSTGRES_USER: app_user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - ./db/init:/docker-entrypoint-initdb.d

volumes:
  node_modules:
EOT

# ----------------------------
# frontend package.json
# ----------------------------
cat << 'EOT' > corenambo/frontend/package.json
{
  "name": "frontend",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite"
  },
  "dependencies": {
    "react": "^18.0.0",
    "react-dom": "^18.0.0"
  },
  "devDependencies": {
    "vite": "^5.0.0"
  }
}
EOT

# ----------------------------
# API main.go
# ----------------------------
cat << 'EOT' > corenambo/api/cmd/server/main.go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
EOT

# ----------------------------
# DB init
# ----------------------------
cat << 'EOT' > corenambo/db/init/01_extensions.sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
EOT

echo "Setup complete!"
echo "Next:"
echo "cd corenambo"
echo "docker compose up -d --build"

