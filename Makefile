.PHONY: help dev build test clean deploy k8s-deploy terraform-init terraform-plan terraform-apply docker-build docker-push start

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

start: ## Quick start - Run the entire stack locally
	@echo "Starting OffGridFlow..."
	@chmod +x scripts/dev-start.sh
	@./scripts/dev-start.sh

dev: ## Start local development environment
	docker-compose up

dev-build: ## Build and start local development environment
	docker-compose up --build

dev-down: ## Stop local development environment
	docker-compose down

dev-clean: ## Clean local development environment (removes volumes)
	docker-compose down -v

test: ## Run all tests
	go test -v -race -coverprofile=coverage.out ./...
	cd web && npm test

lint: ## Run linters
	go fmt ./...
	go vet ./...
	cd web && npm run lint

build: ## Build binaries
	CGO_ENABLED=0 go build -o bin/api ./cmd/api
	CGO_ENABLED=0 go build -o bin/worker ./cmd/worker

docker-build: ## Build Docker images
	docker build -t offgridflow-api:latest .
	docker build --target worker -t offgridflow-worker:latest .
	docker build -t offgridflow-web:latest ./web

docker-push: ## Push Docker images to registry
	docker tag offgridflow-api:latest ghcr.io/example/offgridflow-api:latest
	docker tag offgridflow-worker:latest ghcr.io/example/offgridflow-worker:latest
	docker tag offgridflow-web:latest ghcr.io/example/offgridflow-web:latest
	docker push ghcr.io/example/offgridflow-api:latest
	docker push ghcr.io/example/offgridflow-worker:latest
	docker push ghcr.io/example/offgridflow-web:latest

k8s-deploy: ## Deploy to Kubernetes
	kubectl apply -f infra/k8s/namespace.yaml
	kubectl apply -f infra/k8s/configmap.yaml
	kubectl apply -f infra/k8s/services.yaml
	kubectl apply -f infra/k8s/api-deployment.yaml
	kubectl apply -f infra/k8s/worker-deployment.yaml
	kubectl apply -f infra/k8s/web-deployment.yaml
	kubectl apply -f infra/k8s/hpa.yaml
	kubectl apply -f infra/k8s/ingress.yaml

k8s-status: ## Check Kubernetes deployment status
	kubectl get pods,svc,hpa -n offgridflow

k8s-logs-api: ## Tail API logs
	kubectl logs -f deployment/offgridflow-api -n offgridflow

k8s-logs-worker: ## Tail worker logs
	kubectl logs -f deployment/offgridflow-worker -n offgridflow

terraform-init: ## Initialize Terraform
	cd infra/terraform && terraform init

terraform-plan: ## Plan Terraform changes
	cd infra/terraform && terraform plan

terraform-apply: ## Apply Terraform changes
	cd infra/terraform && terraform apply

terraform-destroy: ## Destroy Terraform infrastructure
	cd infra/terraform && terraform destroy

migrate-up: ## Run database migrations
	./bin/api migrate up

migrate-down: ## Rollback database migration
	./bin/api migrate down

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf coverage.out
	go clean
