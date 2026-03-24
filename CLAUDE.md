# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Local (macOS)
```bash
make build    # compile binary to ./puzzle-printer
make install  # install to $GOPATH/bin
make run      # go run (no compile step)
make test     # run all tests
make launchd  # install + load the macOS launchd job (runs at 7am)
```

### Docker (NAS)
```bash
make docker-build   # build the image
make docker-up      # start container in background (cron runs at 7am)
make docker-down    # stop container
make docker-logs    # tail container logs
make docker-run     # trigger a one-off run right now
```

## Architecture

A Go CLI that logs into the NYT, fetches the daily crossword PDF, and sends it to a printer. No PDF rendering — NYT serves the file directly.

### Flow

1. **Config** (`internal/config/config.go`) — credentials loaded in priority order:
   1. Environment variables (`NYT_EMAIL`, `NYT_PASSWORD`, `PRINTER`) — used by Docker
   2. macOS Keychain (`security` CLI) — macOS only, fails gracefully elsewhere
   3. `op://` references in config file, resolved via 1Password CLI
   4. Plaintext values in `~/.config/puzzle-printer/config.toml`
2. **Auth** (`internal/nyt/client.go`) — POSTs email/password to `myaccount.nytimes.com/svc/ios/v2/login`, gets a `NYT-S` token used as a cookie on all subsequent requests
3. **Fetch** (`internal/nyt/puzzle.go`) — two paths based on day of week:
   - **Sunday**: constructs a date-slug URL (`/svc/crosswords/v2/puzzle/print/MonDDYY.pdf`) for the larger print edition
   - **Mon–Sat**: fetches daily metadata JSON to get puzzle ID, then downloads `/svc/crosswords/v2/puzzle/<id>.pdf`
4. **Print** (`internal/print/print.go`) — writes PDF to a temp file, calls `lp` (or `open -a Preview` for `--no-print`)

### CLI flags

| Flag | Description |
|------|-------------|
| `--date YYYY-MM-DD` | Fetch a past (or future) date instead of today |
| `--output FILE.pdf` | Save PDF to path; skip printing |
| `--no-print` | Open in Preview instead of sending to printer (macOS only) |

## Docker setup

The container runs `dcron` with `cups-client` for IPP network printing. Cron schedule is in `docker/crontab` (default: 7:00 AM, timezone from `TZ` env var).

Credentials and printer go in `.env` (never committed — see `.env.example`):
```
NYT_EMAIL=you@example.com
NYT_PASSWORD=yourpassword
PRINTER=ipp://192.168.1.x/ipp/print
```

Logs are written to `./logs/` on the host via a volume mount.

### launchd (macOS alternative)

`launchd/com.puzzle-printer.plist` runs the binary daily at 7:00 AM. Binary path is hardcoded to `/usr/local/bin/puzzle-printer` — update if `go env GOPATH` differs.
