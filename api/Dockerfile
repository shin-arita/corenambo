FROM golang:1.25-bookworm

WORKDIR /app

RUN apt-get update && apt-get install -y \
    ca-certificates \
    git \
    curl \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/air-verse/air@v1.63.0
RUN go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]