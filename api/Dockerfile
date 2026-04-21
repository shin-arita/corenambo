FROM golang:1.25-bookworm

WORKDIR /app

RUN go install github.com/air-verse/air@v1.63.0

COPY go.mod ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
