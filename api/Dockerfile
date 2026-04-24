FROM golang:1.25-bookworm

WORKDIR /app

RUN apt-get update && apt-get install -y \
    ca-certificates \
    git \
    curl \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/air-verse/air@v1.63.0 && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1 && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 && \
    cp /go/bin/air /usr/local/bin/air && \
    cp /go/bin/migrate /usr/local/bin/migrate && \
    cp /go/bin/golangci-lint /usr/local/bin/golangci-lint

COPY go.mod ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]