FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main cmd/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /app/main .

# Profile uploads are written to ./public/profiles at runtime, so the
# directory must exist and be writable by the non-root user.
RUN mkdir -p /app/public/profiles \
    && chown -R appuser:appgroup /app/public

USER appuser

EXPOSE 8080

CMD ["./main"]
