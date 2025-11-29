.PHONY: help auth-gcloud auth-gcloud-ad auth-firebase auth-cloudflare auth-all export-cloudflare-token go-test go-lint go-fmt go-cover all

all: go-fmt go-lint go-test go-cover

help:
	@echo "Available targets:"
	@echo "  make (default)        - Run go vet and go test."
	@echo "  make auth-gcloud      - Authenticate gcloud CLI (user credentials)."
	@echo "  make auth-gcloud-ad   - Authenticate gcloud application default credentials."
	@echo "  make auth-firebase    - Authenticate Firebase CLI."
	@echo "  make auth-cloudflare  - Authenticate Cloudflare via cloudflared."
	@echo "  make auth-all         - Run all authentication targets."
	@echo "  make export-cloudflare-token TOKEN=<token> - Export Cloudflare API token in this shell."
	@echo "  make go-test          - Run Go unit tests."
	@echo "  make go-lint          - Run Go static analysis (go vet)."
	@echo "  make go-fmt           - Format Go sources with gofmt."
	@echo "  make go-cover         - Run Go tests with coverage and generate coverage.out."

auth-gcloud:
	@echo "Launching gcloud auth flow..."
	gcloud auth login

auth-gcloud-ad:
	@echo "Launching gcloud application-default auth flow..."
	gcloud auth application-default login

auth-firebase:
	@echo "Launching Firebase auth flow..."
	firebase login

auth-cloudflare:
	@echo "Launching Cloudflare auth flow..."
	cloudflared tunnel login

auth-all: auth-gcloud auth-gcloud-ad auth-firebase auth-cloudflare

export-cloudflare-token:
ifndef TOKEN
	$(error TOKEN is undefined. Call as: make export-cloudflare-token TOKEN=your_api_token)
endif
	@echo "export CLOUDFLARE_API_TOKEN=$(TOKEN)" >> .env.cloudflare
	@echo "Token stored in .env.cloudflare. Source it with: source .env.cloudflare"

go-test:
	@echo "Running Go unit tests..."
	go test ./...

go-lint:
	@echo "Running go vet..."
	go vet ./...

go-fmt:
	@echo "Running gofmt..."
	gofmt -w $$(find . -name '*.go' -not -path './node_modules/*' -not -path './.git/*')

go-cover:
	@echo "Running Go tests with coverage..."
	go test ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out

