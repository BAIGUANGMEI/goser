# GoSer Makefile

.PHONY: all build clean frontend daemon cli app dev

DIST = dist
FRONTEND_DIR = cmd/app/frontend

all: build

# Build everything
build: frontend daemon cli app

# Build frontend only
frontend:
	cd $(FRONTEND_DIR) && npm install && npx vite build

# Build daemon
daemon:
	go build -ldflags="-s -w" -o $(DIST)/goserd.exe ./cmd/goserd

# Build CLI
cli:
	go build -ldflags="-s -w" -o $(DIST)/goser.exe ./cmd/goser

# Build GUI app (requires Wails build tags)
app: frontend
	go build -tags "desktop,production" -ldflags="-s -w -H=windowsgui" -o $(DIST)/goser-app.exe ./cmd/app

# Dev mode - run daemon
dev-daemon:
	go run ./cmd/goserd

# Dev mode - run CLI
dev-cli:
	go run ./cmd/goser $(ARGS)

# Clean
clean:
	rm -rf $(DIST)
	rm -rf $(FRONTEND_DIR)/node_modules
	rm -rf $(FRONTEND_DIR)/dist

# Install Go dependencies
deps:
	go mod tidy
