# ---- build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o puzzle-printer ./cmd/puzzle-printer

# ---- runtime stage ----
FROM alpine:3.19

# cups-client provides lp for IPP network printing
# supercronic is a container-friendly cron daemon (no setpgid required)
# tzdata allows TZ env var to set the container timezone
RUN apk add --no-cache cups-client ghostscript openssh-client tzdata wget && \
    ARCH=$(uname -m) && \
    [ "$ARCH" = "x86_64" ] && SA=amd64 || SA=arm64 && \
    wget -qO /usr/local/bin/supercronic \
      "https://github.com/aptible/supercronic/releases/download/v0.2.33/supercronic-linux-${SA}" && \
    chmod +x /usr/local/bin/supercronic

COPY --from=builder /app/puzzle-printer /usr/local/bin/puzzle-printer
COPY docker/crontab /etc/crontabs/root

# Logs live here; mount a volume to persist them
RUN mkdir -p /var/log/puzzle-printer /run/secrets
ENV TZ=America/New_York

# Run supercronic in the foreground
CMD ["supercronic", "/etc/crontabs/root"]
