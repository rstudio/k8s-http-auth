.PHONY: all
all: test show-cover

.PHONY: test
test:
	go test -coverprofile=coverage.out -v ./...

.PHONY: integration-test
integration-test:
	$(MAKE) -C examples/full
	$(MAKE) -C examples/full start-port-forwarding
	$(MAKE) -C examples/full test

.PHONY: show-cover
show-cover:
	go tool cover -func=coverage.out
