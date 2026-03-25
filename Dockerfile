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
# tzdata allows TZ env var to set the container timezone
RUN apk add --no-cache cups-client ghostscript openssh-client tzdata

COPY --from=builder /app/puzzle-printer /usr/local/bin/puzzle-printer
COPY docker/entrypoint.sh /entrypoint.sh

# Logs live here; mount a volume to persist them
RUN mkdir -p /var/log/puzzle-printer /run/secrets
ENV TZ=America/New_York

CMD ["/entrypoint.sh"]
