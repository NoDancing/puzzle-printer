# puzzle-printer

Fetches the NYT daily crossword PDF and sends it to a printer. Runs on a schedule via macOS launchd or Docker (for a NAS).

## Setup

Copy `.env.example` to `.env` and fill in your credentials:

```
NYT_S=your_nyt_s_cookie_value
PRINTER=ipp://192.168.1.x/ipp/print
```

**Getting your NYT-S cookie:** Log into nytimes.com in your browser, open DevTools → Application → Cookies → `nytimes.com`, find the `NYT-S` cookie and copy its value.

**Getting your printer's IPP URI:** Run `lpstat -v` on macOS to find the IP, then the URI is typically `ipp://<printer-ip>/ipp/print`.

> **Note:** The NYT login API is blocked by Cloudflare, so email/password auth doesn't work. The `NYT-S` cookie is the only supported auth method.

## Usage

```bash
# Print today's puzzle
puzzle-printer

# Print a specific date
puzzle-printer --date 2025-12-21

# Save to file instead of printing
puzzle-printer --output puzzle.pdf

# Open in Preview instead of printing (macOS)
puzzle-printer --no-print
```

## Running on a schedule

**macOS** — runs daily at 7am via launchd:
```bash
make launchd
```

**Docker / NAS** — runs daily at 7am via cron:
```bash
make docker-up
```

## Requirements

- NYT subscription with a browser-accessible `NYT-S` cookie
- A printer that supports IPP (most modern network printers do)
- Docker (for NAS deployment) or Go 1.26+ (for local builds)

## Docker notes

Credentials go in `.env` (never committed — see `.env.example`). Logs are written to `./logs/` via a volume mount. The timezone for the cron schedule is set via the `TZ` environment variable (default: `America/New_York`).

The container converts PDF to URF (AirPrint raster format) using Ghostscript before sending to the printer, since most IPP printers don't accept PDF directly.
