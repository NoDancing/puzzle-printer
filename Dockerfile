# ---- build stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o puzzle-printer ./cmd/puzzle-printer

# ---- runtime stage ----
FROM alpine:3.19

# cups-client provides lp for IPP network printing
# dcron is a lightweight cron daemon
# tzdata allows TZ env var to set the container timezone
RUN apk add --no-cache cups-client dcron tzdata

COPY --from=builder /app/puzzle-printer /usr/local/bin/puzzle-printer
COPY docker/crontab /etc/crontabs/root

# Logs live here; mount a volume to persist them
RUN mkdir -p /var/log/puzzle-printer
ENV TZ=America/New_York

# Run cron in the foreground
CMD ["crond", "-f", "-l", "2"]
