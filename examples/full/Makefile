SHELL := /bin/bash

all: build kind-ensure-cluster kind-load apply

build: build-api build-db

build-%:
	cd $* && docker build -t k8s-http-auth/$*:local .

clean:
	rm -vf port-forward.log

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

test: wait-for-it.sh
	./wait-for-it.sh 127.0.0.1:18080 -- curl -fsiSL http://127.0.0.1:18080 && \
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
