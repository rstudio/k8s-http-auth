all: build-api build-db

build-%:
	cd $* && docker build -t k8s-http-auth/$*:local .
