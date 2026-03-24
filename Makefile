.PHONY: build install run test launchd docker-build docker-up docker-down docker-logs

# --- local (macOS) ---

build:
	go build ./cmd/puzzle-printer

install:
	go install ./cmd/puzzle-printer

run:
	go run ./cmd/puzzle-printer

test:
	go test ./...

launchd:
	cp launchd/com.puzzle-printer.plist ~/Library/LaunchAgents/
	launchctl load ~/Library/LaunchAgents/com.puzzle-printer.plist
	@echo "Installed. To test: launchctl start com.puzzle-printer"

# --- docker (NAS) ---

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# Run the puzzle-printer once right now inside a temporary container
docker-run:
	docker compose run --rm puzzle-printer puzzle-printer
