# Image settings
IMAGE_NAME ?= aplulu/hakoniwa
IMAGE_TAG ?= latest

# Platforms for multi-arch build
PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: build
build:
	go build -o bin/hakoniwa cmd/serve/main.go

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf ui/dist/

# Docker commands
.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) -f docker/app/Dockerfile .

.PHONY: docker-buildx
docker-buildx:
	# Ensure builder exists: docker buildx create --use
	docker buildx build \
		--platform $(PLATFORMS) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-f docker/app/Dockerfile \
		.

.PHONY: docker-push
docker-push:
	docker buildx build \
		--platform $(PLATFORMS) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-f docker/app/Dockerfile \
		--push \
		.
