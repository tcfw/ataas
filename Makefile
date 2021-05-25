GOBUILD := CGO_ENABLED=0 go build
BUILD_DIR := ./bin
BUILD_FLAGS := -ldflags="-s -w" -trimpath

build:
	@mkdir -p ./bin
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/ataas ./.

.PHONY: run
run:
	@go run ./.

.PHONY: protos
protos:
	@./scripts/genproto.sh

docker:
	docker build -t tcfw/ataas .