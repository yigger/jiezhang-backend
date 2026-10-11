SERVER_ARCH ?= amd64

.PHONY: test build build-server build-import swagger vet check

test:
	go test ./...

build:
	go build -o bin/jiezhang-server .
	$(MAKE) build-import

build-import:
	go build -o bin/bill-import ./cmd/bill-import

build-server:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(SERVER_ARCH) go build -o bin/linux-$(SERVER_ARCH)/jiezhang-backend .

swagger:
	go tool swag init --parseInternal -g main.go -o docs/swagger

vet:
	go vet ./...

check: swagger
	$(MAKE) test vet build
