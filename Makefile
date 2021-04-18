.PHONY: all
all: test show-cover

.PHONY: test
test:
	go test -coverprofile=coverage.out -v ./...

.PHONY: show-cover
show-cover:
	go tool cover -func=coverage.out
