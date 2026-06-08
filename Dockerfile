# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/pos-saas-backend \
    ./cmd/api


FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Jakarta

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/pos-saas-backend /app/pos-saas-backend

RUN mkdir -p /app/uploads/products \
    && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["/app/pos-saas-backend"]