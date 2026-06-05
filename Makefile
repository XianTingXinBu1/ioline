.PHONY: help dev stop status dev-backend stop-backend dev-reset dev-frontend build build-backend build-frontend check check-backend check-frontend test-backend test-backend-smoke

help:
	@echo "Available targets:"
	@echo "  make dev            # start backend and frontend dev services"
	@echo "  make stop           # stop backend and frontend dev services"
	@echo "  make status         # show backend and frontend dev status"
	@echo "  make dev-backend    # start backend dev server only"
	@echo "  make stop-backend   # stop managed backend dev server"
	@echo "  make dev-reset      # clean backend dev processes and pid files"
	@echo "  make dev-frontend   # start frontend dev server only"
	@echo "  make build          # build backend and frontend"
	@echo "  make build-backend  # build backend binary to bin/ioline-server"
	@echo "  make build-frontend # build frontend assets"
	@echo "  make check          # run conservative syntax/build checks"
	@echo "  make check-backend  # run go build check"
	@echo "  make check-frontend # run frontend build check"
	@echo "  make test-backend   # run backend unit and handler tests"
	@echo "  make test-backend-smoke # run backend smoke test script against a running dev server"

dev:
	bash scripts/dev-start.sh

stop:
	bash scripts/dev-stop.sh

status:
	bash scripts/dev-status.sh

dev-backend:
	bash scripts/dev-backend.sh

stop-backend:
	@if [ -f .tmp/ioline-backend.pid ]; then \
		PID=$$(cat .tmp/ioline-backend.pid); \
		kill $$PID 2>/dev/null || true; \
		rm -f .tmp/ioline-backend.pid; \
		echo "[ioline] stopped managed backend $$PID"; \
	else \
		echo "[ioline] no managed backend pid file found"; \
	fi

dev-reset:
	-pkill -f 'go run ./apps/server' || true
	-pkill -f '/data/data/.*/project/ioline/server' || true
	-pkill -f '/data/data/.*/project/ioline/bin/ioline-server' || true
	-pkill -f "sh -c \(trap 'kill 0' INT TERM EXIT;.*make dev-backend.*make dev-frontend.*wait\)" || true
	-fuser -k -n tcp 9650 >/dev/null 2>&1 || true
	-rm -f .tmp/ioline-backend.pid .tmp/ioline-backend.log .tmp/dev-backend.out .tmp/dev-backend.wrapper.pid
	@echo "[ioline] backend dev processes and pid files cleaned"

dev-frontend:
	bash scripts/dev-frontend.sh

build: build-backend build-frontend

build-backend:
	mkdir -p bin
	go build -o ./bin/ioline-server ./apps/server

build-frontend:
	cd web && npm run build

check: check-backend check-frontend

check-backend:
	go build ./apps/server

check-frontend:
	cd web && npm run build

test-backend:
	go test ./...

test-backend-smoke:
	bash scripts/test_backend.sh
