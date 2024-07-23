timestamp:=$(shell date +%s)
CTR_REGISTRY ?= localhost:5000
IMAGE_TAG ?= kind-$(timestamp)
my-app-controller-image=$(CTR_REGISTRY)//my-app-controller:$(IMAGE_TAG)

.PHONY:
build-dist:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags='-s' -o=dist/my-app-controller ./cmds/my-app-controller

.PHONY:
docker-build: build-dist docker-build-only

.PHONY:
docker-push: docker-build docker-push-no-build

.PHONY:
docker-build-only:
	docker build -t $(my-app-controller-image) -f docker/Dockerfile.my-app-controller .

docker-push-no-build: docker-build-only
	docker push $(my-app-controller-image)

.PHONY: lint-all lint-go lint-yaml lint-k8s
lint-all: lint-go lint-yaml lint-k8s

lint-go: install-golangci-lint
	golangci-lint run --timeout 5m

.PHONY:
install-golangci-lint:
	which golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.56.2

.PHONY:
install-kube-linter:
	which kube-linter || go install golang.stackrox.io/kube-linter/cmd/kube-linter@latest

.PHONY:
install-kind:
	which kind || go install sigs.k8s.io/kind@v0.23.0

kind-up: install-kind
	./scripts/kind-with-registry.sh

kind-down:
	kind delete cluster --name my-app

deploy-kind: kind-up docker-push
	kubectl apply --context kind-my-app -f configs/...