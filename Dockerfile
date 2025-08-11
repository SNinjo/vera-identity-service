FROM golang:1.23.0-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/...

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/api ./api

RUN chown -R appuser:appgroup /app
USER appuser

ENV GIN_MODE=release

EXPOSE 8080
CMD ["./main"]
