.PHONY: help dev dev-backend dev-frontend build build-backend build-frontend check check-backend check-frontend

help:
	@echo "Available targets:"
	@echo "  make dev            # start backend and frontend dev servers"
	@echo "  make dev-backend    # start backend dev server"
	@echo "  make dev-frontend   # start frontend dev server"
	@echo "  make build          # build backend and frontend"
	@echo "  make build-backend  # build backend binary to bin/ioline-server"
	@echo "  make build-frontend # build frontend assets"
	@echo "  make check          # run conservative syntax/build checks"
	@echo "  make check-backend  # run go build check"
	@echo "  make check-frontend # run frontend build check"

dev:
	@echo "Starting backend and frontend..."
	@(trap 'kill 0' INT TERM EXIT; \
		$(MAKE) dev-backend & \
		$(MAKE) dev-frontend & \
		wait)

dev-backend:
	go run ./apps/server

dev-frontend:
	cd web && npm run dev -- --host 0.0.0.0

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
