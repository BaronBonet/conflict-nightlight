export VERSION := $(shell ./scripts/get-version.sh)

OUT_DIR=out
ARCH=arm64
OS=linux

guard-%:
	@ if [ "${${*}}" = "" ]; then \
        echo "Variable $* not set"; \
        exit 1; \
    fi

generate-proto-%:
	@echo "Generating $* protobuf stubs"
	@buf generate --template "proto/buf.gen.$*.yaml"


.PHONY: build-go-aws
build-go-aws: guard-APP generate-go
	@echo "Building $(APP) for AWS version ${VERSION}"
	@cd ./lambdas/go/ \
		go get ./... && \
		GOOS=${OS} GOARCH=${ARCH} go build -ldflags \
		 "-X 'github.com/BaronBonet/conflict-nightlight/internal/infrastructure.Version=${VERSION}'" \
		 -o ${OUT_DIR}/${APP}/handler/bootstrap cmd/${APP}/*go
	@zip -jrm "./lambdas/go/${OUT_DIR}/$(APP)/handler/main.zip" "./lambdas/go/${OUT_DIR}/$(APP)/handler/"*
	@echo "Done building $(APP) for AWS."

.PHONY: generate-go
generate-go:
	@rm -rf ./lambdas/go/generated
	@$(MAKE) generate-proto-go
	@cd  ./lambdas/go/ && go generate ./...
	@echo "Done generating go"

.PHONY: generate-python
generate-python:
	@rm -rf lambdas/python/generated
	@$(MAKE) generate-proto-python
	@echo "Done generating python"
