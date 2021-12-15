TOP := $(shell git rev-parse --show-toplevel)
SHELL := /bin/bash

all: copy-self build kind-ensure-cluster kind-load apply wait-for-db wait-for-api

copy-self: api/k8s-http-auth/go.mod db/k8s-http-auth/go.mod

%/k8s-http-auth/go.mod:
	mkdir -p $(dir $@) && \
	cd $(TOP) && \
	rsync -av \
	  --exclude Makefile \
	  --exclude README.md \
	  --exclude examples/ \
  	  --exclude .git/ \
  	  --exclude .github/ \
  	  --exclude .gitignore \
  	  --exclude LICENSE \
  	  --exclude coverage.out \
	  ./ \
	  examples/full/$(dir $@)

build: build-api build-db

build-%:
	cd $* && docker build -t k8s-http-auth/$*:local .

clean:
	rm -rvf port-forward.log api/k8s-http-auth/ db/k8s-http-auth/

kind-load: kind-load-api kind-load-db

kind-load-%:
	kind load docker-image k8s-http-auth/$*:local

kind-ensure-cluster:
	kind get clusters | if ! grep -q ^kind; then \
		kind create cluster; \
	fi

apply:
	kubectl apply -f ./deployment.yaml

delete:
	kubectl delete -f ./deployment.yaml

start-port-forwarding:
	./port-forwarding start

stop-port-forwarding:
	./port-forwarding stop

wait-for-%:
	kubectl -n k8s-http-auth-system wait --for=condition=ready --timeout=5m pod -l app=$*

test: wait-for-it.sh
	./wait-for-it.sh 127.0.0.1:18080 -- curl --retry-connrefused --retry 5 -fsiSL http://127.0.0.1:18080 && \
		kubectl -n k8s-http-auth-system logs --tail 2 svc/api && \
		kubectl -n k8s-http-auth-system logs --tail 2 svc/db

dump-logs:
	kubectl -n k8s-http-auth-system get events
	kubectl -n k8s-http-auth-system logs svc/api && \
		kubectl -n k8s-http-auth-system logs svc/db && \
		cat port-forward.log

wait-for-it.sh:
	curl -fsSL -o $@ https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh
	chmod -v +x $@
