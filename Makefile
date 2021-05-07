GOBUILD := CGO_ENABLED=0 go build
BUILD_DIR := ./bin
BUILD_FLAGS := -ldflags="-s -w" -trimpath

build:
	@mkdir -p ./bin
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/trader ./.

.PHONY: run
run:
	@go run ./.

.PHONY: protos
protos:
	@./scripts/genproto.sh