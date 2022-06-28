GO_FLAGS   ?=
NAME       := mirage
OUTPUT_BIN ?= bin/${NAME}
PACKAGE    := github.com/sbchaos/$(NAME)
#GIT_REV    ?= $(shell git rev-parse --short HEAD)
SOURCE_DATE_EPOCH ?= $(shell date +%s)
DATE       ?= $(shell date -u -d @${SOURCE_DATE_EPOCH} +"%Y-%m-%dT%H:%M:%SZ")
#VERSION    ?= "$(shell git describe --tags)"
IMG_NAME   := sbchaos/mirage
IMAGE      := ${IMG_NAME}:${VERSION}

default: build

test:   ## Run all tests
	@go test ./...

cover:  ## Run test coverage suite
	@go test ./... --coverprofile=cov.out
	@go tool cover --html=cov.out

build:
	@go build -o ${OUTPUT_BIN} .

build-later:  ## Builds the CLI
	@go build ${GO_FLAGS} \
	-ldflags "-w -s -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT_REV} -X ${PACKAGE}/cmd.date=${DATE}" \
	-o ${OUTPUT_BIN} .

img:    ## Build Docker Image
	@docker build --rm -t ${IMG_NAME} .
