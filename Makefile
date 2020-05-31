GO     ?= CGO_ENABLED=0 go
DOCKER ?= docker
DC     ?= docker-compose
DC     := $(DC) --compatibility

APP ?= promdiscovery
IMG ?= sgoroshko/$(APP)
VSN ?= $(shell cat VERSION)

MAIN_GO := main.go
LDFLAGS := "-s -w -X main.VSN=$(VSN)"
PKGS    := $(shell go list ./...)
TARGET  := " ----> [$@]"

.PHONY: all
all: help

## help: print this message
.PHONY: help
help: Makefile
	@echo
	@echo 'Usage: make <TARGETS> ... <OPTIONS>'
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
	@echo
	@echo 'By default print this message.'
	@echo

## clean: remove output binary
.PHONY: clean
clean:
	@echo $(TARGET)
	rm -rf $(APP)

## deps: download dependencies
.PHONY: deps
deps:
	@echo $(TARGET)
	$(GO) mod tidy
	$(GO) mod download

## fmt: running "go fmt" on sources packages
.PHONY: fmt
fmt:
	@echo $(TARGET)
	@$(GO) fmt $(PKGS)

## vet: running "go vet" on sources packages
.PHONY: vet
vet: fmt
	@echo $(TARGET)
	@$(GO) vet -composites=false $(PKGS)

## tests: running "go test" on sources packages
.PHONY: tests
tests: fmt vet
	@echo $(TARGET)
	@$(GO) test $(PKGS) -count 3

## build: compile packages and dependencies
.PHONY: build
build: clean tests
	@echo $(TARGET)
	$(GO) build -ldflags $(LDFLAGS) -o $(APP) $(MAIN_GO)

## image: build docker image
.PHONY: image
image:
	@echo $(TARGET)
	$(DOCKER) build . --tag $(IMG):latest --tag $(IMG):$(VSN)

## docker-up:
docker-up:
	@echo $(TARGET)
	$(DC) up -d $(ARGS)

## docker-stop:
docker-stop:
	@echo $(TARGET)
	$(DC) stop
