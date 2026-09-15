.PHONY: test build swagger vet check

test:
	go test ./...

build:
	go build -o bin/jiezhang-server .

swagger:
	go tool swag init --parseInternal -g main.go -o docs/swagger

vet:
	go vet ./...

check: swagger
	$(MAKE) test vet build
