FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /main ./cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /main .
COPY --from=builder /app/serviceAccountKey.json .
COPY --from=builder /app/.env .

EXPOSE 8080

CMD ["./main"]