.PHONY: all
all: test show-cover

.PHONY: test
test:
	go test -coverprofile=coverage.out -v ./...

.PHONY: deps-up
deps-up:
	go get -t -u ./... && $(MAKE) -C examples/full deps-up

.PHONY: integration-test
integration-test:
	$(MAKE) -C examples/full

.PHONY: show-cover
show-cover:
	go tool cover -func=coverage.out
