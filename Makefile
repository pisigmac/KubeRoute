# Kuberoute AI — Build System
.PHONY: all proto build test lint clean docker-up docker-down kind-up kind-down start stop restart status logs doctor seed

start:
	./scripts/start_all.sh

stop:
	./scripts/stop_all.sh

restart:
	./scripts/restart_all.sh

status:
	./scripts/status.sh

logs:
	./scripts/logs.sh $(SERVICE)

doctor:
	./scripts/doctor.sh

seed:
	./scripts/seed_data.sh

all: proto build

# Protocol Buffers
proto:
	protoc --proto_path=packages/shared-types/proto --go_out=services/orchestrator/proto --go_opt=paths=source_relative --go-grpc_out=services/orchestrator/proto --go-grpc_opt=paths=source_relative agent.proto api.proto
	protoc --proto_path=packages/shared-types/proto --go_out=services/agent/proto --go_opt=paths=source_relative --go-grpc_out=services/agent/proto --go-grpc_opt=paths=source_relative agent.proto api.proto

# Build all services
build:
	cd services/orchestrator && go build -o ../../bin/orchestrator ./cmd
	cd services/agent && go build -o ../../bin/agent ./cmd
	cd cli && go build -o ../bin/kr ./main.go
	cd services/ai-engine && uv pip install -e .

# Test all services
test:
	cd services/orchestrator && go test -race -coverprofile=coverage.out ./...
	cd services/agent && go test -race -coverprofile=coverage.out ./...
	cd cli && go test -race -coverprofile=coverage.out ./...
	cd services/ai-engine && pytest --cov=ai_engine --cov-report=term-missing tests/

# Lint
lint:
	cd services/orchestrator && golangci-lint run
	cd services/agent && golangci-lint run
	cd cli && golangci-lint run
	cd services/ai-engine && ruff check .

# Docker Compose (local dev)
docker-up:
	docker-compose -f docker-compose.dev.yml up -d --build

docker-down:
	docker-compose -f docker-compose.dev.yml down -v

# Kind cluster
kind-up:
	kind create cluster --name kuberoute-dev --config infra/k8s/kind-config.yaml
	kubectl apply -f infra/k8s/namespace.yaml
	kubectl apply -f services/agent/manifests/

kind-down:
	kind delete cluster --name kuberoute-dev

# Clean
clean:
	rm -rf bin/
	rm -f services/orchestrator/coverage.out
	rm -f services/agent/coverage.out
	rm -f cli/coverage.out
