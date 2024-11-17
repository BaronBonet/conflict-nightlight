export VERSION := $(shell ./scripts/get-version.sh)

OUT_DIR=out

.PHONY: build-cli, build-go-aws, generate-go, generate-python, dependencies-install-go, dependencies-check-go, test-go

guard-%:
	@ if [ "${${*}}" = "" ]; then \
        echo "Variable $* not set"; \
        exit 1; \
    fi

generate-proto-%:
	@buf generate --template "proto/buf.gen.$*.yaml"
	@echo "Done generating $* protobuf stubs"


build-cli: generate-go
	@echo "Building CLI version ${VERSION}"
	@cd ./lambdas/go/ \
		go get ./... && \
		go build -ldflags "-X 'github.com/BaronBonet/conflict-nightlight/internal/infrastructure.Version=${VERSION}'" \
		-o ${OUT_DIR}/map-controller cmd/cli/*go
	@mv ./lambdas/go/${OUT_DIR}/map-controller .
	@chmod +x map-controller
	@echo "Done building CLI."

build-go-aws: guard-APP generate-go
	@echo "Building $(APP) for AWS version ${VERSION}"
	@cd ./lambdas/go/ \
		go get ./... && \
		GOOS=linux GOARCH=arm64 go build -ldflags \
		 "-X 'github.com/BaronBonet/conflict-nightlight/internal/infrastructure.Version=${VERSION}'" \
		 -o ${OUT_DIR}/${APP}/handler/bootstrap cmd/${APP}/*go
	@zip -jrm "./lambdas/go/${OUT_DIR}/$(APP)/handler/main.zip" "./lambdas/go/${OUT_DIR}/$(APP)/handler/"*
	@echo "Done building $(APP) for AWS."

dependencies-install-go:
	@./lambdas/go/scripts/dependencies/install.sh

dependencies-check-go:
	@./lambdas/go/scripts/dependencies/check.sh

generate-go: dependencies-check-go
	@rm -rf ./lambdas/go/generated
	@$(MAKE) generate-proto-go
	@cd ./lambdas/go/ && export PATH=$$(pwd)/.local/bin::$(PATH); go generate ./...
	@echo "Done generating go"

generate-python:
	@rm -rf lambdas/python/generated
	@$(MAKE) generate-proto-python
	@echo "Done generating python"

test-go:
	@cd lambdas/go/ && go test ./...


