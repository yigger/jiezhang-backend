SERVER_ARCH ?= amd64

.PHONY: test build build-server swagger vet check

test:
	go test ./...

build:
	go build -o bin/jiezhang-server .

build-server:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(SERVER_ARCH) go build -o bin/linux-$(SERVER_ARCH)/jiezhang-backend .

swagger:
	go tool swag init --parseInternal -g main.go -o docs/swagger

vet:
	go vet ./...

check: swagger
	$(MAKE) test vet build
